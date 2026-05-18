package usecase

import (
	"context"
	"time"

	"github.com/ssklv/mixfood-menu-service/internal/domain"
)

type menuUsecase struct {
	repository MenuRepository
}

func NewMenuUsecase(rep MenuRepository) MenuUsecase {
	return &menuUsecase{
		repository: rep,
	}
}

func (mu *menuUsecase) CreateDish(ctx context.Context, dish *domain.Dish) error {
	if err := validateDishName(dish.Name); err != nil {
		return err
	}
	if err := validateDishCategory(dish.Category); err != nil {
		return err
	}
	if err := validateDishPrice(dish.Price); err != nil {
		return err
	}
	if err := validateMeasurements(dish.Category, dish.Weight, dish.Volume); err != nil {
		return err
	}
	if err := validateDishBJU(dish.Proteins); err != nil {
		return err
	}
	if err := validateDishBJU(dish.Fats); err != nil {
		return err
	}
	if err := validateDishBJU(dish.Carbs); err != nil {
		return err
	}
	if err := validateDishCalories(dish.Calories); err != nil {
		return err
	}
	if err := validateDishImageURL(dish.ImageURL); err != nil {
		return err
	}

	dish.CreatedAt = time.Now()
	dish.UpdatedAt = time.Now()

	return mu.repository.CreateDish(ctx, dish)
}

func (mu *menuUsecase) UpdateDish(ctx context.Context, input *domain.UpdateDishParams) (*domain.Dish, error) {
	current, err := mu.repository.GetDishByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	finalCategory := current.Category
	if input.Category != nil {
		finalCategory = *input.Category
		if err := validateDishCategory(finalCategory); err != nil {
			return nil, err
		}
	}

	finalWeight := current.Weight
	if input.Weight != nil {
		finalWeight = *input.Weight
	}

	finalVolume := current.Volume
	if input.Volume != nil {
		finalVolume = *input.Volume
	}

	if err := validateMeasurements(finalCategory, finalWeight, finalVolume); err != nil {
		return nil, err
	}

	if input.Name != nil {
		if err := validateDishName(*input.Name); err != nil {
			return nil, err
		}
	}
	if input.Price != nil {
		if err := validateDishPrice(*input.Price); err != nil {
			return nil, err
		}
	}
	if input.Proteins != nil {
		if err := validateDishBJU(*input.Proteins); err != nil {
			return nil, err
		}
	}
	if input.Fats != nil {
		if err := validateDishBJU(*input.Fats); err != nil {
			return nil, err
		}
	}
	if input.Carbs != nil {
		if err := validateDishBJU(*input.Carbs); err != nil {
			return nil, err
		}
	}
	if input.Calories != nil {
		if err := validateDishCalories(*input.Calories); err != nil {
			return nil, err
		}
	}
	if input.ImageURL != nil {
		if err := validateDishImageURL(*input.ImageURL); err != nil {
			return nil, err
		}
	}

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
