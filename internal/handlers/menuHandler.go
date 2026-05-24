package handlers

import (
	"io"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-menu-service/internal/domain"
	"github.com/ssklv/mixfood-menu-service/internal/usecase"

	"github.com/gofiber/contrib/v3/swaggerui"
	_ "github.com/ssklv/mixfood-menu-service/docs"
)

type FileStorage interface {
	SaveFile(file io.Reader, filename string) (string, error)
}

type Logger interface {
	Error(msg string, fields ...any)
	Warn(msg string, fields ...any)
}

type MenuHandler interface {
	RegisterRoutes(app *fiber.App)
	AuthMiddleware() fiber.Handler
}

type menuHandler struct {
	usecase       usecase.MenuUsecase
	tokenProvider usecase.TokenProvider
	log           Logger
	fileStorage   FileStorage
}

func NewMenuHandler(uc usecase.MenuUsecase, tp usecase.TokenProvider, log Logger, fs FileStorage) MenuHandler {
	return &menuHandler{usecase: uc, tokenProvider: tp, log: log, fileStorage: fs}
}

func (mh *menuHandler) AuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Токен отсутствует"})
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		userID, role, err := mh.tokenProvider.ParseToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неверный токен"})
		}
		c.Locals("userID", userID)
		c.Locals("userRole", role)
		return c.Next()
	}
}

func (mh *menuHandler) RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals("userRole").(string)
		if !ok || userRole == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ запрещен"})
		}
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "У вас недостаточно прав"})
	}
}

func (mh *menuHandler) RegisterRoutes(app *fiber.App) {
	app.Use(swaggerui.New(swaggerui.Config{BasePath: "/", FilePath: "./docs/swagger.json", Path: "swagger"}))
	menu := app.Group("/api/menu")
	menu.Get("/", mh.getAllDishes)
	menu.Get("/:id", mh.getDishByID)
	menu.Post("/upload", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.uploadImage)
	menu.Post("/", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.createDish)
	menu.Patch("/:id", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.updateDish)
	menu.Delete("/:id", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.deleteDish)
}

func (mh *menuHandler) createDish(c fiber.Ctx) error {
	var dish domain.Dish
	if err := c.Bind().Body(&dish); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if err := mh.usecase.CreateDish(c.Context(), &dish); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(dish)
}

func (mh *menuHandler) getAllDishes(c fiber.Ctx) error {
	dishes, err := mh.usecase.GetAllDishes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
	return c.JSON(dishes)
}

func (mh *menuHandler) updateDish(c fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	var params domain.UpdateDishParams
	if err := c.Bind().Body(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	params.ID = id
	updated, err := mh.usecase.UpdateDish(c.Context(), &params)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(updated)
}

func (mh *menuHandler) getDishByID(c fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	dish, err := mh.usecase.GetDishByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(dish)
}

func (mh *menuHandler) deleteDish(c fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := mh.usecase.DeleteDish(c.Context(), id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

func (mh *menuHandler) uploadImage(c fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "image required"})
	}
	src, _ := file.Open()
	defer src.Close()
	url, err := mh.fileStorage.SaveFile(src, file.Filename)
	if err != nil {
		mh.log.Error("failed to save", err)
		return c.Status(500).JSON(fiber.Map{"error": "save failed"})
	}
	return c.JSON(fiber.Map{"url": url})
}
