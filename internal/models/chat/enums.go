package chat

// RoomKind distinguishes the three flavors of chatroom we support.
type RoomKind string

const (
	// RoomKindDirect is a 1:1 chat between exactly two tenants.
	RoomKindDirect RoomKind = "direct"
	// RoomKindGroup is a tenant-created group chat with N members.
	RoomKindGroup RoomKind = "group"
	// RoomKindEvent is the dedicated chatroom attached to a forum.Event.
	RoomKindEvent RoomKind = "event"
)

// MembershipRole controls who can manage a Room (rename, add/remove members,
// archive, etc). Direct rooms only ever have plain members.
type MembershipRole string

const (
	MembershipRoleMember MembershipRole = "member"
	MembershipRoleAdmin  MembershipRole = "admin"
)
