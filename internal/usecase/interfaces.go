package usecase

import (
	"context"

	"github.com/ssklv/mixfood-menu-service/internal/domain"
)

type MenuRepository interface {
	CreateDish(ctx context.Context, dish *domain.Dish) error
	GetDishByID(ctx context.Context, id int64) (*domain.Dish, error)
	GetAllDishes(ctx context.Context) ([]*domain.Dish, error)
	UpdateDish(ctx context.Context, input *domain.UpdateDishParams) (*domain.Dish, error)
	DeleteDish(ctx context.Context, id int64) error
}

type MenuUsecase interface {
	CreateDish(ctx context.Context, dish *domain.Dish) error
	GetDishByID(ctx context.Context, id int64) (*domain.Dish, error)
	GetAllDishes(ctx context.Context) ([]*domain.Dish, error)
	UpdateDish(ctx context.Context, input *domain.UpdateDishParams) (*domain.Dish, error)
	DeleteDish(ctx context.Context, id int64) error
}

type TokenProvider interface {
	ParseToken(tokenStr string) (int64, string, error)
}
