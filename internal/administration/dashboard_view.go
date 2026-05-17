package administration

import (
	"time"

	adminviews "dorm-man/web/templates/administration"
)

func dashboardStatsView(s DashboardStats) adminviews.DashboardStatsView {
	return adminviews.DashboardStatsView{
		TenantsTotal:           s.TenantsTotal,
		TenantsActive:          s.TenantsActive,
		TenantsInactive:        s.TenantsInactive,
		TenantsUnassigned:      s.TenantsUnassigned,
		RoomsTotal:             s.RoomsTotal,
		RoomsOccupied:          s.RoomsOccupied,
		RoomsAvailable:         s.RoomsAvailable,
		RoomsMaintenanceNeeded: s.RoomsMaintenanceNeeded,
		InventoryTotal:         s.InventoryTotal,
		InventoryAttention:     s.InventoryAttention,
		MaintenanceTotal:       s.MaintenanceTotal,
		MaintenancePending:     s.MaintenancePending,
		MaintenanceOpen:        s.MaintenanceOpen,
		JobsTotal:              s.JobsTotal,
		JobsToday:              s.JobsToday,
		NewsTotal:              s.NewsTotal,
		ActivitiesTotal:        s.ActivitiesTotal,
		EventsTotal:            s.EventsTotal,
		AuditTotal:             s.AuditTotal,
	}
}

func dashboardPageData(
	stats DashboardStats,
	roomRows []RoomOldestAssignedTicket,
	openTickets []DashboardOpenMaintenanceTicket,
	jobsToday []DashboardJobToday,
	publications []DashboardPublication,
	events []DashboardEventRow,
) adminviews.DashboardPageData {
	return adminviews.DashboardPageData{
		Stats:                  dashboardStatsView(stats),
		JobsTodayDate:          time.Now().UTC().Format("2006-01-02"),
		RoomsOldestMaintenance: dashboardRoomMaintenanceRows(roomRows),
		OpenMaintenance:        dashboardOpenMaintenanceRows(openTickets),
		JobsToday:              dashboardJobTodayRows(jobsToday),
		RecentPublications:     dashboardPublicationRows(publications),
		RecentEvents:           dashboardEventRows(events),
	}
}

func dashboardRoomMaintenanceRows(rows []RoomOldestAssignedTicket) []adminviews.DashboardRoomMaintenanceRow {
	out := make([]adminviews.DashboardRoomMaintenanceRow, len(rows))
	for i, row := range rows {
		out[i] = adminviews.DashboardRoomMaintenanceRow{
			RoomID:     row.RoomID.String(),
			RoomNumber: row.RoomNumber,
			TicketID:   row.TicketID.String(),
			Status:     string(row.Status),
		}
	}
	return out
}

func dashboardOpenMaintenanceRows(rows []DashboardOpenMaintenanceTicket) []adminviews.DashboardOpenMaintenanceRow {
	out := make([]adminviews.DashboardOpenMaintenanceRow, len(rows))
	for i, row := range rows {
		out[i] = adminviews.DashboardOpenMaintenanceRow{
			TicketID:   row.ID.String(),
			Category:   string(row.Category),
			Severity:   string(row.Severity),
			Status:     string(row.Status),
			RoomNumber: row.RoomNumber,
		}
	}
	return out
}

func dashboardJobTodayRows(rows []DashboardJobToday) []adminviews.DashboardJobTodayRow {
	out := make([]adminviews.DashboardJobTodayRow, len(rows))
	for i, row := range rows {
		out[i] = adminviews.DashboardJobTodayRow{
			JobID:        row.ID.String(),
			Title:        row.Title,
			AssigneeName: row.AssigneeName,
			StartsAt:     formatDashboardDateTime(row.StartsAt),
			EndsAt:       formatDashboardDateTime(row.EndsAt),
			Status:       string(row.Status),
		}
	}
	return out
}

func dashboardPublicationRows(rows []DashboardPublication) []adminviews.DashboardPublicationRow {
	out := make([]adminviews.DashboardPublicationRow, len(rows))
	for i, row := range rows {
		out[i] = adminviews.DashboardPublicationRow{
			ID:        row.ID.String(),
			Kind:      row.Kind,
			Title:     row.Title,
			State:     string(row.State),
			CreatedAt: formatDashboardDateTime(row.CreatedAt),
		}
	}
	return out
}

func dashboardEventRows(rows []DashboardEventRow) []adminviews.DashboardEventRow {
	out := make([]adminviews.DashboardEventRow, len(rows))
	for i, row := range rows {
		out[i] = adminviews.DashboardEventRow{
			EventID:  row.ID.String(),
			Title:    row.Title,
			StartsAt: formatDashboardDateTime(row.StartsAt),
			EndsAt:   formatDashboardDateTime(row.EndsAt),
			State:    string(row.State),
		}
	}
	return out
}

func formatDashboardDateTime(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04")
}
