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
	if err := validateDishCategory(d.Category); err != nil {
		return err
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
	if err := validateMeasurements(d.Category, d.Weight, d.Volume); err != nil {
		return err
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

func validateDishCategory(category string) error {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "пицца", "бургеры", "закуски", "салаты", "десерты", "напитки":
		return nil
	default:
		return ErrInvalidCategory
	}
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
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		return ErrInvalidImageURL
	}
	return nil
}

func validateMeasurements(category string, weight *int, volume *float64) error {
	cleanCategory := strings.ToLower(strings.TrimSpace(category))
	if cleanCategory == "напитки" {
		if volume == nil || *volume <= 0 {
			return ErrInvalidVolume
		}
		if weight != nil {
			return ErrMeasurementMismatch
		}
	} else {
		if weight == nil || *weight <= 0 {
			return ErrInvalidWeight
		}
		if volume != nil {
			return ErrMeasurementMismatch
		}
	}
	return nil
}
