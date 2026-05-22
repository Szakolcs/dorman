package chat

import (
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
	result := SearchResult{Query: query}
	if query == "" {
		result.Rooms = memberships
		return result, nil
	}
	needle := strings.ToLower(query)
	for _, membership := range memberships {
		if membership.ChatRoom == nil {
			continue
		}
		title := strings.ToLower(membership.ChatRoom.Title)
		kind := strings.ToLower(string(membership.ChatRoom.Kind))
		if strings.Contains(title, needle) || strings.Contains(kind, needle) {
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

func (s *Service) OpenDirectChat(viewerTenantID, otherTenantID uuid.UUID) (models.ChatRoom, error) {
	if viewerTenantID == uuid.Nil || otherTenantID == uuid.Nil {
		return models.ChatRoom{}, ErrValidation
	}
	if viewerTenantID == otherTenantID {
		return models.ChatRoom{}, ErrValidation
	}
	return s.store.OpenDirectRoom(viewerTenantID, otherTenantID)
}

func (s *Service) CreateGroup(creatorTenantID uuid.UUID, req CreateGroupRequest) (models.ChatRoom, error) {
	return s.store.CreateGroupRoom(creatorTenantID, req.Title, req.Topic)
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
