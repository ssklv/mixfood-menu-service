package domain

type UpdateCategoryParams struct {
	ID   int64   `json:"-"`
	Name *string `json:"name,omitempty"`
	Slug *string `json:"slug,omitempty"`
}

type UpdateDishParams struct {
	ID          int64    `json:"-"`
	CategoryID  *int64   `json:"categoryId,omitempty"`
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Weight      *int     `json:"weight,omitempty"`
	Volume      *float64 `json:"volume,omitempty"`
	Proteins    *float64 `json:"proteins,omitempty"`
	Fats        *float64 `json:"fats,omitempty"`
	Carbs       *float64 `json:"carbs,omitempty"`
	Calories    *int     `json:"calories,omitempty"`
	ImageURL    *string  `json:"imageUrl,omitempty"`
}
