package main

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/ssklv/mixfood-menu-service/internal/config"
	"github.com/ssklv/mixfood-menu-service/internal/handlers"
	"github.com/ssklv/mixfood-menu-service/internal/infrastructure"
	"github.com/ssklv/mixfood-menu-service/internal/usecase"
	"github.com/ssklv/pizza-shared/pkg/logger"
)

type zapAdapter struct{}

func (za *zapAdapter) Error(msg string, fields ...any) {
	logger.Logger.Error(msg)
}

func (za *zapAdapter) Warn(msg string, fields ...any) {
	logger.Logger.Warn(msg)
}

func main() {
	logger.InitLogger()
	defer logger.Logger.Sync()
	logger.Logger.Info("Логгер из pizza-shared успешно запущен в сервисе меню!")

	if err := godotenv.Load(); err != nil {
		logger.Logger.Warn("Файл .env не найден, Go будет использовать системные переменные окружения")
	}

	cfg := config.Load()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	app := fiber.New(fiber.Config{
		AppName: "MixFood Menu Service v1.0",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
	}))
	conn, err := infrastructure.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Logger.Fatal("Критическая ошибка: не удалось подключиться к БД меню: " + err.Error())
	}
	logger.Logger.Info("Успешное подключение к базе данных PostgreSQL (Menu)")
	defer conn.Close()

	tokenProvider := infrastructure.NewTokenProvider(cfg.JWTSecret, 15)

	menuRepo := infrastructure.NewMenuRepository(conn, psql)
	menuUsecase := usecase.NewMenuUsecase(menuRepo)

	logAdapter := &zapAdapter{}

	menuHandler := handlers.NewMenuHandler(menuUsecase, tokenProvider, logAdapter)
	menuHandler.RegisterRoutes(app)

	logger.Logger.Info(fmt.Sprintf("Сервер MixFood Menu успешно стартовал на порту :%s", cfg.ServerPort))
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		logger.Logger.Fatal("Сервер меню аварийно завершил работу: " + err.Error())
	}
}
