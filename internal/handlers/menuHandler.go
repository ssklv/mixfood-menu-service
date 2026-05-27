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
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Access denied"})
		}
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Insufficient permissions"})
	}
}

func (mh *menuHandler) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	menu := router.Group("/menu")

	menu.Get("", mh.getAllDishes)
	menu.Get("/:id", mh.getDishByID)

	menu.Post("", authMiddleware, mh.RequireRole("admin"), mh.createDish)
	menu.Patch("/:id", authMiddleware, mh.RequireRole("admin"), mh.updateDish)
	menu.Delete("/:id", authMiddleware, mh.RequireRole("admin"), mh.deleteDish)
	menu.Post("/upload", authMiddleware, mh.RequireRole("admin"), mh.uploadImage)
}

// @Summary Create a new dish
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param dish body domain.Dish true "Dish data"
// @Success 201 {object} domain.Dish
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/menu [post]
func (mh *menuHandler) createDish(c fiber.Ctx) error {
	var dish domain.Dish
	if err := c.Bind().Body(&dish); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	if err := mh.usecase.CreateDish(c.Context(), &dish); err != nil {
		return mh.handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dish)
}

// @Summary Get all dishes
// @Produce json
// @Success 200 {array} domain.Dish
// @Failure 500 {object} ErrorResponse
// @Router /api/menu [get]
func (mh *menuHandler) getAllDishes(c fiber.Ctx) error {
	dishes, err := mh.usecase.GetAllDishes(c.Context())
	if err != nil {
		mh.log.Error("failed to get all dishes", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Internal server error"})
	}
	return c.JSON(dishes)
}

// @Summary Get dish by ID
// @Produce json
// @Param id path int true "Dish ID"
// @Success 200 {object} domain.Dish
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/menu/{id} [get]
func (mh *menuHandler) getDishByID(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid dish ID"})
	}

	dish, err := mh.usecase.GetDishByID(c.Context(), id)
	if err != nil {
		return mh.handleError(c, err)
	}
	return c.JSON(dish)
}

// @Summary Update dish parameters
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Dish ID"
// @Param params body domain.UpdateDishParams true "Fields to update"
// @Success 200 {object} domain.Dish
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/menu/{id} [patch]
func (mh *menuHandler) updateDish(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid dish ID"})
	}

	var params domain.UpdateDishParams
	if err := c.Bind().Body(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}
	params.ID = id

	updated, err := mh.usecase.UpdateDish(c.Context(), &params)
	if err != nil {
		return mh.handleError(c, err)
	}
	return c.JSON(updated)
}

// @Summary Delete dish from menu
// @Security BearerAuth
// @Param id path int true "Dish ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/menu/{id} [delete]
func (mh *menuHandler) deleteDish(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid dish ID"})
	}

	if err := mh.usecase.DeleteDish(c.Context(), id); err != nil {
		return mh.handleError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Dish successfully deleted from menu"})
}

// @Summary Upload dish image
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/menu/upload [post]
func (mh *menuHandler) uploadImage(c fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Image file is required"})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Failed to read uploaded file"})
	}
	defer src.Close()

	url, err := mh.fileStorage.SaveFile(src, file.Filename)
	if err != nil {
		mh.log.Error("failed to store uploaded image", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to save file on server"})
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
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Requested dish not found"})

	case errors.Is(err, usecase.ErrNoChanges):
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})

	default:
		mh.log.Error("unexpected internal error", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Internal server error"})
	}
}
