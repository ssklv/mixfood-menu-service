package main

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/joho/godotenv"

	"github.com/ssklv/mixfood-menu-service/internal/config"
	"github.com/ssklv/mixfood-menu-service/internal/handlers"
	"github.com/ssklv/mixfood-menu-service/internal/infrastructure"
	"github.com/ssklv/mixfood-menu-service/internal/usecase"
	"github.com/ssklv/pizza-shared/pkg/logger"
)

// @title Mixfood Menu API
// @version 1.0
// @description API для управления меню
// @host localhost:8082
// @BasePath /
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name jwt
func main() {
	logger.InitLogger()
	defer logger.Logger.Sync()
	_ = godotenv.Load()

	cfg := config.Load()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://localhost:8080",
			"http://localhost:8082",
			"http://localhost:8083",
		},
		AllowCredentials: true,
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
	}))

	app.Get("/swagger/*", func(c fiber.Ctx) error {
		if c.Path() == "/swagger/doc.json" {
			return c.SendFile("./docs/swagger.json")
		}
		c.Set("Content-Type", "text/html")
		return c.SendString(`<!DOCTYPE html>
            <html>
            <head><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.19.0/swagger-ui.css"></head>
            <body><div id="swagger-ui"></div>
            <script src="https://unpkg.com/swagger-ui-dist@5.19.0/swagger-ui-bundle.js"></script>
            <script>SwaggerUIBundle({url: "/swagger/doc.json", dom_id: '#swagger-ui'});</script>
            </body></html>`)
	})

	app.Use("/uploads", static.New("./uploads"))

	conn, err := infrastructure.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Logger.Fatal("Ошибка БД: " + err.Error())
	}
	defer conn.Close()

	fileStorage, err := infrastructure.NewLocalFileStorage("./uploads")
	if err != nil {
		logger.Logger.Fatal("Ошибка хранилища: " + err.Error())
	}

	tokenProvider := infrastructure.NewTokenProvider(cfg.JWTSecret, 15)
	menuUsecase := usecase.NewMenuUsecase(infrastructure.NewMenuRepository(conn, psql))

	handlers.NewMenuHandler(menuUsecase, tokenProvider, nil, fileStorage).RegisterRoutes(app)

	logger.Logger.Info(fmt.Sprintf("Сервер стартовал на :%s", cfg.ServerPort))
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		logger.Logger.Fatal("Сервер упал: " + err.Error())
	}
}
