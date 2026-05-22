package chat

import (
	"strings"
	"time"

	"dorm-man/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store interface {
	GetTenantByUserID(userID uuid.UUID) (models.Tenant, error)
	ListMemberships(tenantID uuid.UUID) ([]models.Membership, error)
	GetMembership(tenantID, roomID uuid.UUID) (models.Membership, error)
	ListMessages(roomID uuid.UUID) ([]models.Message, error)
	CreateMessage(message models.Message) error
	SearchTenants(excludeTenantID uuid.UUID, query string, limit int) ([]models.Tenant, error)
	OpenDirectRoom(tenantA, tenantB uuid.UUID) (models.ChatRoom, error)
	CreateGroupRoom(creatorTenantID uuid.UUID, title, topic string) (models.ChatRoom, error)
}

type GormStore struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (s *GormStore) GetTenantByUserID(userID uuid.UUID) (models.Tenant, error) {
	var tenant models.Tenant
	err := s.db.
		Preload("User").
		Where("user_id = ?", userID).
		First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Tenant{}, ErrTenantNotFound
		}
		return models.Tenant{}, err
	}
	return tenant, nil
}

func (s *GormStore) ListMemberships(tenantID uuid.UUID) ([]models.Membership, error) {
	var memberships []models.Membership
	err := s.db.
		Preload("ChatRoom").
		Where("tenant_id = ?", tenantID).
		Order("joined_at DESC").
		Find(&memberships).Error
	return memberships, err
}

func (s *GormStore) GetMembership(tenantID, roomID uuid.UUID) (models.Membership, error) {
	var membership models.Membership
	err := s.db.
		Where("tenant_id = ? AND room_id = ?", tenantID, roomID).
		First(&membership).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Membership{}, ErrNotAMember
		}
		return models.Membership{}, err
	}
	return membership, nil
}

func (s *GormStore) ListMessages(roomID uuid.UUID) ([]models.Message, error) {
	var messages []models.Message
	err := s.db.
		Preload("Sender").
		Preload("Sender.User").
		Where("room_id = ?", roomID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func (s *GormStore) CreateMessage(message models.Message) error {
	return s.db.Create(&message).Error
}

func (s *GormStore) SearchTenants(excludeTenantID uuid.UUID, query string, limit int) ([]models.Tenant, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}
	pattern := "%" + strings.ToLower(query) + "%"
	var tenants []models.Tenant
	err := s.db.
		Preload("User").
		Joins("LEFT JOIN users ON users.id = tenants.user_id").
		Where("tenants.id != ? AND tenants.is_active = ?", excludeTenantID, true).
		Where(
			"LOWER(tenants.student_code) LIKE ? OR LOWER(users.name) LIKE ? OR LOWER(users.nickname) LIKE ?",
			pattern, pattern, pattern,
		).
		Order("tenants.student_code ASC").
		Limit(limit).
		Find(&tenants).Error
	return tenants, err
}

func directPairKey(a, b uuid.UUID) string {
	sa, sb := a.String(), b.String()
	if sa < sb {
		return sa + ":" + sb
	}
	return sb + ":" + sa
}

func (s *GormStore) OpenDirectRoom(tenantA, tenantB uuid.UUID) (models.ChatRoom, error) {
	if tenantA == uuid.Nil || tenantB == uuid.Nil || tenantA == tenantB {
		return models.ChatRoom{}, ErrValidation
	}
	key := directPairKey(tenantA, tenantB)
	var existing models.ChatRoom
	err := s.db.Where("direct_pair_key = ?", key).First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return models.ChatRoom{}, err
	}

	var room models.ChatRoom
	err = s.db.Transaction(func(tx *gorm.DB) error {
		pairKey := key
		room = models.ChatRoom{
			Kind:          models.RoomKindDirect,
			Title:         "Direct chat",
			DirectPairKey: &pairKey,
		}
		if err := tx.Create(&room).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, tenantID := range []uuid.UUID{tenantA, tenantB} {
			if err := tx.Create(&models.Membership{
				RoomID:   room.ID,
				TenantID: tenantID,
				Role:     models.MembershipRoleMember,
				JoinedAt: now,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return room, err
}

func (s *GormStore) CreateGroupRoom(creatorTenantID uuid.UUID, title, topic string) (models.ChatRoom, error) {
	if creatorTenantID == uuid.Nil {
		return models.ChatRoom{}, ErrValidation
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return models.ChatRoom{}, ErrValidation
	}

	var room models.ChatRoom
	err := s.db.Transaction(func(tx *gorm.DB) error {
		room = models.ChatRoom{
			Kind:  models.RoomKindGroup,
			Title: title,
			Topic: strings.TrimSpace(topic),
		}
		if err := tx.Create(&room).Error; err != nil {
			return err
		}
		return tx.Create(&models.Membership{
			RoomID:   room.ID,
			TenantID: creatorTenantID,
			Role:     models.MembershipRoleAdmin,
			JoinedAt: time.Now(),
		}).Error
	})
	return room, err
}
