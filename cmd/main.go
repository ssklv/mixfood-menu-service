package main

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/joho/godotenv"

	"github.com/ssklv/mixfood-menu-service/internal/config"
	"github.com/ssklv/mixfood-menu-service/internal/handlers"
	"github.com/ssklv/mixfood-menu-service/internal/infrastructure"
	"github.com/ssklv/mixfood-menu-service/internal/usecase"
	"github.com/ssklv/pizza-shared/pkg/logger"

	_ "github.com/ssklv/mixfood-menu-service/docs"
)

type zapAdapter struct{}

func (za *zapAdapter) Error(msg string, fields ...any) {
	if logger.Logger != nil {
		logger.Logger.Sugar().Errorw(msg, fields...)
	}
}

func (za *zapAdapter) Warn(msg string, fields ...any) {
	if logger.Logger != nil {
		logger.Logger.Sugar().Warnw(msg, fields...)
	}
}

// @title                       Mixfood Menu Service API
// @version                     1.0
// @description                 API for managing menu categories and dishes
// @host                        localhost:8082
// @BasePath                    /

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Enter token in format: Bearer <token>

// @securityDefinitions.apikey  CookieAuth
// @in                          cookie
// @name                        access_token
// @description                 Access token (access_token) automatically retrieved from cookies
func main() {
	logger.InitLogger()
	if logger.Logger != nil {
		defer logger.Logger.Sync()
	}

	if err := godotenv.Load(); err != nil && logger.Logger != nil {
		logger.Logger.Warn(".env file not found")
	}

	cfg := config.Load()
	logAdapter := &zapAdapter{}

	conn, err := infrastructure.Connect(cfg.DatabaseURL)
	if err != nil && logger.Logger != nil {
		logger.Logger.Fatal("Database connection error: " + err.Error())
	}
	defer conn.Close()

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	fileStorage, err := infrastructure.NewLocalFileStorage("./uploads")
	if err != nil && logger.Logger != nil {
		logger.Logger.Fatal("File storage initialization error: " + err.Error())
	}

	tokenProvider := infrastructure.NewTokenProvider(cfg.JWTSecret, 15)
	menuRepository := infrastructure.NewMenuRepository(conn, psql)
	menuUsecase := usecase.NewMenuUsecase(menuRepository)

	app := fiber.New(fiber.Config{
		AppName: "MixFood Menu Service",
	})

	app.Use("/uploads", static.New("./uploads"))

	handlers.ConfigureApp(app, menuUsecase, tokenProvider, logAdapter, fileStorage)

	if logger.Logger != nil {
		logger.Logger.Info(fmt.Sprintf("Menu service started on port :%s", cfg.ServerPort))
	}

	if err := app.Listen(":" + cfg.ServerPort); err != nil && logger.Logger != nil {
		logger.Logger.Fatal("Critical server error: " + err.Error())
	}
}
