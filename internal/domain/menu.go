package domain

import (
	"time"
)

type Dish struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Weight      int       `json:"weight"`
	Proteins    float64   `json:"proteins"`
	Fats        float64   `json:"fats"`
	Carbs       float64   `json:"carbs"`
	Calories    int       `json:"calories"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
