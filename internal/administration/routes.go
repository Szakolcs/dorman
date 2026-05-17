package administration

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB) {
	store := NewStore(db)
	service := NewService(store)
	handler := &Handler{service: service}

	pages := e.Group("/administration")
	pages.GET("", handler.dashboardPage)
	pages.GET("/tenants", handler.tenantsPage)
	pages.GET("/tenants/:id", handler.tenantDetailPage)
	pages.POST("/tenants", handler.createTenantView)
	pages.POST("/tenants/:id/deactivate", handler.deactivateTenantView)
	pages.GET("/rooms", handler.roomsPage)
	pages.GET("/rooms/:id", handler.roomDetailPage)
	pages.POST("/rooms/assign", handler.assignTenantView)
	pages.POST("/rooms/allocation/generate", handler.generatePlanView)
	pages.POST("/rooms/allocation/approve", handler.approvePlanView)
	pages.GET("/inventory", handler.inventoryPage)
	pages.POST("/inventory", handler.createInventoryView)
	pages.POST("/inventory/:id/status", handler.updateInventoryStatusView)
	pages.GET("/maintenance", handler.maintenancePage)
	pages.GET("/maintenance/tickets/:id", handler.maintenanceTicketDetailPage)
	pages.POST("/maintenance/tickets", handler.createMaintenanceTicketView)
	pages.POST("/maintenance/tickets/:id/approve", handler.approveMaintenanceTicketView)
	pages.POST("/maintenance/tickets/:id/transition", handler.transitionMaintenanceTicketView)
	pages.GET("/jobs", handler.jobsPage)
	pages.GET("/jobs/:id", handler.jobDetailPage)
	pages.POST("/jobs", handler.createJobView)
	pages.GET("/publications", handler.publicationsPage)
	pages.GET("/publications/news/:id", handler.newsDetailPage)
	pages.GET("/publications/activities/:id", handler.activityDetailPage)
	pages.GET("/publications/events/:id", handler.eventDetailPage)
	pages.POST("/publications/news", handler.createNewsView)
	pages.POST("/publications/activities", handler.createActivityView)
	pages.POST("/publications/events", handler.createEventView)
	pages.GET("/audit", handler.auditPage)

	api := e.Group("/api/administration")

	api.GET("/tenants", handler.listTenants)
	api.GET("/tenants/:id", handler.getTenant)
	api.POST("/tenants", handler.createTenant)
	api.POST("/tenants/:id/activate", handler.activateTenant)
	api.POST("/tenants/:id/deactivate", handler.deactivateTenant)

	api.GET("/rooms", handler.listRooms)
	api.GET("/rooms/:id", handler.getRoom)
	api.POST("/assignments", handler.assignTenant)
	api.POST("/room-allocation-plan/generate", handler.generatePlan)
	api.POST("/room-allocation-plan/approve", handler.approvePlan)

	api.GET("/inventory", handler.listInventory)
	api.POST("/inventory", handler.createInventory)
	api.PATCH("/inventory/:id/status", handler.updateInventoryStatus)

	api.GET("/maintenance/tickets", handler.listMaintenanceTickets)
	api.POST("/maintenance/tickets", handler.createMaintenanceTicket)
	api.POST("/maintenance/tickets/:id/approve", handler.approveMaintenanceTicket)
	api.POST("/maintenance/tickets/:id/transition", handler.transitionMaintenanceTicket)

	api.GET("/jobs", handler.listJobs)
	api.POST("/jobs", handler.createJob)

	api.GET("/news", handler.listNews)
	api.POST("/news", handler.createNews)
	api.POST("/news/:id/publish", handler.publishNews)

	api.GET("/activities", handler.listActivities)
	api.POST("/activities", handler.createActivity)
	api.POST("/activities/:id/publish", handler.publishActivity)

	api.GET("/events", handler.listEvents)
	api.POST("/events", handler.createEvent)
	api.POST("/events/:id/state", handler.updateEventState)
}
