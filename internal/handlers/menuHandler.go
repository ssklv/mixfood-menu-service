package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-menu-service/internal/domain"
	"github.com/ssklv/mixfood-menu-service/internal/infrastructure"
	"github.com/ssklv/mixfood-menu-service/internal/usecase"
)

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
}

const (
	accessCookie = "access_token"
)

func NewMenuHandler(uc usecase.MenuUsecase, tp usecase.TokenProvider, log Logger) MenuHandler {
	return &menuHandler{
		usecase:       uc,
		tokenProvider: tp,
		log:           log,
	}
}

type createDishReq struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Weight      *int     `json:"weight"`
	Volume      *float64 `json:"volume"`
	Proteins    float64  `json:"proteins"`
	Fats        float64  `json:"fats"`
	Carbs       float64  `json:"carbs"`
	Calories    int      `json:"calories"`
	ImageURL    string   `json:"image_url"`
}

type updateDishReq struct {
	Name        *string  `json:"name"`
	Category    *string  `json:"category"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Weight      *int     `json:"weight"`
	Volume      *float64 `json:"volume"`
	Proteins    *float64 `json:"proteins"`
	Fats        *float64 `json:"fats"`
	Carbs       *float64 `json:"carbs"`
	Calories    *int     `json:"calories"`
	ImageURL    *string  `json:"image_url"`
}

func (mh *menuHandler) AuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		tokenStr := c.Cookies(accessCookie)
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Вы не авторизованы (токен отсутствует)"})
		}
		userID, role, err := mh.tokenProvider.ParseToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Сессия устарела или токен невалиден"})
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
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ запрещен (роль не определена)"})
		}

		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				return c.Next()
			}
		}

		mh.log.Warn("Попытка несанкционированного доступа к меню", "user_id", c.Locals("userID"), "role", userRole)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "У вас недостаточно прав"})
	}
}

func (mh *menuHandler) RegisterRoutes(app *fiber.App) {
	menu := app.Group("/api/menu")

	menu.Get("/", mh.getAllDishes)
	menu.Get("/:id", mh.getDishByID)

	menu.Post("/", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.createDish)
	menu.Patch("/:id", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.updateDish)
	menu.Delete("/:id", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.deleteDish)
}

func (mh *menuHandler) createDish(c fiber.Ctx) error {
	var req createDishReq
	if err := c.Bind().Body(&req); err != nil {
		mh.log.Error("invalid request body in createDish", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	dish := &domain.Dish{
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
		Price:       req.Price,
		Weight:      req.Weight,
		Volume:      req.Volume,
		Proteins:    req.Proteins,
		Fats:        req.Fats,
		Carbs:       req.Carbs,
		Calories:    req.Calories,
		ImageURL:    req.ImageURL,
	}

	if err := mh.usecase.CreateDish(c.Context(), dish); err != nil {
		if errors.Is(err, usecase.ErrInvalidName) ||
			errors.Is(err, usecase.ErrInvalidCategory) ||
			errors.Is(err, usecase.ErrInvalidPrice) ||
			errors.Is(err, usecase.ErrInvalidWeight) ||
			errors.Is(err, usecase.ErrInvalidVolume) ||
			errors.Is(err, usecase.ErrInvalidBJU) ||
			errors.Is(err, usecase.ErrInvalidImageURL) ||
			errors.Is(err, usecase.ErrMeasurementMismatch) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		mh.log.Error("failed to create dish", err, "name", req.Name)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(dish)
}

func (mh *menuHandler) updateDish(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid dish id"})
	}

	var req updateDishReq
	if err := c.Bind().Body(&req); err != nil {
		mh.log.Error("invalid request body in updateDish", err, "dishID", id)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	params := &domain.UpdateDishParams{
		ID:          id,
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
		Price:       req.Price,
		Proteins:    req.Proteins,
		Fats:        req.Fats,
		Carbs:       req.Carbs,
		Calories:    req.Calories,
		ImageURL:    req.ImageURL,
	}

	if req.Weight != nil {
		params.Weight = &req.Weight
	}
	if req.Volume != nil {
		params.Volume = &req.Volume
	}

	updatedDish, err := mh.usecase.UpdateDish(c.Context(), params)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidName) ||
			errors.Is(err, usecase.ErrInvalidCategory) ||
			errors.Is(err, usecase.ErrInvalidPrice) ||
			errors.Is(err, usecase.ErrInvalidWeight) ||
			errors.Is(err, usecase.ErrInvalidVolume) ||
			errors.Is(err, usecase.ErrInvalidBJU) ||
			errors.Is(err, usecase.ErrMeasurementMismatch) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		if errors.Is(err, infrastructure.ErrDishNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "dish not found"})
		}

		mh.log.Error("failed to update dish", err, "dishID", id)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(updatedDish)
}

func (mh *menuHandler) getDishByID(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid dish id"})
	}

	dish, err := mh.usecase.GetDishByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, infrastructure.ErrDishNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "dish not found"})
		}

		mh.log.Error("failed to get dish by id", err, "dishID", id)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(dish)
}

func (mh *menuHandler) getAllDishes(c fiber.Ctx) error {
	dishes, err := mh.usecase.GetAllDishes(c.Context())
	if err != nil {
		mh.log.Error("failed to get all dishes", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(dishes)
}

func (mh *menuHandler) deleteDish(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid dish id"})
	}

	if err := mh.usecase.DeleteDish(c.Context(), id); err != nil {
		if errors.Is(err, infrastructure.ErrDishNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "dish not found"})
		}

		mh.log.Error("failed to delete dish", err, "dishID", id)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "dish successfully deleted"})
}
