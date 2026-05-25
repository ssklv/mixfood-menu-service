package domain

import (
	"time"
)

type Category struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type Dish struct {
	ID          int64     `json:"id" db:"id"`
	CategoryID  int64     `json:"categoryId" db:"category_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	Weight      *int      `json:"weight,omitempty" db:"weight"`
	Volume      *float64  `json:"volume,omitempty" db:"volume"`
	Proteins    float64   `json:"proteins" db:"proteins"`
	Fats        float64   `json:"fats" db:"fats"`
	Carbs       float64   `json:"carbs" db:"carbs"`
	Calories    int       `json:"calories" db:"calories"`
	ImageURL    string    `json:"imageUrl" db:"image_url"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}
