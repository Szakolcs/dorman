package doorman

type AccessStatus string

const (
	CheckedIn  AccessStatus = "checked_in"
	CheckedOut AccessStatus = "checked_out"
)
