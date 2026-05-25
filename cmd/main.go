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
// @description                 API для управления категориями и блюдами меню
// @host                        localhost:8082
// @BasePath                    /

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Введите токен в формате: Bearer <token>

// @securityDefinitions.apikey  CookieAuth
// @in                          cookie
// @name                        access_token
// @description                 Токен доступа (access_token), автоматически извлекаемый из Cookie
func main() {
	logger.InitLogger()
	if logger.Logger != nil {
		defer logger.Logger.Sync()
	}

	if err := godotenv.Load(); err != nil && logger.Logger != nil {
		logger.Logger.Warn("Файл .env не найден")
	}

	cfg := config.Load()
	logAdapter := &zapAdapter{}

	conn, err := infrastructure.Connect(cfg.DatabaseURL)
	if err != nil && logger.Logger != nil {
		logger.Logger.Fatal("Ошибка подключения к БД: " + err.Error())
	}
	defer conn.Close()

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	fileStorage, err := infrastructure.NewLocalFileStorage("./uploads")
	if err != nil && logger.Logger != nil {
		logger.Logger.Fatal("Ошибка инициализации локального хранилища: " + err.Error())
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
		logger.Logger.Info(fmt.Sprintf("Сервер меню запущен на порту :%s", cfg.ServerPort))
	}

	if err := app.Listen(":" + cfg.ServerPort); err != nil && logger.Logger != nil {
		logger.Logger.Fatal("Критическая ошибка сервера: " + err.Error())
	}
}
