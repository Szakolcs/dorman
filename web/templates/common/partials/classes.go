package partials

// Cls maps shared layout tokens to Tailwind component classes.
type classes struct {
	Shell            string
	Page             string
	PageHdr          string
	PageTitle        string
	PageSub          string
	Card             string
	CardTitle        string
	TableWrap        string
	Table            string
	TableHead        string
	TableTr          string
	TableTh          string
	TableTd          string
	Nav              string
	NavLink          string
	NavActions       string
	NavButton        string
	NavButtonDanger  string
	Form             string
	FormGrid         string
	FormActions      string
	Field            string
	Label            string
	Input            string
	Select           string
	Textarea         string
	Button           string
	ButtonSecondary  string
	ButtonDanger     string
	DetailRow        string
	DetailLabel      string
	DetailValue      string
	Stat             string
	StatLabel        string
	StatValue        string
	StatGrid         string
	SectionGrid      string
	SectionGrid3     string
	Link             string
	BackLink         string
	Empty            string
	LoginPanel       string
	FilterBar        string
	AlertError       string
}

var Cls = classes{
	Shell:           "dm-shell",
	Page:            "dm-page",
	PageHdr:         "dm-page-header",
	PageTitle:       "dm-page-title",
	PageSub:         "dm-page-subtitle",
	Card:            "dm-card",
	CardTitle:       "dm-card-title",
	TableWrap:       "dm-table-wrap",
	Table:           "dm-table",
	TableHead:       "",
	TableTr:         "",
	TableTh:         "",
	TableTd:         "",
	Nav:             "dm-nav",
	NavLink:         "dm-nav-link",
	NavActions:      "dm-nav-actions",
	NavButton:       "dm-nav-button",
	NavButtonDanger: "dm-nav-button-danger",
	Form:            "dm-form",
	FormGrid:        "dm-form-grid",
	FormActions:     "dm-form-actions",
	Field:           "dm-field",
	Label:           "dm-label",
	Input:           "dm-input",
	Select:          "dm-select",
	Textarea:        "dm-textarea",
	Button:          "dm-button",
	ButtonSecondary: "dm-button-secondary",
	ButtonDanger:    "dm-button dm-button-danger",
	DetailRow:       "dm-detail-row",
	DetailLabel:     "dm-detail-label",
	DetailValue:     "dm-detail-value",
	Stat:            "dm-stat",
	StatLabel:       "dm-stat-label",
	StatValue:       "dm-stat-value",
	StatGrid:        "dm-stat-grid",
	SectionGrid:     "dm-section-grid",
	SectionGrid3:    "dm-section-grid-3",
	Link:            "dm-link",
	BackLink:        "dm-back-link",
	Empty:           "dm-empty",
	LoginPanel:      "dm-login-panel",
	FilterBar:       "dm-filter-bar",
	AlertError:      "dm-alert dm-alert-error",
}
