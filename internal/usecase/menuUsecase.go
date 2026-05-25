package usecase

import (
	"context"

	"github.com/ssklv/mixfood-menu-service/internal/domain"
)

type menuUsecase struct {
	repository MenuRepository
}

func NewMenuUsecase(rep MenuRepository) MenuUsecase {
	return &menuUsecase{repository: rep}
}

func (mu *menuUsecase) CreateDish(ctx context.Context, d *domain.Dish) error {
	if err := validateDish(d); err != nil {
		return err
	}
	return mu.repository.CreateDish(ctx, d)
}

func (mu *menuUsecase) UpdateDish(ctx context.Context, input *domain.UpdateDishParams) (*domain.Dish, error) {
	return mu.repository.UpdateDish(ctx, input)
}

func (mu *menuUsecase) GetDishByID(ctx context.Context, id int64) (*domain.Dish, error) {
	return mu.repository.GetDishByID(ctx, id)
}

func (mu *menuUsecase) GetAllDishes(ctx context.Context) ([]*domain.Dish, error) {
	return mu.repository.GetAllDishes(ctx)
}

func (mu *menuUsecase) DeleteDish(ctx context.Context, id int64) error {
	return mu.repository.DeleteDish(ctx, id)
}
