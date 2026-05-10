package app

import (
	"dorm-man/internal/config"
	"dorm-man/internal/database"
	"dorm-man/internal/router"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Server struct {
	echo   *echo.Echo
	config config.Config
	db     *gorm.DB
}

func New() (*Server, error) {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	e := echo.New()
	router.Register(e)

	return &Server{
		echo:   e,
		config: cfg,
		db:     db,
	}, nil
}

func (s *Server) Start() error {
	return s.echo.Start(":" + s.config.Port)
}
