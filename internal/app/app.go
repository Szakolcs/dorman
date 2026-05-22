package app

import (
	administrationModule "dorm-man/internal/administration"
	cross_cutting "dorm-man/internal/cross-cutting"
	doormanModule "dorm-man/internal/doorman"
	maintenanceModule "dorm-man/internal/maintenance"
	"dorm-man/internal/models"
	"fmt"
	"net/http"

	"dorm-man/internal/config"

	"dorm-man/internal/middleware"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type App struct {
	cfg    config.Config
	echo   *echo.Echo
	db     *gorm.DB
	server *http.Server
}

func New() (*App, error) {
	cfg := config.Load()

	db, err := config.OpenDB(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	toMigrate := append([]any(nil), models.All()...)

	if err := db.AutoMigrate(toMigrate...); err != nil {
		return nil, fmt.Errorf("auto migrate models: %w", err)
	}
	if err := config.CreateAuditInfrastructure(db); err != nil {
		return nil, fmt.Errorf("create audit infrastructure: %w", err)
	}
	e := echo.New()
	e.HideBanner = true
	e.Static("/static", "web/static")

	auth := middleware.Authenticate(cfg.SessionSecret)

	cross_cutting.RegisterRoutes(e, db, cfg)
	maintenanceModule.RegisterRoutes(e, db, auth)
	administrationModule.RegisterRoutes(e, db, auth)
	doormanModule.RegisterRoutes(e, db, auth)
	// forumModule.RegisterRoutes(e, db)
	// chatModule.RegisterRoutes(e, db)

	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	return &App{
		cfg:  cfg,
		echo: e,
		db:   db,
	}, nil
}

func (a *App) Start() error {
	addr := ":" + a.cfg.Port
	return a.echo.Start(addr)
}
