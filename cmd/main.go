package main

import (
	"fmt"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors" // Импорт для статики
	"github.com/joho/godotenv"

	"github.com/ssklv/mixfood-menu-service/internal/config"
	"github.com/ssklv/mixfood-menu-service/internal/handlers"
	"github.com/ssklv/mixfood-menu-service/internal/infrastructure"
	"github.com/ssklv/mixfood-menu-service/internal/usecase"
	"github.com/ssklv/pizza-shared/pkg/logger"
)

type zapAdapter struct{}

func (za *zapAdapter) Error(msg string, fields ...any) { logger.Logger.Error(msg) }
func (za *zapAdapter) Warn(msg string, fields ...any)  { logger.Logger.Warn(msg) }

func main() {
	logger.InitLogger()
	defer logger.Logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Logger.Warn("Файл .env не найден")
	}

	cfg := config.Load()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	app := fiber.New()

	// Исправленная статика для Fiber v3
	app.Get("/uploads/*", adaptor.HTTPHandler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads")))))

	app.Get("/docs/*", adaptor.HTTPHandler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs")))))

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	conn, err := infrastructure.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Logger.Fatal("Ошибка БД: " + err.Error())
	}
	defer conn.Close()

	fileStorage, err := infrastructure.NewLocalFileStorage("./uploads")
	if err != nil {
		logger.Logger.Fatal("Ошибка инициализации хранилища: " + err.Error())
	}

	tokenProvider := infrastructure.NewTokenProvider(cfg.JWTSecret, 15)
	menuRepo := infrastructure.NewMenuRepository(conn, psql)
	menuUsecase := usecase.NewMenuUsecase(menuRepo)

	menuHandler := handlers.NewMenuHandler(menuUsecase, tokenProvider, &zapAdapter{}, fileStorage)
	menuHandler.RegisterRoutes(app)

	logger.Logger.Info(fmt.Sprintf("Сервер стартовал на :%s", cfg.ServerPort))
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		logger.Logger.Fatal("Сервер упал: " + err.Error())
	}
}
