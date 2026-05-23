package handlers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-menu-service/internal/domain"
	"github.com/ssklv/mixfood-menu-service/internal/infrastructure"
	"github.com/ssklv/mixfood-menu-service/internal/usecase"

	//сваггер
	//"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/contrib/v3/swaggerui"
	_ "github.com/ssklv/mixfood-menu-service/docs"
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
	app.Use(swaggerui.New(swaggerui.Config{
		BasePath: "/",
		FilePath: "./docs/swagger.json",
		Path:     "swagger",
		Title:    "Mixfood Menu API Docs",
	}))

	menu := app.Group("/api/menu")

	menu.Get("/", mh.getAllDishes)
	menu.Get("/:id", mh.getDishByID)

	menu.Post("/", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.createDish)
	menu.Patch("/:id", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.updateDish)
	menu.Delete("/:id", mh.AuthMiddleware(), mh.RequireRole("admin"), mh.deleteDish)
}

// createDish godoc
// @Summary      Create a new dish
// @Description  Add a new dish or drink to the menu. Accessible only by users with the 'admin' role.
// @Tags         menu
// @Accept       json
// @Produce      json
// @Param        request  body      createDishReq  true  "New dish data"
// @Success      201      {object}  domain.Dish
// @Failure      400      {object}  map[string]string "Invalid request body or validation error"
// @Failure      401      {object}  map[string]string "Unauthorized (missing or invalid token)"
// @Failure      403      {object}  map[string]string "Forbidden (insufficient permissions)"
// @Failure      500      {object}  map[string]string "Internal server error"
// @Router       /menu [post]
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

// getAllDishes godoc
// @Summary      get all the dishes on the menu
// @Description  returns the full list of available food and drinks. A public endpoint
// @Tags         menu
// @Produce      json
// @Success      200  {array}   domain.Dish
// @Failure      500  {object}  map[string]string "internal server error"
// @Router       /menu [get]
func (mh *menuHandler) getAllDishes(c fiber.Ctx) error {
	dishes, err := mh.usecase.GetAllDishes(c.Context())
	if err != nil {
		mh.log.Error("failed to get all dishes", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(dishes)
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

// getDishByID godoc
// @Summary      Get dish by ID
// @Description  Get detailed information about a specific dish or drink by its unique ID. Public endpoint.
// @Tags         menu
// @Produce      json
// @Param        id   path      int  true  "Dish ID"
// @Success      200  {object}  domain.Dish
// @Failure      400  {object}  map[string]string "Invalid dish ID format"
// @Failure      404  {object}  map[string]string "Dish not found"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /menu/{id} [get]
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
