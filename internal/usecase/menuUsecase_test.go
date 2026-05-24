package usecase

import (
	"context"
	"testing"

	"github.com/ssklv/mixfood-menu-service/internal/domain"
	"github.com/ssklv/mixfood-menu-service/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMenuUsecase_CreateDish(t *testing.T) {
	repo := new(mocks.MenuRepository)
	uc := NewMenuUsecase(repo)
	w := 500

	t.Run("Success", func(t *testing.T) {
		dish := &domain.Dish{Name: "Пицца", Category: "пицца", Price: 100, Weight: &w}
		repo.On("CreateDish", mock.Anything, mock.Anything).Return(nil).Once()

		err := uc.CreateDish(context.Background(), dish)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func TestMenuUsecase_GetDishByID(t *testing.T) {
	repo := new(mocks.MenuRepository)
	uc := NewMenuUsecase(repo)

	t.Run("Found", func(t *testing.T) {
		repo.On("GetDishByID", mock.Anything, int64(1)).Return(&domain.Dish{ID: 1}, nil).Once()

		dish, err := uc.GetDishByID(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), dish.ID)
		repo.AssertExpectations(t)
	})
}
