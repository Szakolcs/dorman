package tenantviews

import (
	"net/url"
	"strings"

	"dorm-man/internal/maintenance"
	"dorm-man/internal/models"
)

type TicketListFilter struct {
	From       string
	To         string
	Statuses   []string
	Severities []string
}

func TicketListFilterFromQuery(filter maintenance.TicketFilter, from, to string) TicketListFilter {
	out := TicketListFilter{From: from, To: to}
	for _, s := range filter.Statuses {
		out.Statuses = append(out.Statuses, string(s))
	}
	for _, s := range filter.Severities {
		out.Severities = append(out.Severities, string(s))
	}
	return out
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
	for _, s := range f.Severities {
		values.Add("severity", s)
	}
	query := values.Encode()
	if query == "" {
		return "/tenant/tickets"
	}
	return "/tenant/tickets?" + query
}

func (f TicketListFilter) ToggleStatus(value string) TicketListFilter {
	f.Statuses = toggleValue(f.Statuses, value)
	return f
}

func (f TicketListFilter) ToggleSeverity(value string) TicketListFilter {
	f.Severities = toggleValue(f.Severities, value)
	return f
}

func toggleValue(values []string, value string) []string {
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

func impactLabel(impact models.Impact) string {
	switch impact {
	case models.ImpactLifeThreatening:
		return "Life threatening"
	case models.ImpactAffectsDaily:
		return "Affects daily life"
	case models.ImpactInconvenience:
		return "Inconvenience"
	case models.ImpactBeautyFlaw:
		return "Beauty flaw"
	default:
		return string(impact)
	}
}

func ticketStatusClass(status models.Status) string {
	return "dm-ticket-status-" + string(status)
}

func ticketCardTitle(ticket models.Ticket) string {
	line := ticket.Description
	if idx := strings.IndexAny(line, "\r\n"); idx >= 0 {
		line = line[:idx]
	}
	line = strings.TrimSpace(line)
	if len(line) > 96 {
		return line[:93] + "..."
	}
	if line != "" {
		return line
	}
	return severityLabel(ticket.Severity)
}

func ticketReporterLabel(ticket models.Ticket) string {
	if ticket.CreatedByUser != nil && ticket.CreatedByUser.Name != "" {
		return ticket.CreatedByUser.Name
	}
	return "Unknown reporter"
}

var AllTicketStatuses = []models.Status{
	models.StatusReported,
	models.StatusDuplicate,
	models.StatusInProgress,
	models.StatusHalted,
	models.StatusResolved,
	models.StatusClosed,
}

var AllTicketSeverities = []models.Severity{
	models.SeverityHigh,
	models.SeverityMedium,
	models.SeverityLow,
}
