package usecase

import (
	"strings"
	"unicode/utf8"
)

const (
	maxDishNameLen = 100
)

func validateDishName(name string) error {
	count := utf8.RuneCountInString(strings.TrimSpace(name))
	if count == 0 || count > maxDishNameLen {
		return ErrInvalidName
	}
	return nil
}

func validateDishCategory(category string) error {
	cleanCategory := strings.ToLower(strings.TrimSpace(category))

	switch cleanCategory {
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

func validateDishBJU(value float64) error {
	if value < 0 {
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
	if trimmed != "" && !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
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
