package usecase

import (
	"context"
	"testing"

	"github.com/ssklv/mixfood-menu-service/internal/domain"
	"github.com/ssklv/mixfood-menu-service/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateDish_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewMenuRepository(t)
	uc := &menuUsecase{repository: mockRepo}

	weight := 500
	dish := &domain.Dish{
		Name:       "Pizza",
		CategoryID: 1,
		Price:      100.0,
		Weight:     &weight,
		Proteins:   10,
		Fats:       10,
		Carbs:      10,
		Calories:   200,
	}

	mockRepo.On("CreateDish", ctx, dish).Return(nil)

	err := uc.CreateDish(ctx, dish)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateDish_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewMenuRepository(t)
	uc := &menuUsecase{repository: mockRepo}

	// Хелперы для указателей
	weight := 100
	volume := 500.0

	tests := []struct {
		name        string
		dish        *domain.Dish
		expectedErr error
	}{
		{
			name:        "Invalid Name",
			dish:        &domain.Dish{Name: ""},
			expectedErr: ErrInvalidName,
		},
		{
			name:        "Invalid Category",
			dish:        &domain.Dish{Name: "Pizza", CategoryID: 0},
			expectedErr: ErrInvalidCategory,
		},
		{
			name:        "Invalid Price",
			dish:        &domain.Dish{Name: "Pizza", CategoryID: 1, Price: -1},
			expectedErr: ErrInvalidPrice,
		},
		{
			name: "Measurement Mismatch (Category 6 has weight)",
			dish: &domain.Dish{
				Name:       "Cola",
				CategoryID: 6, // Напиток
				Price:      100,
				Volume:     &volume, // Объем есть (проходим проверку объема)
				Weight:     &weight, // Но мы добавили вес, что запрещено
			},
			expectedErr: ErrMeasurementMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.CreateDish(ctx, tt.dish)
			assert.ErrorIs(t, err, tt.expectedErr)
			mockRepo.AssertNotCalled(t, "CreateDish", mock.Anything, mock.Anything)
		})
	}
}

func TestGetDishByID_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewMenuRepository(t)
	uc := &menuUsecase{repository: mockRepo}

	expectedDish := &domain.Dish{ID: 1, Name: "Pizza"}
	mockRepo.On("GetDishByID", ctx, int64(1)).Return(expectedDish, nil)

	dish, err := uc.GetDishByID(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedDish, dish)
}
