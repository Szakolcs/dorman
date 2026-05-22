package administration

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
