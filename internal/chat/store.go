package chat

import (
	"errors"

	"dorm-man/internal/administration"
	"dorm-man/internal/pagination"

	adm "dorm-man/internal/models/administration"
	cm "dorm-man/internal/models/chat"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store interface {
	Transaction(fn func(Store) error) error

	GetTenant(tenantID uuid.UUID) (adm.Tenant, error)
	GetTenantByUserID(userID uuid.UUID) (adm.Tenant, error)
	GetFlat(flatID uuid.UUID) (adm.Flat, error)
	HasActiveAssignmentInFlat(tenantID, flatID uuid.UUID) (bool, error)

	GetRoom(roomID uuid.UUID) (cm.ChatRoom, error)
	GetFlatRoom(flatID uuid.UUID) (cm.ChatRoom, error)
	GetDirectRoom(lowID, highID uuid.UUID) (cm.ChatRoom, error)
	CreateRoom(room cm.ChatRoom) (cm.ChatRoom, error)
	UpdateRoom(room cm.ChatRoom) (cm.ChatRoom, error)

	GetActiveMember(roomID, tenantID uuid.UUID) (cm.ChatRoomMember, error)
	ListActiveMembers(roomID uuid.UUID) ([]cm.ChatRoomMember, error)
	ListActiveMembershipsForTenant(tenantID uuid.UUID) ([]cm.ChatRoomMember, error)
	CreateMember(member cm.ChatRoomMember) (cm.ChatRoomMember, error)
	UpdateMember(member cm.ChatRoomMember) (cm.ChatRoomMember, error)

	GetMessage(messageID uuid.UUID) (cm.ChatMessage, error)
	GetMessageByClientKey(roomID, authorTenantID uuid.UUID, clientMessageID uuid.UUID) (cm.ChatMessage, error)
	GetLatestMessageInRoom(roomID uuid.UUID) (cm.ChatMessage, error)
	ListMessages(roomID uuid.UUID, filter MessageListFilter) ([]cm.ChatMessage, int64, error)
	CreateMessage(msg cm.ChatMessage) (cm.ChatMessage, error)
	CountUnreadMessages(roomID, readerTenantID uuid.UUID, lastReadMessageID *uuid.UUID) (int64, error)

	GetProfile(tenantID uuid.UUID) (cm.ChatTenantProfile, error)
	GetProfiles(tenantIDs []uuid.UUID) (map[uuid.UUID]cm.ChatTenantProfile, error)
	UpsertProfile(profile cm.ChatTenantProfile) (cm.ChatTenantProfile, error)

	CreateMembershipSyncLog(entry cm.ChatMembershipSyncLog) error
	CreateAudit(event adm.AuditEvent) error
}

type GormStore struct {
	db      *gorm.DB
	housing administration.Store
}

func NewStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db, housing: administration.NewStore(db)}
}

func (s *GormStore) Transaction(fn func(Store) error) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return fn(&GormStore{db: tx, housing: s.housing})
	})
}

func (s *GormStore) GetTenant(tenantID uuid.UUID) (adm.Tenant, error) {
	tenant, err := s.housing.GetTenant(tenantID)
	if errors.Is(err, administration.ErrNotFound) {
		return adm.Tenant{}, ErrTenantNotFound
	}
	return tenant, err
}

func (s *GormStore) GetTenantByUserID(userID uuid.UUID) (adm.Tenant, error) {
	var tenant adm.Tenant
	err := s.db.Where("user_id = ?", userID).First(&tenant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return adm.Tenant{}, ErrTenantNotFound
	}
	return tenant, err
}

func (s *GormStore) GetFlat(flatID uuid.UUID) (adm.Flat, error) {
	var flat adm.Flat
	err := s.db.First(&flat, "id = ?", flatID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return adm.Flat{}, ErrNotFound
	}
	return flat, err
}

func (s *GormStore) HasActiveAssignmentInFlat(tenantID, flatID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.Model(&adm.RoomAssignment{}).
		Joins("JOIN rooms ON rooms.id = room_assignments.room_id").
		Where("room_assignments.tenant_id = ? AND room_assignments.ended_at IS NULL AND rooms.flat_id = ?", tenantID, flatID).
		Count(&count).Error
	return count > 0, err
}

func (s *GormStore) GetRoom(roomID uuid.UUID) (cm.ChatRoom, error) {
	var room cm.ChatRoom
	err := s.db.Preload("Flat").Preload("TenantLow").Preload("TenantHigh").
		First(&room, "id = ?", roomID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cm.ChatRoom{}, ErrNotFound
	}
	return room, err
}

func (s *GormStore) GetFlatRoom(flatID uuid.UUID) (cm.ChatRoom, error) {
	var room cm.ChatRoom
	err := s.db.Where("kind = ? AND flat_id = ?", cm.ChatRoomKindFlat, flatID).First(&room).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cm.ChatRoom{}, ErrNotFound
	}
	return room, err
}

func (s *GormStore) GetDirectRoom(lowID, highID uuid.UUID) (cm.ChatRoom, error) {
	var room cm.ChatRoom
	err := s.db.Where(
		"kind = ? AND tenant_low_id = ? AND tenant_high_id = ?",
		cm.ChatRoomKindDirect, lowID, highID,
	).First(&room).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cm.ChatRoom{}, ErrNotFound
	}
	return room, err
}

func (s *GormStore) CreateRoom(room cm.ChatRoom) (cm.ChatRoom, error) {
	if err := s.db.Create(&room).Error; err != nil {
		return cm.ChatRoom{}, err
	}
	return room, nil
}

func (s *GormStore) UpdateRoom(room cm.ChatRoom) (cm.ChatRoom, error) {
	if err := s.db.Save(&room).Error; err != nil {
		return cm.ChatRoom{}, err
	}
	return room, nil
}

func (s *GormStore) GetActiveMember(roomID, tenantID uuid.UUID) (cm.ChatRoomMember, error) {
	var member cm.ChatRoomMember
	err := s.db.Where("room_id = ? AND tenant_id = ? AND left_at IS NULL", roomID, tenantID).
		Preload("Tenant").
		First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cm.ChatRoomMember{}, ErrNotAMember
	}
	return member, err
}

func (s *GormStore) ListActiveMembers(roomID uuid.UUID) ([]cm.ChatRoomMember, error) {
	var members []cm.ChatRoomMember
	err := s.db.Where("room_id = ? AND left_at IS NULL", roomID).
		Preload("Tenant").
		Order("joined_at ASC").
		Find(&members).Error
	return members, err
}

func (s *GormStore) ListActiveMembershipsForTenant(tenantID uuid.UUID) ([]cm.ChatRoomMember, error) {
	var members []cm.ChatRoomMember
	err := s.db.Where("tenant_id = ? AND left_at IS NULL", tenantID).
		Preload("Room").
		Preload("Room.Flat").
		Preload("Room.TenantLow").
		Preload("Room.TenantHigh").
		Find(&members).Error
	return members, err
}

func (s *GormStore) CreateMember(member cm.ChatRoomMember) (cm.ChatRoomMember, error) {
	if err := s.db.Create(&member).Error; err != nil {
		return cm.ChatRoomMember{}, err
	}
	return member, nil
}

func (s *GormStore) UpdateMember(member cm.ChatRoomMember) (cm.ChatRoomMember, error) {
	if err := s.db.Save(&member).Error; err != nil {
		return cm.ChatRoomMember{}, err
	}
	return member, nil
}

func (s *GormStore) GetMessage(messageID uuid.UUID) (cm.ChatMessage, error) {
	var msg cm.ChatMessage
	err := s.db.Preload("AuthorTenant").First(&msg, "id = ?", messageID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cm.ChatMessage{}, ErrNotFound
	}
	return msg, err
}

func (s *GormStore) GetMessageByClientKey(roomID, authorTenantID uuid.UUID, clientMessageID uuid.UUID) (cm.ChatMessage, error) {
	var msg cm.ChatMessage
	err := s.db.Where(
		"room_id = ? AND author_tenant_id = ? AND client_message_id = ?",
		roomID, authorTenantID, clientMessageID,
	).First(&msg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cm.ChatMessage{}, ErrNotFound
	}
	return msg, err
}

func (s *GormStore) GetLatestMessageInRoom(roomID uuid.UUID) (cm.ChatMessage, error) {
	var msg cm.ChatMessage
	err := s.db.Where("room_id = ?", roomID).
		Order("created_at DESC, id DESC").
		First(&msg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cm.ChatMessage{}, ErrNotFound
	}
	return msg, err
}

func (s *GormStore) ListMessages(roomID uuid.UUID, filter MessageListFilter) ([]cm.ChatMessage, int64, error) {
	limit := filter.Params.PageSize
	if limit <= 0 || limit > 50 {
		limit = 50
	}

	q := s.db.Model(&cm.ChatMessage{}).Where("room_id = ?", roomID)
	if filter.BeforeMessageID != nil {
		anchor, err := s.GetMessage(*filter.BeforeMessageID)
		if err != nil {
			return nil, 0, err
		}
		q = q.Where(
			"(created_at < ? OR (created_at = ? AND id < ?))",
			anchor.CreatedAt, anchor.CreatedAt, anchor.ID,
		)
		var list []cm.ChatMessage
		err = q.Order("created_at DESC, id DESC").
			Limit(limit).
			Preload("AuthorTenant").
			Find(&list).Error
		if err != nil {
			return nil, 0, err
		}
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
		return list, int64(len(list)), nil
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []cm.ChatMessage
	err := q.Order("created_at DESC, id DESC").
		Scopes(pagination.Scope(filter.Params)).
		Preload("AuthorTenant").
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, total, nil
}

func (s *GormStore) CreateMessage(msg cm.ChatMessage) (cm.ChatMessage, error) {
	if err := s.db.Create(&msg).Error; err != nil {
		return cm.ChatMessage{}, err
	}
	return msg, nil
}

func (s *GormStore) CountUnreadMessages(roomID, readerTenantID uuid.UUID, lastReadMessageID *uuid.UUID) (int64, error) {
	q := s.db.Model(&cm.ChatMessage{}).
		Where("room_id = ? AND author_tenant_id <> ?", roomID, readerTenantID)

	if lastReadMessageID != nil {
		anchor, err := s.GetMessage(*lastReadMessageID)
		if err != nil {
			if !errors.Is(err, ErrNotFound) {
				return 0, err
			}
		} else {
			q = q.Where(
				"(created_at > ? OR (created_at = ? AND id > ?))",
				anchor.CreatedAt, anchor.CreatedAt, anchor.ID,
			)
		}
	}

	var count int64
	err := q.Count(&count).Error
	return count, err
}

func (s *GormStore) GetProfile(tenantID uuid.UUID) (cm.ChatTenantProfile, error) {
	var profile cm.ChatTenantProfile
	err := s.db.First(&profile, "tenant_id = ?", tenantID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cm.ChatTenantProfile{}, ErrNotFound
	}
	return profile, err
}

func (s *GormStore) GetProfiles(tenantIDs []uuid.UUID) (map[uuid.UUID]cm.ChatTenantProfile, error) {
	out := make(map[uuid.UUID]cm.ChatTenantProfile)
	if len(tenantIDs) == 0 {
		return out, nil
	}
	var profiles []cm.ChatTenantProfile
	if err := s.db.Where("tenant_id IN ?", tenantIDs).Find(&profiles).Error; err != nil {
		return nil, err
	}
	for _, p := range profiles {
		out[p.TenantID] = p
	}
	return out, nil
}

func (s *GormStore) UpsertProfile(profile cm.ChatTenantProfile) (cm.ChatTenantProfile, error) {
	var existing cm.ChatTenantProfile
	err := s.db.Where("tenant_id = ?", profile.TenantID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := s.db.Create(&profile).Error; err != nil {
			return cm.ChatTenantProfile{}, err
		}
		return profile, nil
	}
	if err != nil {
		return cm.ChatTenantProfile{}, err
	}
	existing.Nickname = profile.Nickname
	existing.Bio = profile.Bio
	existing.AvatarStorageKey = profile.AvatarStorageKey
	existing.ProfileUpdatedAt = profile.ProfileUpdatedAt
	if err := s.db.Save(&existing).Error; err != nil {
		return cm.ChatTenantProfile{}, err
	}
	return existing, nil
}

func (s *GormStore) CreateMembershipSyncLog(entry cm.ChatMembershipSyncLog) error {
	return s.db.Create(&entry).Error
}

func (s *GormStore) CreateAudit(event adm.AuditEvent) error {
	return s.housing.CreateAudit(event)
}
