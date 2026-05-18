package domain

type UpdateDishParams struct {
	ID          int64
	Name        *string
	Category    *string
	Description *string
	Price       *float64
	Weight      **int
	Volume      **float64
	Proteins    *float64
	Fats        *float64
	Carbs       *float64
	Calories    *int
	ImageURL    *string
}
