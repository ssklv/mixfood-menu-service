package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-menu-service/internal/domain"
	"github.com/ssklv/mixfood-menu-service/internal/infrastructure"
	"github.com/ssklv/mixfood-menu-service/internal/usecase"

	_ "github.com/ssklv/mixfood-menu-service/docs"
)

type MenuHandler interface {
	RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler)
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

func (mh *menuHandler) RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals("userRole").(string)
		if !ok || userRole == "" {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Доступ запрещен"})
		}
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "У вас недостаточно прав"})
	}
}

func (mh *menuHandler) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	menu := router.Group("/menu")

	menu.Get("/", mh.getAllDishes)
	menu.Get("/:id", mh.getDishByID)

	menu.Post("/", authMiddleware, mh.RequireRole("admin"), mh.createDish)
	menu.Patch("/:id", authMiddleware, mh.RequireRole("admin"), mh.updateDish)
	menu.Delete("/:id", authMiddleware, mh.RequireRole("admin"), mh.deleteDish)
	menu.Post("/upload", authMiddleware, mh.RequireRole("admin"), mh.uploadImage)
}

// @Summary Создать новое блюдо
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param dish body domain.Dish true "Данные создаваемого блюда"
// @Success 201 {object} domain.Dish
// @Failure 400 {object} ErrorResponse
// @Router /api/menu [post]
func (mh *menuHandler) createDish(c fiber.Ctx) error {
	var dish domain.Dish
	if err := c.Bind().Body(&dish); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Некорректное тело запроса"})
	}

	if err := mh.usecase.CreateDish(c.Context(), &dish); err != nil {
		return mh.handleError(c, err)
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
		mh.log.Error("failed to get all dishes", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Внутренняя ошибка сервера"})
	}
	return c.JSON(dishes)
}

// @Summary Получить блюдо по ID
// @Produce json
// @Param id path int true "Идентификатор блюда"
// @Success 200 {object} domain.Dish
// @Router /api/menu/{id} [get]
func (mh *menuHandler) getDishByID(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Некорректный идентификатор блюда"})
	}

	dish, err := mh.usecase.GetDishByID(c.Context(), id)
	if err != nil {
		return mh.handleError(c, err)
	}
	return c.JSON(dish)
}

// @Summary Обновить параметры блюда
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Идентификатор блюда"
// @Param params body domain.UpdateDishParams true "Поля для обновления"
// @Success 200 {object} domain.Dish
// @Router /api/menu/{id} [patch]
func (mh *menuHandler) updateDish(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Некорректный идентификатор блюда"})
	}

	var params domain.UpdateDishParams
	if err := c.Bind().Body(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Некорректное тело запроса"})
	}
	params.ID = id

	updated, err := mh.usecase.UpdateDish(c.Context(), &params)
	if err != nil {
		return mh.handleError(c, err)
	}
	return c.JSON(updated)
}

// @Summary Удалить блюдо из меню
// @Security BearerAuth
// @Param id path int true "Идентификатор блюда"
// @Success 200 {object} map[string]string
// @Router /api/menu/{id} [delete]
func (mh *menuHandler) deleteDish(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Некорректный идентификатор блюда"})
	}

	if err := mh.usecase.DeleteDish(c.Context(), id); err != nil {
		return mh.handleError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Блюдо успешно удалено из меню"})
}

// @Summary Загрузить изображение
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Файл изображения"
// @Success 200 {object} map[string]string
// @Router /api/menu/upload [post]
func (mh *menuHandler) uploadImage(c fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Необходимо передать файл в параметре image"})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Не удалось прочитать загруженный файл"})
	}
	defer src.Close()

	url, err := mh.fileStorage.SaveFile(src, file.Filename)
	if err != nil {
		mh.log.Error("failed to store uploaded image", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Не удалось сохранить файл на сервере"})
	}

	return c.JSON(fiber.Map{"url": url})
}

func (mh *menuHandler) handleError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, usecase.ErrInvalidName),
		errors.Is(err, usecase.ErrInvalidCategory),
		errors.Is(err, usecase.ErrInvalidPrice),
		errors.Is(err, usecase.ErrInvalidWeight),
		errors.Is(err, usecase.ErrInvalidVolume),
		errors.Is(err, usecase.ErrInvalidBJU),
		errors.Is(err, usecase.ErrInvalidImageURL),
		errors.Is(err, usecase.ErrMeasurementMismatch):
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})

	case errors.Is(err, infrastructure.ErrDishNotFound):
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Запрашиваемое блюдо не найдено"})

	case errors.Is(err, infrastructure.ErrNoChanges):
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Нет данных для обновления"})

	default:
		mh.log.Error("unexpected internal error", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Внутренняя ошибка сервера"})
	}
}
