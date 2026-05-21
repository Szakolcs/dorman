package maintenance

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
