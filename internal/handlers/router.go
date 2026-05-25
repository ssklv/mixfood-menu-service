package handlers

import (
	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/ssklv/mixfood-menu-service/internal/usecase"
)

func ConfigureApp(
	app *fiber.App,
	menuUC usecase.MenuUsecase,
	tokenProvider usecase.TokenProvider,
	log Logger,
	fileStorage FileStorage,
) {
	app.Use(recover.New())
	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:8080", "http://localhost:8082"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))

	app.Get("/swagger/*", swaggo.HandlerDefault)

	authMiddleware := NewAuthMiddleware(tokenProvider, log)
	apiGroup := app.Group("/api")

	NewMenuHandler(menuUC, tokenProvider, log, fileStorage).RegisterRoutes(apiGroup, authMiddleware)
}
