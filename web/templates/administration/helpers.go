package administration

import (
	"fmt"
	"time"

	adm "dorm-man/internal/models/administration"
)

func inventoryLocation(item adm.InventoryItem) string {
	if item.Room != nil && item.Room.Number != "" {
		return fmt.Sprintf("Room %s", item.Room.Number)
	}
	if item.Flat != nil && item.Flat.Name != "" {
		return item.Flat.Name
	}
	if item.Building != nil && item.Building.Name != "" {
		return item.Building.Name
	}
	return "—"
}

func tenantStatus(active bool) string {
	if active {
		return "active"
	}
	return "inactive"
}

func userName(u adm.User) string {
	if u.Name != "" {
		return u.Name
	}
	return u.ID.String()
}

func userNamePtr(u *adm.User) string {
	if u == nil {
		return "—"
	}
	return userName(*u)
}

func publishDate(t *time.Time) string {
	if t == nil {
		return "—"
	}
	return t.Format("2006-01-02")
}

func tenantsPreserve(data TenantsPageData) map[string]string {
	m := map[string]string{}
	if data.Search != "" {
		m["search"] = data.Search
	}
	if data.Status != "" {
		m["status"] = data.Status
	}
	return m
}

func roomsPreserve(data RoomsPageData) map[string]string {
	m := map[string]string{}
	if data.State != "" {
		m["state"] = data.State
	}
	if data.Search != "" {
		m["search"] = data.Search
	}
	if data.PlanJSON != "" {
		m["plan_json"] = data.PlanJSON
	}
	return m
}

func jobsPreserve(data JobsPageData) map[string]string {
	m := map[string]string{}
	if data.AssigneeUserID != "" {
		m["assignee_user_id"] = data.AssigneeUserID
	}
	if data.Date != "" {
		m["date"] = data.Date
	}
	return m
}

func maintenancePreserve(data MaintenancePageData) map[string]string {
	m := map[string]string{}
	if data.ApprovalState != "" {
		m["approval_state"] = data.ApprovalState
	}
	if data.Status != "" {
		m["status"] = data.Status
	}
	return m
}

func auditActor(e adm.AuditEvent) string {
	if e.ActorUser != nil && e.ActorUser.Name != "" {
		return e.ActorUser.Name
	}
	if e.ActorUserID != nil {
		return e.ActorUserID.String()
	}
	return "—"
}
