package chat

import (
	"errors"

	"dorm-man/internal/models"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized     = errors.New("authorization denied")
	ErrNotFound         = errors.New("not found")
	ErrValidation       = errors.New("validation error")
	ErrNotAMember       = errors.New("not_a_member")
	ErrNotGroupOwner    = errors.New("not_group_owner")
	ErrFlatMembership   = errors.New("flat_membership_managed")
	ErrTenantNotFound   = errors.New("tenant_not_found")
	ErrRoomKindMismatch = errors.New("room_kind_mismatch")
)

type SearchResult struct {
	Rooms      []models.Membership
	Tenants    []models.Tenant
	Query      string
	RoomTitles map[uuid.UUID]string
}

func (r SearchResult) RoomTitle(roomID uuid.UUID) string {
	if r.RoomTitles == nil {
		return ""
	}
	return r.RoomTitles[roomID]
}

type CreateGroupRequest struct {
	Title     string
	Topic     string
	MemberIDs []uuid.UUID
}
