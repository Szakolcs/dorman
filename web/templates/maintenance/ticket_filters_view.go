package maintenanceviews

import (
	"net/url"

	"dorm-man/internal/models"
)

const maintenanceTicketsPanelID = "maintenance-tickets-panel"

type TicketListFilter struct {
	From       string
	To         string
	Statuses   []string
	Categories []string
	Severities []string
}

func (f TicketListFilter) URL() string {
	values := url.Values{}
	if f.From != "" {
		values.Set("from", f.From)
	}
	if f.To != "" {
		values.Set("to", f.To)
	}
	for _, s := range f.Statuses {
		values.Add("status", s)
	}
	for _, c := range f.Categories {
		values.Add("category", c)
	}
	for _, s := range f.Severities {
		values.Add("severity", s)
	}
	query := values.Encode()
	if query == "" {
		return "/maintenance/tickets"
	}
	return "/maintenance/tickets?" + query
}

func (f TicketListFilter) ToggleStatus(value string) TicketListFilter {
	f.Statuses = toggleFilterValue(f.Statuses, value)
	return f
}

func (f TicketListFilter) ToggleCategory(value string) TicketListFilter {
	f.Categories = toggleFilterValue(f.Categories, value)
	return f
}

func (f TicketListFilter) ToggleSeverity(value string) TicketListFilter {
	f.Severities = toggleFilterValue(f.Severities, value)
	return f
}

func toggleFilterValue(values []string, value string) []string {
	for i, v := range values {
		if v == value {
			return append(values[:i], values[i+1:]...)
		}
	}
	return append(values, value)
}

func filterValueActive(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func statusLabel(status models.Status) string {
	switch status {
	case models.StatusReported:
		return "Reported"
	case models.StatusDuplicate:
		return "Duplicate"
	case models.StatusInProgress:
		return "In progress"
	case models.StatusHalted:
		return "Halted"
	case models.StatusResolved:
		return "Resolved"
	case models.StatusClosed:
		return "Closed"
	default:
		return string(status)
	}
}

func categoryLabel(category models.Category) string {
	switch category {
	case models.CategoryYearly:
		return "Yearly"
	case models.CategoryMonthly:
		return "Monthly"
	case models.CategoryWeekly:
		return "Weekly"
	case models.CategoryIncident:
		return "Incident"
	default:
		return string(category)
	}
}

func severityLabel(severity models.Severity) string {
	switch severity {
	case models.SeverityHigh:
		return "High"
	case models.SeverityMedium:
		return "Medium"
	case models.SeverityLow:
		return "Low"
	default:
		return string(severity)
	}
}
