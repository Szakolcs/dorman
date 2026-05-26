package chat

import (
	"sort"
	"strings"

	"dorm-man/internal/models"

	"github.com/google/uuid"
)

const maxMessageBodyLen = 10_000

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) GetTenantByUserID(userID uuid.UUID) (models.Tenant, error) {
	if userID == uuid.Nil {
		return models.Tenant{}, ErrUnauthorized
	}
	return s.store.GetTenantByUserID(userID)
}

func (s *Service) ListRooms(tenantID uuid.UUID) ([]models.Membership, error) {
	if tenantID == uuid.Nil {
		return nil, ErrValidation
	}
	return s.store.ListMemberships(tenantID)
}

func (s *Service) ListMessages(tenantID, roomID uuid.UUID) ([]models.Message, error) {
	if tenantID == uuid.Nil || roomID == uuid.Nil {
		return nil, ErrValidation
	}
	if _, err := s.store.GetMembership(tenantID, roomID); err != nil {
		return nil, err
	}
	return s.store.ListMessages(roomID)
}

func (s *Service) SearchChats(viewerTenantID uuid.UUID, query string) (SearchResult, error) {
	if viewerTenantID == uuid.Nil {
		return SearchResult{}, ErrValidation
	}
	query = strings.TrimSpace(query)
	memberships, err := s.store.ListMemberships(viewerTenantID)
	if err != nil {
		return SearchResult{}, err
	}
	titles, err := s.roomTitles(viewerTenantID, memberships)
	if err != nil {
		return SearchResult{}, err
	}
	result := SearchResult{Query: query, RoomTitles: titles}
	if query == "" {
		result.Rooms = memberships
		return result, nil
	}
	needle := strings.ToLower(query)
	for _, membership := range memberships {
		if membership.ChatRoom == nil {
			continue
		}
		displayTitle := strings.ToLower(titles[membership.ChatRoom.ID])
		if strings.Contains(displayTitle, needle) {
			result.Rooms = append(result.Rooms, membership)
		}
	}
	tenants, err := s.store.SearchTenants(viewerTenantID, query, 10)
	if err != nil {
		return SearchResult{}, err
	}
	result.Tenants = tenants
	return result, nil
}

func (s *Service) RoomDisplayTitle(viewerTenantID, roomID uuid.UUID) (string, error) {
	if roomID == uuid.Nil {
		return "", nil
	}
	memberships, err := s.store.ListMemberships(viewerTenantID)
	if err != nil {
		return "", err
	}
	titles, err := s.roomTitles(viewerTenantID, memberships)
	if err != nil {
		return "", err
	}
	return titles[roomID], nil
}

func (s *Service) roomTitles(viewerTenantID uuid.UUID, memberships []models.Membership) (map[uuid.UUID]string, error) {
	titles := make(map[uuid.UUID]string, len(memberships))
	var directRoomIDs []uuid.UUID
	for _, membership := range memberships {
		if membership.ChatRoom == nil {
			continue
		}
		if membership.ChatRoom.Kind == models.RoomKindDirect {
			directRoomIDs = append(directRoomIDs, membership.ChatRoom.ID)
			continue
		}
		titles[membership.ChatRoom.ID] = membership.ChatRoom.Title
	}
	peers, err := s.store.DirectPeerTenants(viewerTenantID, directRoomIDs)
	if err != nil {
		return nil, err
	}
	for roomID, peer := range peers {
		titles[roomID] = tenantName(peer)
	}
	for _, roomID := range directRoomIDs {
		if titles[roomID] == "" {
			titles[roomID] = "Direct chat"
		}
	}
	return titles, nil
}

func tenantName(tenant models.Tenant) string {
	if tenant.User != nil && tenant.User.Name != "" {
		return tenant.User.Name
	}
	return "Tenant"
}

func (s *Service) OpenDirectChat(viewerTenantID, otherTenantID uuid.UUID) (models.ChatRoom, error) {
	if viewerTenantID == uuid.Nil || otherTenantID == uuid.Nil {
		return models.ChatRoom{}, ErrValidation
	}
	if viewerTenantID == otherTenantID {
		return models.ChatRoom{}, ErrValidation
	}
	return s.store.OpenDirectRoom(viewerTenantID, otherTenantID)
}

func (s *Service) ListDirectChatPeers(viewerTenantID uuid.UUID) ([]models.Tenant, error) {
	if viewerTenantID == uuid.Nil {
		return nil, ErrValidation
	}
	memberships, err := s.store.ListMemberships(viewerTenantID)
	if err != nil {
		return nil, err
	}
	var directRoomIDs []uuid.UUID
	for _, membership := range memberships {
		if membership.ChatRoom != nil && membership.ChatRoom.Kind == models.RoomKindDirect {
			directRoomIDs = append(directRoomIDs, membership.ChatRoom.ID)
		}
	}
	peerMap, err := s.store.DirectPeerTenants(viewerTenantID, directRoomIDs)
	if err != nil {
		return nil, err
	}
	peers := make([]models.Tenant, 0, len(peerMap))
	for _, peer := range peerMap {
		peers = append(peers, peer)
	}
	sort.Slice(peers, func(i, j int) bool {
		return strings.ToLower(tenantName(peers[i])) < strings.ToLower(tenantName(peers[j]))
	})
	return peers, nil
}

func (s *Service) CreateGroup(creatorTenantID uuid.UUID, req CreateGroupRequest) (models.ChatRoom, error) {
	allowedPeers, err := s.ListDirectChatPeers(creatorTenantID)
	if err != nil {
		return models.ChatRoom{}, err
	}
	allowed := make(map[uuid.UUID]struct{}, len(allowedPeers))
	for _, peer := range allowedPeers {
		allowed[peer.ID] = struct{}{}
	}
	memberIDs := make([]uuid.UUID, 0, len(req.MemberIDs))
	for _, id := range req.MemberIDs {
		if id == uuid.Nil || id == creatorTenantID {
			continue
		}
		if _, ok := allowed[id]; !ok {
			return models.ChatRoom{}, ErrValidation
		}
		memberIDs = append(memberIDs, id)
	}
	return s.store.CreateGroupRoom(creatorTenantID, req.Title, req.Topic, memberIDs)
}

func (s *Service) SendMessage(tenantID, roomID uuid.UUID, body string) (models.Message, error) {
	if tenantID == uuid.Nil || roomID == uuid.Nil {
		return models.Message{}, ErrValidation
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return models.Message{}, ErrValidation
	}
	if len(body) > maxMessageBodyLen {
		return models.Message{}, ErrValidation
	}
	if _, err := s.store.GetMembership(tenantID, roomID); err != nil {
		return models.Message{}, err
	}
	message := models.Message{
		RoomID:         roomID,
		SenderTenantID: tenantID,
		Body:           body,
	}
	if err := s.store.CreateMessage(message); err != nil {
		return models.Message{}, err
	}
	messages, err := s.store.ListMessages(roomID)
	if err != nil {
		return message, nil
	}
	for _, m := range messages {
		if m.ID == message.ID {
			return m, nil
		}
	}
	return message, nil
}
