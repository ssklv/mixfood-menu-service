package usecase

import (
	"errors"
	"testing"

	"github.com/ssklv/mixfood-menu-service/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestValidateDish(t *testing.T) {
	w := 500
	v := 250.0

	tests := []struct {
		name    string
		dish    *domain.Dish
		wantErr error
	}{
		{
			"Valid pizza",
			&domain.Dish{Name: "Pizza", Category: "пицца", Price: 100, Proteins: 10, Fats: 10, Carbs: 10, Calories: 100, Weight: &w},
			nil,
		},
		{
			"Invalid category",
			&domain.Dish{Name: "Dish", Category: "суши", Price: 100},
			ErrInvalidCategory,
		},
		{
			"Drink with volume",
			&domain.Dish{Name: "Cola", Category: "напитки", Price: 100, Volume: &v},
			nil,
		},
		{
			"Drink error (weight instead of volume)",
			&domain.Dish{Name: "Cola", Category: "напитки", Price: 100, Weight: &w, Volume: &v},
			ErrMeasurementMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDish(tt.dish)
			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
