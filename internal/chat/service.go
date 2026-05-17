package chat

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"dorm-man/internal/chatviews"

	adm "dorm-man/internal/models/administration"
	cm "dorm-man/internal/models/chat"

	"github.com/google/uuid"
)

const (
	maxMessageBodyLen = 10_000
	maxBioLen         = 2_000
	maxGroupTitleLen  = 200
	maxPreviewLen     = 500
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ResolveTenantPrincipal(actorUserID uuid.UUID) (TenantPrincipal, error) {
	tenant, err := s.store.GetTenantByUserID(actorUserID)
	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return TenantPrincipal{}, ErrUnauthorized
		}
		return TenantPrincipal{}, err
	}
	if !tenant.IsActive {
		return TenantPrincipal{}, ErrUnauthorized
	}
	uid := actorUserID
	return TenantPrincipal{TenantID: tenant.ID, UserID: &uid}, nil
}

func (s *Service) ListConversations(p TenantPrincipal) ([]chatviews.ConversationSummary, error) {
	memberships, err := s.store.ListActiveMembershipsForTenant(p.TenantID)
	if err != nil {
		return nil, err
	}

	profiles, err := s.loadPeerProfiles(memberships, p.TenantID)
	if err != nil {
		return nil, err
	}

	out := make([]chatviews.ConversationSummary, 0, len(memberships))
	for _, m := range memberships {
		unread, err := s.store.CountUnreadMessages(m.RoomID, p.TenantID, m.LastReadMessageID)
		if err != nil {
			return nil, err
		}
		title, avatar, otherID := conversationDisplay(m.Room, p.TenantID, profiles)
		out = append(out, chatviews.ConversationSummary{
			Room:                    m.Room,
			UnreadCount:             unread,
			DisplayTitle:            title,
			DisplayAvatarStorageKey: avatar,
			OtherTenantID:           otherID,
		})
	}

	sortConversations(out)
	return out, nil
}

func (s *Service) GetRoom(p TenantPrincipal, roomID uuid.UUID) (chatviews.RoomDetail, error) {
	room, err := s.store.GetRoom(roomID)
	if err != nil {
		return chatviews.RoomDetail{}, err
	}
	if _, err := s.store.GetActiveMember(roomID, p.TenantID); err != nil {
		if errors.Is(err, ErrNotAMember) {
			return chatviews.RoomDetail{}, ErrNotFound
		}
		return chatviews.RoomDetail{}, err
	}
	members, err := s.store.ListActiveMembers(roomID)
	if err != nil {
		return chatviews.RoomDetail{}, err
	}

	tenantIDs := make([]uuid.UUID, 0, len(members))
	for _, m := range members {
		tenantIDs = append(tenantIDs, m.TenantID)
	}
	profiles, err := s.store.GetProfiles(tenantIDs)
	if err != nil {
		return chatviews.RoomDetail{}, err
	}

	views := make([]chatviews.MemberView, 0, len(members))
	var selfView chatviews.MemberView
	for _, m := range members {
		view := memberView(m, profiles[m.TenantID])
		views = append(views, view)
		if m.TenantID == p.TenantID {
			selfView = view
		}
	}

	return chatviews.RoomDetail{Room: room, Members: views, Self: selfView}, nil
}

func (s *Service) ListMessages(p TenantPrincipal, roomID uuid.UUID, filter MessageListFilter) ([]cm.ChatMessage, error) {
	if _, err := s.store.GetActiveMember(roomID, p.TenantID); err != nil {
		if errors.Is(err, ErrNotAMember) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.store.ListMessages(roomID, filter.BeforeMessageID, filter.Limit)
}

func (s *Service) SendMessage(p TenantPrincipal, roomID uuid.UUID, in SendMessageInput) (cm.ChatMessage, error) {
	if _, err := s.store.GetActiveMember(roomID, p.TenantID); err != nil {
		return cm.ChatMessage{}, err
	}
	body := strings.TrimSpace(in.Body)
	if body == "" || utf8.RuneCountInString(body) > maxMessageBodyLen {
		return cm.ChatMessage{}, ErrValidation
	}
	if in.ClientMessageID == nil {
		return cm.ChatMessage{}, ErrValidation
	}

	if existing, err := s.store.GetMessageByClientKey(roomID, p.TenantID, *in.ClientMessageID); err == nil {
		return existing, nil
	} else if !errors.Is(err, ErrNotFound) {
		return cm.ChatMessage{}, err
	}

	now := time.Now().UTC()
	preview := body
	if len(preview) > maxPreviewLen {
		preview = preview[:maxPreviewLen]
	}

	var created cm.ChatMessage
	err := s.store.Transaction(func(tx Store) error {
		msg, err := tx.CreateMessage(cm.ChatMessage{
			RoomID:          roomID,
			AuthorTenantID:  p.TenantID,
			Body:            body,
			ClientMessageID: in.ClientMessageID,
		})
		if err != nil {
			return err
		}
		created = msg

		room, err := tx.GetRoom(roomID)
		if err != nil {
			return err
		}
		room.LastMessageAt = &now
		room.LastMessagePreview = preview
		_, err = tx.UpdateRoom(room)
		return err
	})
	if err != nil {
		return cm.ChatMessage{}, err
	}
	return created, nil
}

func (s *Service) MarkRoomRead(p TenantPrincipal, roomID uuid.UUID, in MarkReadInput) (cm.ChatRoomMember, error) {
	member, err := s.store.GetActiveMember(roomID, p.TenantID)
	if err != nil {
		return cm.ChatRoomMember{}, err
	}

	var targetID uuid.UUID
	if in.MessageID != nil {
		msg, err := s.store.GetMessage(*in.MessageID)
		if err != nil {
			return cm.ChatRoomMember{}, err
		}
		if msg.RoomID != roomID {
			return cm.ChatRoomMember{}, ErrValidation
		}
		targetID = msg.ID
	} else {
		latest, err := s.store.GetLatestMessageInRoom(roomID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return member, nil
			}
			return cm.ChatRoomMember{}, err
		}
		targetID = latest.ID
	}

	if member.LastReadMessageID != nil {
		current, err := s.store.GetMessage(*member.LastReadMessageID)
		if err == nil {
			anchor, err := s.store.GetMessage(targetID)
			if err != nil {
				return cm.ChatRoomMember{}, err
			}
			if !messageAfter(anchor, current) {
				return member, nil
			}
		}
	}

	member.LastReadMessageID = &targetID
	return s.store.UpdateMember(member)
}

func (s *Service) OpenDirect(p TenantPrincipal, in OpenDirectInput) (cm.ChatRoom, error) {
	if in.OtherTenantID == p.TenantID {
		return cm.ChatRoom{}, ErrValidation
	}
	if _, err := s.store.GetTenant(in.OtherTenantID); err != nil {
		return cm.ChatRoom{}, err
	}

	low, high := canonicalTenantPair(p.TenantID, in.OtherTenantID)
	if room, err := s.store.GetDirectRoom(low, high); err == nil {
		return room, nil
	} else if !errors.Is(err, ErrNotFound) {
		return cm.ChatRoom{}, err
	}

	now := time.Now().UTC()
	var room cm.ChatRoom
	err := s.store.Transaction(func(tx Store) error {
		var err error
		room, err = tx.CreateRoom(cm.ChatRoom{
			Kind:         cm.ChatRoomKindDirect,
			TenantLowID:  &low,
			TenantHighID: &high,
		})
		if err != nil {
			return err
		}
		for _, tid := range []uuid.UUID{low, high} {
			_, err = tx.CreateMember(cm.ChatRoomMember{
				RoomID:   room.ID,
				TenantID: tid,
				Role:     cm.ChatRoomMemberRoleMember,
				JoinedAt: now,
				Source:   cm.ChatRoomMemberSourceCreated,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return room, err
}

func (s *Service) CreateGroup(p TenantPrincipal, in CreateGroupInput) (cm.ChatRoom, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || utf8.RuneCountInString(title) > maxGroupTitleLen {
		return cm.ChatRoom{}, ErrValidation
	}

	now := time.Now().UTC()
	var room cm.ChatRoom
	err := s.store.Transaction(func(tx Store) error {
		var err error
		room, err = tx.CreateRoom(cm.ChatRoom{
			Kind:             cm.ChatRoomKindGroup,
			Title:            title,
			AvatarStorageKey: strings.TrimSpace(in.AvatarStorageKey),
		})
		if err != nil {
			return err
		}
		_, err = tx.CreateMember(cm.ChatRoomMember{
			RoomID:   room.ID,
			TenantID: p.TenantID,
			Role:     cm.ChatRoomMemberRoleOwner,
			JoinedAt: now,
			Source:   cm.ChatRoomMemberSourceCreated,
		})
		if err != nil {
			return err
		}
		for _, tid := range in.MemberTenantIDs {
			if tid == p.TenantID {
				continue
			}
			if _, err := tx.GetTenant(tid); err != nil {
				return err
			}
			_, err = tx.CreateMember(cm.ChatRoomMember{
				RoomID:   room.ID,
				TenantID: tid,
				Role:     cm.ChatRoomMemberRoleMember,
				JoinedAt: now,
				Source:   cm.ChatRoomMemberSourceInvited,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return room, err
}

func (s *Service) AddGroupMember(p TenantPrincipal, roomID uuid.UUID, tenantID uuid.UUID) (cm.ChatRoomMember, error) {
	room, err := s.requireGroupOwner(p, roomID)
	if err != nil {
		return cm.ChatRoomMember{}, err
	}
	_ = room
	if _, err := s.store.GetTenant(tenantID); err != nil {
		return cm.ChatRoomMember{}, err
	}
	if _, err := s.store.GetActiveMember(roomID, tenantID); err == nil {
		return cm.ChatRoomMember{}, ErrValidation
	} else if !errors.Is(err, ErrNotAMember) {
		return cm.ChatRoomMember{}, err
	}

	member := cm.ChatRoomMember{
		RoomID:   roomID,
		TenantID: tenantID,
		Role:     cm.ChatRoomMemberRoleMember,
		JoinedAt: time.Now().UTC(),
		Source:   cm.ChatRoomMemberSourceInvited,
	}
	return s.store.CreateMember(member)
}

func (s *Service) RemoveGroupMember(p TenantPrincipal, roomID, tenantID uuid.UUID) (cm.ChatRoomMember, error) {
	if _, err := s.requireGroupOwner(p, roomID); err != nil {
		return cm.ChatRoomMember{}, err
	}
	member, err := s.store.GetActiveMember(roomID, tenantID)
	if err != nil {
		return cm.ChatRoomMember{}, err
	}
	now := time.Now().UTC()
	member.LeftAt = &now
	updated, err := s.store.UpdateMember(member)
	if err != nil {
		return cm.ChatRoomMember{}, err
	}
	if p.UserID != nil {
		_ = s.store.CreateAudit(newAudit(*p.UserID, "chat.group.member.remove", "chat_room_member", member.ID, adm.AuditOutcomeSuccess))
	}
	return updated, nil
}

func (s *Service) LeaveGroup(p TenantPrincipal, roomID uuid.UUID) (cm.ChatRoomMember, error) {
	room, err := s.store.GetRoom(roomID)
	if err != nil {
		return cm.ChatRoomMember{}, err
	}
	if room.Kind != cm.ChatRoomKindGroup {
		return cm.ChatRoomMember{}, ErrRoomKindMismatch
	}
	member, err := s.store.GetActiveMember(roomID, p.TenantID)
	if err != nil {
		return cm.ChatRoomMember{}, err
	}
	now := time.Now().UTC()
	member.LeftAt = &now
	return s.store.UpdateMember(member)
}

func (s *Service) GetProfile(p TenantPrincipal) (chatviews.TenantProfileView, error) {
	tenant, err := s.store.GetTenant(p.TenantID)
	if err != nil {
		return chatviews.TenantProfileView{}, err
	}
	profile, err := s.store.GetProfile(p.TenantID)
	if errors.Is(err, ErrNotFound) {
		return chatviews.TenantProfileView{
			Tenant:      tenant,
			DisplayName: displayName(tenant, cm.ChatTenantProfile{}),
		}, nil
	}
	if err != nil {
		return chatviews.TenantProfileView{}, err
	}
	return chatviews.TenantProfileView{
		Profile:     &profile,
		Tenant:      tenant,
		DisplayName: displayName(tenant, profile),
	}, nil
}

func (s *Service) UpdateProfile(p TenantPrincipal, in UpdateProfileInput) (chatviews.TenantProfileView, error) {
	tenant, err := s.store.GetTenant(p.TenantID)
	if err != nil {
		return chatviews.TenantProfileView{}, err
	}

	existing, profileErr := s.store.GetProfile(p.TenantID)
	if profileErr != nil && !errors.Is(profileErr, ErrNotFound) {
		return chatviews.TenantProfileView{}, profileErr
	}
	if errors.Is(profileErr, ErrNotFound) {
		existing = cm.ChatTenantProfile{TenantID: p.TenantID}
	}

	if in.Nickname != nil {
		nick := strings.TrimSpace(*in.Nickname)
		if nick == "" {
			existing.Nickname = nil
		} else {
			existing.Nickname = &nick
		}
	}
	if in.Bio != nil {
		bio := strings.TrimSpace(*in.Bio)
		if utf8.RuneCountInString(bio) > maxBioLen {
			return chatviews.TenantProfileView{}, ErrValidation
		}
		existing.Bio = bio
	}
	if in.AvatarStorageKey != nil {
		existing.AvatarStorageKey = strings.TrimSpace(*in.AvatarStorageKey)
	}
	existing.ProfileUpdatedAt = time.Now().UTC()

	saved, err := s.store.UpsertProfile(existing)
	if err != nil {
		return chatviews.TenantProfileView{}, err
	}
	if p.UserID != nil {
		_ = s.store.CreateAudit(newAudit(*p.UserID, "chat.profile.update", "chat_tenant_profile", saved.TenantID, adm.AuditOutcomeSuccess))
	}
	return chatviews.TenantProfileView{
		Profile:     &saved,
		Tenant:      tenant,
		DisplayName: displayName(tenant, saved),
	}, nil
}

// SyncFlatMembershipForTenant applies assignment-driven flat-room membership (FR-CM-002).
func (s *Service) SyncFlatMembershipForTenant(tenantID, flatID uuid.UUID, eventType cm.ChatMembershipSyncEventType) error {
	flat, err := s.store.GetFlat(flatID)
	if err != nil {
		return err
	}

	shouldMember, err := s.store.HasActiveAssignmentInFlat(tenantID, flatID)
	if err != nil {
		return err
	}

	room, err := s.ensureFlatRoom(flat)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if shouldMember {
		if _, err := s.store.GetActiveMember(room.ID, tenantID); err == nil {
			return s.logSync(tenantID, flatID, eventType, map[string]any{"action": "noop_already_member"})
		} else if !errors.Is(err, ErrNotAMember) {
			return err
		}
		_, err = s.store.CreateMember(cm.ChatRoomMember{
			RoomID:   room.ID,
			TenantID: tenantID,
			Role:     cm.ChatRoomMemberRoleMember,
			JoinedAt: now,
			Source:   cm.ChatRoomMemberSourceDerived,
		})
		if err != nil {
			return err
		}
		return s.logSync(tenantID, flatID, eventType, map[string]any{"action": "added"})
	}

	member, err := s.store.GetActiveMember(room.ID, tenantID)
	if errors.Is(err, ErrNotAMember) {
		return s.logSync(tenantID, flatID, eventType, map[string]any{"action": "noop_not_member"})
	}
	if err != nil {
		return err
	}
	member.LeftAt = &now
	if _, err := s.store.UpdateMember(member); err != nil {
		return err
	}
	return s.logSync(tenantID, flatID, eventType, map[string]any{"action": "removed"})
}

func (s *Service) ensureFlatRoom(flat adm.Flat) (cm.ChatRoom, error) {
	if room, err := s.store.GetFlatRoom(flat.ID); err == nil {
		return room, nil
	} else if !errors.Is(err, ErrNotFound) {
		return cm.ChatRoom{}, err
	}
	title := strings.TrimSpace(flat.Name)
	if title == "" {
		title = "Flat chat"
	}
	fid := flat.ID
	return s.store.CreateRoom(cm.ChatRoom{
		Kind:   cm.ChatRoomKindFlat,
		Title:  title,
		FlatID: &fid,
	})
}

func (s *Service) requireGroupOwner(p TenantPrincipal, roomID uuid.UUID) (cm.ChatRoom, error) {
	room, err := s.store.GetRoom(roomID)
	if err != nil {
		return cm.ChatRoom{}, err
	}
	if room.Kind != cm.ChatRoomKindGroup {
		return cm.ChatRoom{}, ErrRoomKindMismatch
	}
	member, err := s.store.GetActiveMember(roomID, p.TenantID)
	if err != nil {
		return cm.ChatRoom{}, err
	}
	if member.Role != cm.ChatRoomMemberRoleOwner {
		return cm.ChatRoom{}, ErrNotGroupOwner
	}
	return room, nil
}

func (s *Service) loadPeerProfiles(memberships []cm.ChatRoomMember, selfID uuid.UUID) (map[uuid.UUID]cm.ChatTenantProfile, error) {
	ids := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for _, m := range memberships {
		room := m.Room
		switch room.Kind {
		case cm.ChatRoomKindDirect:
			var other uuid.UUID
			if room.TenantLowID != nil && *room.TenantLowID != selfID {
				other = *room.TenantLowID
			} else if room.TenantHighID != nil {
				other = *room.TenantHighID
			}
			if _, ok := seen[other]; other != uuid.Nil && !ok {
				seen[other] = struct{}{}
				ids = append(ids, other)
			}
		}
	}
	return s.store.GetProfiles(ids)
}

func (s *Service) logSync(tenantID, flatID uuid.UUID, eventType cm.ChatMembershipSyncEventType, payload map[string]any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.store.CreateMembershipSyncLog(cm.ChatMembershipSyncLog{
		TenantID:  tenantID,
		FlatID:    flatID,
		EventType: eventType,
		AppliedAt: time.Now().UTC(),
		Payload:   raw,
	})
}

func canonicalTenantPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if bytes.Compare(a[:], b[:]) < 0 {
		return a, b
	}
	return b, a
}

func messageAfter(later, earlier cm.ChatMessage) bool {
	if later.CreatedAt.After(earlier.CreatedAt) {
		return true
	}
	return later.CreatedAt.Equal(earlier.CreatedAt) && bytes.Compare(later.ID[:], earlier.ID[:]) > 0
}

func displayName(tenant adm.Tenant, profile cm.ChatTenantProfile) string {
	if profile.Nickname != nil {
		n := strings.TrimSpace(*profile.Nickname)
		if n != "" {
			return n
		}
	}
	return strings.TrimSpace(tenant.Name)
}

func memberView(m cm.ChatRoomMember, profile cm.ChatTenantProfile) chatviews.MemberView {
	return chatviews.MemberView{
		Member:      m,
		Tenant:      m.Tenant,
		DisplayName: displayName(m.Tenant, profile),
	}
}

func conversationDisplay(room cm.ChatRoom, selfID uuid.UUID, profiles map[uuid.UUID]cm.ChatTenantProfile) (title, avatar string, otherID *uuid.UUID) {
	switch room.Kind {
	case cm.ChatRoomKindFlat:
		if room.Flat != nil && room.Flat.Name != "" {
			return room.Flat.Name, room.AvatarStorageKey, nil
		}
		return room.Title, room.AvatarStorageKey, nil
	case cm.ChatRoomKindGroup:
		return room.Title, room.AvatarStorageKey, nil
	case cm.ChatRoomKindDirect:
		var other adm.Tenant
		switch {
		case room.TenantLowID != nil && *room.TenantLowID != selfID && room.TenantLow != nil:
			other = *room.TenantLow
		case room.TenantHighID != nil && *room.TenantHighID != selfID && room.TenantHigh != nil:
			other = *room.TenantHigh
		case room.TenantLowID != nil && *room.TenantLowID != selfID:
			oid := *room.TenantLowID
			otherID = &oid
			return "Direct chat", "", otherID
		case room.TenantHighID != nil:
			oid := *room.TenantHighID
			otherID = &oid
			return "Direct chat", "", otherID
		}
		oid := other.ID
		otherID = &oid
		profile := profiles[other.ID]
		return displayName(other, profile), profile.AvatarStorageKey, otherID
	default:
		return room.Title, room.AvatarStorageKey, nil
	}
}

func sortConversations(list []chatviews.ConversationSummary) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if conversationLess(list[j], list[i]) {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

func conversationLess(a, b chatviews.ConversationSummary) bool {
	at := a.Room.LastMessageAt
	bt := b.Room.LastMessageAt
	if at == nil && bt == nil {
		return a.Room.CreatedAt.After(b.Room.CreatedAt)
	}
	if at == nil {
		return false
	}
	if bt == nil {
		return true
	}
	return at.After(*bt)
}

func newAudit(actorID uuid.UUID, action, targetType string, targetID uuid.UUID, outcome adm.AuditOutcome) adm.AuditEvent {
	return adm.AuditEvent{
		ActorUserID: &actorID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		Outcome:     outcome,
		OccurredAt:  time.Now().UTC(),
	}
}
