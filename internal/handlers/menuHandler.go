package handlers

import (
	"io"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-menu-service/internal/domain"
	"github.com/ssklv/mixfood-menu-service/internal/usecase"

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
		tokenStr := c.Cookies("access_token")
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Токен отсутствует"})
		}

		userID, role, err := mh.tokenProvider.ParseToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Ошибка валидации"})
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
	menu := app.Group("/api/menu")
	menu.Get("/", mh.getAllDishes)
	menu.Get("/:id", mh.getDishByID)
	menu.Post("/upload", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.uploadImage)
	menu.Post("/", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.createDish)
	menu.Patch("/:id", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.updateDish)
	menu.Delete("/:id", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.deleteDish)
}

// @Summary Создать блюдо
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param dish body domain.Dish true "Данные блюда"
// @Success 201 {object} domain.Dish
// @Router /api/menu [post]
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

// @Summary Получить все блюда
// @Produce json
// @Success 200 {array} domain.Dish
// @Router /api/menu [get]
func (mh *menuHandler) getAllDishes(c fiber.Ctx) error {
	dishes, err := mh.usecase.GetAllDishes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
	return c.JSON(dishes)
}

// @Summary Обновить блюдо
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID блюда"
// @Param params body domain.UpdateDishParams true "Параметры обновления"
// @Success 200 {object} domain.Dish
// @Router /api/menu/{id} [patch]
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

// @Summary Получить блюдо по ID
// @Produce json
// @Param id path int true "ID блюда"
// @Success 200 {object} domain.Dish
// @Router /api/menu/{id} [get]
func (mh *menuHandler) getDishByID(c fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	dish, err := mh.usecase.GetDishByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(dish)
}

// @Summary Удалить блюдо
// @Security BearerAuth
// @Param id path int true "ID блюда"
// @Success 200 {object} map[string]string
// @Router /api/menu/{id} [delete]
func (mh *menuHandler) deleteDish(c fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := mh.usecase.DeleteDish(c.Context(), id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

// @Summary Загрузить изображение
// @Security BearerAuth
// @Produce json
// @Param image formData file true "Изображение"
// @Success 200 {object} map[string]string
// @Router /api/menu/upload [post]
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
