package app

import (
	"fmt"
	"net/http"

	"dorm-man/internal/administration"
	"dorm-man/internal/chat"
	"dorm-man/internal/config"
	"dorm-man/internal/doorman"
	"dorm-man/internal/forum"
	models "dorm-man/internal/models/administration"
	doormanModels "dorm-man/internal/models/doorman"
	chatModels "dorm-man/internal/models/chat"
	forumModels "dorm-man/internal/models/forum"
	"dorm-man/internal/platform"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	toMigrate = append(toMigrate, forumModels.All()...)
	toMigrate = append(toMigrate, chatModels.All()...)
	toMigrate = append(toMigrate, doormanModels.All()...)
	if err := db.AutoMigrate(toMigrate...); err != nil {
		return nil, fmt.Errorf("auto migrate models: %w", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())

	administration.RegisterRoutes(e, db)
	chat.RegisterRoutes(e, db)
	doorman.RegisterRoutes(e, db)
	forum.RegisterRoutes(e, db)
	platform.RegisterRoutes(e, db)

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
