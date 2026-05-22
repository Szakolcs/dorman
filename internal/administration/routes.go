package administration

import (
	"dorm-man/internal/middleware"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB, auth echo.MiddlewareFunc) {
	store := NewStore(db)
	service := NewService(store)
	handler := NewHandler(service)

	admin := e.Group("/administration")
	admin.Use(auth, middleware.RequireRole("admin", "dev"))
	admin.GET("", handler.dashboardPage)
	admin.GET("/tenants", handler.tenantsPage)
	admin.GET("/tenants/:id", handler.tenantDetailPage)
	admin.GET("/inventory", handler.inventoryPage)
	admin.GET("/inventory/:id", handler.inventoryDetailPage)
	admin.POST("/inventory", handler.createInventoryItem)
	admin.PUT("/inventory/:id/status", handler.updateInventoryItemStatus)
	admin.DELETE("/inventory/:id", handler.deleteInventoryItem)
	admin.GET("/jobs", handler.jobsPage)
	admin.GET("/jobs/:id", handler.jobDetailPage)
	admin.POST("/jobs", handler.createJob)
	admin.PUT("/jobs", handler.updateJob)
	admin.DELETE("/jobs", handler.DeleteJob)
	admin.GET("/publications", handler.publicationsPage)
	admin.GET("/publications/:id", handler.publicationDetailPage)
	admin.POST("/publications/news", handler.createNews)
	admin.DELETE("/publications/news/:id", handler.archiveNews)
	admin.POST("/publications/activities", handler.createActivity)
	admin.DELETE("/publications/activities/:id", handler.archiveActivity)
	admin.POST("/publications/events", handler.createEvent)
	admin.DELETE("/publications/events/:id", handler.archiveEvent)
	admin.GET("/housing", handler.housingPage)
	admin.GET("/housing/building/:id", handler.buildingDetail)
	admin.GET("/housing/flat/:id", handler.flatDetail)
	admin.GET("/housing/shared/:id", handler.sharedAreaDetail)
	admin.GET("/housing/room/:id", handler.roomDetail)
	admin.POST("/housing/room/assign", handler.assign)
	admin.POST("/housing/room/mass_assign", handler.massAssignment)
	admin.PUT("/housing/room/assign", handler.updateAssignment)
	admin.DELETE("/housing/room/assign", handler.deleteAssignment)
	admin.GET("/register", handler.registerUserPage)
	admin.POST("/register/new", handler.registerUser)
	admin.PUT("/register/update", handler.updateUser)
	admin.GET("/audit", handler.auditPage)
}
