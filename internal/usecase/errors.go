package usecase

import (
	"errors"
)

var (
	ErrInvalidName         = errors.New("dish name is too short or exceeds maximum length")
	ErrInvalidCategory     = errors.New("invalid category id")
	ErrInvalidPrice        = errors.New("dish price must be greater than zero")
	ErrInvalidWeight       = errors.New("dish weight must be greater than zero")
	ErrInvalidVolume       = errors.New("dish volume must be greater than zero")
	ErrInvalidBJU          = errors.New("proteins, fats, carbs or calories cannot be negative")
	ErrInvalidImageURL     = errors.New("invalid image url format")
	ErrMeasurementMismatch = errors.New("dish must have either weight for food or volume for drinks, not both")

	ErrDishAlreadyExists = errors.New("dish with this name already exists in category")
)
