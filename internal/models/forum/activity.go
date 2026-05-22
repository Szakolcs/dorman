package forum

import (
	"dorm-man/internal/models/crosscutting"

	"github.com/google/uuid"
)

// Activity is a forum post that participants can book a slot for. Each
// activity has a fixed Capacity; BookedCount is a denormalized counter that
// must always satisfy 0 <= BookedCount <= Capacity (enforced at the service
// layer and via the ActivityBooking unique index).
type Activity struct {
	Post
	SharedAreaID *uuid.UUID `gorm:"type:uuid;index"`
	Capacity     int        `gorm:"not null;default:1"`
	BookedCount  int        `gorm:"not null;default:0"`

	Bookings []ActivityBooking `gorm:"foreignKey:ActivityID"`
}

// ActivityBooking is a single user reserving a spot on an Activity. The
// (ActivityID, UserID) pair is unique so a user cannot double-book.
type ActivityBooking struct {
	crosscutting.BaseModel
	ActivityID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_activity_user_booking"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_activity_user_booking"`

	Activity *Activity          `gorm:"foreignKey:ActivityID;references:ID"`
	User     *crosscutting.User `gorm:"foreignKey:UserID;references:ID"`
}
