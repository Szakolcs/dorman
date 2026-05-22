package chat

import (
	"errors"
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
