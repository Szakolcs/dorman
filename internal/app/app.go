package app

import (
	"dorm-man/internal/cross-cutting"
	"fmt"
	"net/http"

	"dorm-man/internal/config"
	administrationModels "dorm-man/internal/models/administration"
	chatModels "dorm-man/internal/models/chat"
	crosscuttingModels "dorm-man/internal/models/cross-cutting"
	doormanModels "dorm-man/internal/models/doorman"
	forumModels "dorm-man/internal/models/forum"

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

	toMigrate := append([]any(nil), crosscuttingModels.All()...)
	toMigrate = append(toMigrate, administrationModels.All()...)
	toMigrate = append(toMigrate, forumModels.All()...)
	toMigrate = append(toMigrate, chatModels.All()...)
	toMigrate = append(toMigrate, doormanModels.All()...)
	if err := db.AutoMigrate(toMigrate...); err != nil {
		return nil, fmt.Errorf("auto migrate models: %w", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Static("/static", "web/static")

	cross_cutting.RegisterRoutes(e, db)
	//administration.RegisterRoutes(e, db)
	//chat.RegisterRoutes(e, db)
	//doorman.RegisterRoutes(e, db)
	//forum.RegisterRoutes(e, db)
	//platform.RegisterRoutes(e, db)

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
