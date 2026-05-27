package usecase

import (
	"strings"
	"unicode/utf8"

	"github.com/ssklv/mixfood-menu-service/internal/domain"
)

const (
	maxDishNameLen = 100
)

func validateDish(d *domain.Dish) error {
	if err := validateDishName(d.Name); err != nil {
		return err
	}
	if d.CategoryID <= 0 {
		return ErrInvalidCategory
	}
	if err := validateDishPrice(d.Price); err != nil {
		return err
	}
	if err := validateDishBJU(d.Proteins, d.Fats, d.Carbs); err != nil {
		return err
	}
	if err := validateDishCalories(d.Calories); err != nil {
		return err
	}
	if err := validateDishImageURL(d.ImageURL); err != nil {
		return err
	}
	if err := validateMeasurements(d.Weight, d.Volume); err != nil {
		return err
	}
	return nil
}

func validateUpdateDish(input *domain.UpdateDishParams) error {
	if input.Name != nil {
		if err := validateDishName(*input.Name); err != nil {
			return err
		}
	}
	if input.CategoryID != nil && *input.CategoryID <= 0 {
		return ErrInvalidCategory
	}
	if input.Price != nil {
		if err := validateDishPrice(*input.Price); err != nil {
			return err
		}
	}
	if input.Proteins != nil || input.Fats != nil || input.Carbs != nil {
		p, f, c := 0.0, 0.0, 0.0
		if input.Proteins != nil {
			p = *input.Proteins
		}
		if input.Fats != nil {
			f = *input.Fats
		}
		if input.Carbs != nil {
			c = *input.Carbs
		}
		if err := validateDishBJU(p, f, c); err != nil {
			return err
		}
	}
	if input.Calories != nil {
		if err := validateDishCalories(*input.Calories); err != nil {
			return err
		}
	}
	if input.ImageURL != nil {
		if err := validateDishImageURL(*input.ImageURL); err != nil {
			return err
		}
	}
	if input.Weight != nil || input.Volume != nil {
		if input.Weight != nil && *input.Weight <= 0 {
			return ErrInvalidWeight
		}
		if input.Volume != nil && *input.Volume <= 0 {
			return ErrInvalidVolume
		}
		if input.Weight != nil && input.Volume != nil {
			return ErrMeasurementMismatch
		}
	}
	return nil
}

func validateDishName(name string) error {
	count := utf8.RuneCountInString(strings.TrimSpace(name))
	if count == 0 || count > maxDishNameLen {
		return ErrInvalidName
	}
	return nil
}

func validateDishPrice(price float64) error {
	if price <= 0 {
		return ErrInvalidPrice
	}
	return nil
}

func validateDishBJU(p, f, c float64) error {
	if p < 0 || f < 0 || c < 0 {
		return ErrInvalidBJU
	}
	return nil
}

func validateDishCalories(calories int) error {
	if calories < 0 {
		return ErrInvalidBJU
	}
	return nil
}

func validateDishImageURL(url string) error {
	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		return nil
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") || strings.HasPrefix(trimmed, "/uploads/") {
		return nil
	}
	return ErrInvalidImageURL
}

func validateMeasurements(weight *int, volume *float64) error {
	if weight == nil && volume == nil {
		return ErrInvalidWeight
	}
	if weight != nil && volume != nil {
		return ErrMeasurementMismatch
	}
	if weight != nil && *weight <= 0 {
		return ErrInvalidWeight
	}
	if volume != nil && *volume <= 0 {
		return ErrInvalidVolume
	}
	return nil
}
