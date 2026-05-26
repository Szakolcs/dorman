package models

// ChatRoom is a chatroom. Memberships are limited to tenants (see Membership);
// the room itself stays domain-agnostic so forum.Event can link to one via
// its ChatRoomID without creating an import cycle.
//
// DirectPairKey is set only for RoomKindDirect rooms and contains the two
// participating tenant UUIDs joined in lexicographic order ("<lo>:<hi>").
// The unique index guarantees that any pair of tenants has at most one
// direct chat. Group and event rooms keep this NULL.
type ChatRoom struct {
	BaseModel
	Kind          RoomKind `gorm:"type:varchar(20);not null;index"`
	Title         string   `gorm:"index"`
	Topic         string   `gorm:"type:text"`
	DirectPairKey *string  `gorm:"type:varchar(80);uniqueIndex"`

	Memberships []Membership `gorm:"foreignKey:RoomID"`
	Messages    []Message    `gorm:"foreignKey:RoomID"`
}
