package models

type DegreeType string

const (
	DegreeBSc DegreeType = "bsc"
	DegreeBA  DegreeType = "ba"
	DegreeMSc DegreeType = "msc"
	DegreeMA  DegreeType = "ma"
	DegreePhD DegreeType = "phd"
)

type SexType string

const (
	SexFemale SexType = "female"
	SexMale   SexType = "male"
	SexOther  SexType = "other"
)

type NationalityType string

const (
	NationalityHungarian     NationalityType = "hungarian"
	NationalityInternational NationalityType = "international"
)

type FacultyType string

const (
	FacultyScience     FacultyType = "science"
	FacultyHumanities  FacultyType = "humanities"
	FacultyEngineering FacultyType = "engineering"
	FacultyMedicine    FacultyType = "medicine"
)

type InventoryCondition string

const (
	InventoryConditionNew     InventoryCondition = "new"
	InventoryConditionGood    InventoryCondition = "good"
	InventoryConditionUsed    InventoryCondition = "used"
	InventoryConditionDamaged InventoryCondition = "damaged"
	InventoryConditionBroken  InventoryCondition = "broken"
)

type InventoryStatus string

const (
	InventoryStatusInStock   InventoryStatus = "in_stock"
	InventoryStatusInUse     InventoryStatus = "in_use"
	InventoryStatusWithdrawn InventoryStatus = "withdrawn"
	InventoryStatusDestroyed InventoryStatus = "destroyed"
)

type JobPriority string

const (
	JobPriorityLow    JobPriority = "low"
	JobPriorityMedium JobPriority = "medium"
	JobPriorityHigh   JobPriority = "high"
)

type JobStatus string

const (
	JobStatusPlanned    JobStatus = "planned"
	JobStatusScheduled  JobStatus = "scheduled"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusFinished   JobStatus = "finished"
	JobStatusCanceled   JobStatus = "canceled"
)

type Category string

const (
	CategoryYearly   Category = "yearly"
	CategoryMonthly  Category = "monthly"
	CategoryWeekly   Category = "weekly"
	CategoryIncident Category = "incident"
)

type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
	SeverityLow    Severity = "low"
)

type Impact string

const (
	ImpactLifeThreatening Impact = "life_threatening"
	ImpactAffectsDaily    Impact = "affects_daily_life"
	ImpactInconvenience   Impact = "inconvenience"
	ImpactBeautyFlaw      Impact = "beauty_flaw"
)

type Status string

const (
	StatusReported   Status = "reported"
	StatusDuplicate  Status = "duplicate"
	StatusInProgress Status = "in_progress"
	StatusHalted     Status = "halted"
	StatusResolved   Status = "resolved"
	StatusClosed     Status = "closed"
)

// EventAttendanceIntent is what a user signals about attending an event.
type EventAttendanceIntent string

const (
	EventAttendanceInterested    EventAttendanceIntent = "interested"
	EventAttendanceNotInterested EventAttendanceIntent = "not_interested"
	EventAttendanceBusy          EventAttendanceIntent = "busy"
)

type AccessStatus string

const (
	CheckedIn  AccessStatus = "checked_in"
	CheckedOut AccessStatus = "checked_out"
)

type PublicationState string

const (
	PublicationStateDraft     PublicationState = "draft"
	PublicationStatePublished PublicationState = "published"
	PublicationStateArchived  PublicationState = "archived"
	PublicationStateCanceled  PublicationState = "canceled"
	PublicationStatePostponed PublicationState = "postponed"
	PublicationStateHidden    PublicationState = "hidden"
)

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
