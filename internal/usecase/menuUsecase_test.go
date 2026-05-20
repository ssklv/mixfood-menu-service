package usecase

import (
	"context"
	"testing"

	"github.com/ssklv/mixfood-menu-service/internal/domain"
)

//func (mu *menuUsecase) CreateDish(ctx context.Context, dish *domain.Dish) error {
//func (mu *menuUsecase) UpdateDish(ctx context.Context, input *domain.UpdateDishParams) (*domain.Dish, error) {
//func (mu *menuUsecase) GetDishByID(ctx context.Context, id int64) (*domain.Dish, error) {
//func (mu *menuUsecase) GetAllDishes(ctx context.Context) ([]*domain.Dish, error) {
//func (mu *menuUsecase) DeleteDish(ctx context.Context, id int64) error {

//Dish struct
// 	ID          int64     `json:"id"`
// 	Name        string    `json:"name"`
// 	Category    string    `json:"category"`
// 	Description string    `json:"description"`
// 	Price       float64   `json:"price"`
// 	Weight      *int      `json:"weight,omitempty"`
// 	Volume      *float64  `json:"volume,omitempty"`
// 	Proteins    float64   `json:"proteins"`
// 	Fats        float64   `json:"fats"`
// 	Carbs       float64   `json:"carbs"`
// 	Calories    int       `json:"calories"`
// 	ImageURL    string    `json:"image_url"`
// 	CreatedAt   time.Time `json:"created_at"`
// 	UpdatedAt   time.Time `json:"updated_at"`

//принимает *testing.T начинается с префикса Test

type mockRepository struct {
	errToReturn  error
	dishToReturn *domain.Dish
	listToReturn []*domain.Dish
}

func (mr *mockRepository) CreateDish(ctx context.Context, dish *domain.Dish) error {
	return mr.errToReturn //
}

func (mr *mockRepository) GetDishByID(ctx context.Context, id int64) (*domain.Dish, error) {
	return mr.dishToReturn, mr.errToReturn
}

func (mr *mockRepository) GetAllDishes(ctx context.Context) ([]*domain.Dish, error) {
	return mr.listToReturn, mr.errToReturn
}

func (mr *mockRepository) DeleteDish(ctx context.Context, id int64) error {
	return mr.errToReturn
}

func (mr *mockRepository) UpdateDish(ctx context.Context, input *domain.UpdateDishParams) (*domain.Dish, error) {
	return mr.dishToReturn, mr.errToReturn
}

//запуск с покрытием go test -cover
//go test -cover ./internal/usecase/

func createValidDish() *domain.Dish {
	weight := 400
	return &domain.Dish{
		Name:     "Бургер с беконом",
		Category: "бургеры",
		Price:    500.0,
		Weight:   &weight,
		Proteins: 10.0,
		Fats:     10.0,
		Carbs:    10.0,
		Calories: 200,
		ImageURL: "http://example.com/image.jpg",
	}
}

func Test_CreateDishInvalidPrice(t *testing.T) {
	mockRepository := &mockRepository{}
	mu := NewMenuUsecase(mockRepository)

	dish := createValidDish()
	dish.Price = -600.0

	err := mu.CreateDish(context.Background(), dish)
	if err == nil {
		t.Errorf("")
	}
}

func Test_CreateDishValidPrice(t *testing.T) {
	mockRepository := &mockRepository{}
	mu := NewMenuUsecase(mockRepository)

	err := mu.CreateDish(context.Background(), createValidDish())
	if err != nil {
		t.Errorf("%v", err)
	}
}

func Test_CreateDishInvalidCategory(t *testing.T) {
	mockRepository := &mockRepository{}
	mu := NewMenuUsecase(mockRepository)

	dish := createValidDish()
	dish.Category = "суши"

	err := mu.CreateDish(context.Background(), dish)
	if err == nil {
		t.Errorf("")
	}
}

func Test_CreateDishValidCategory(t *testing.T) {
	mockRepository := &mockRepository{}
	mu := NewMenuUsecase(mockRepository)

	dish := createValidDish()
	dish.Category = "пицца"

	err := mu.CreateDish(context.Background(), dish)
	if err != nil {
		t.Errorf("%v", err)
	}
}

//

func Test_CreateDishInvalidName(t *testing.T) {
	mockRepository := &mockRepository{}
	mu := NewMenuUsecase(mockRepository)

	dish := createValidDish()
	dish.Name = ""

	err := mu.CreateDish(context.Background(), dish)
	if err == nil {
		t.Errorf("")
	}
}

func Test_CreateDishValidName(t *testing.T) {
	mockRepository := &mockRepository{}
	mu := NewMenuUsecase(mockRepository)

	dish := createValidDish()
	dish.Name = "оливковая"

	err := mu.CreateDish(context.Background(), dish)
	if err != nil {
		t.Errorf("%v", err)
	}
}

func Test_UpdateDishPartial(t *testing.T) {
	mockRepository := &mockRepository{
		dishToReturn: createValidDish(),
	}
	mu := NewMenuUsecase(mockRepository)

	newName := "новая вегетарианская"
	input := &domain.UpdateDishParams{
		ID:   1,
		Name: &newName,
	}

	_, err := mu.UpdateDish(context.Background(), input)
	if err != nil {
		t.Errorf("%v", err)
	}
}

func Test_CreateDishInvalidBJU(t *testing.T) {
	mockRepository := &mockRepository{}
	mu := NewMenuUsecase(mockRepository)

	dish := createValidDish()
	dish.Proteins = -10

	err := mu.CreateDish(context.Background(), dish)
	if err == nil {
		t.Errorf("")
	}
}

func Test_UpdateDish_Full(t *testing.T) {

	mockRepository := &mockRepository{
		dishToReturn: createValidDish(),
	}
	mu := NewMenuUsecase(mockRepository)
	newName := "Обновленное имя"
	newPrice := 999.0
	input := &domain.UpdateDishParams{
		ID:    1,
		Name:  &newName,
		Price: &newPrice,
	}

	_, err := mu.UpdateDish(context.Background(), input)
	if err != nil {
		t.Errorf("%v", err)
	}
}

func Test_ValidateMeasurements_Drinks(t *testing.T) {
	mockRepository := &mockRepository{}
	mu := NewMenuUsecase(mockRepository)

	dish := createValidDish()
	dish.Category = "напитки"
	dish.Weight = nil

	err := mu.CreateDish(context.Background(), dish)
	if err == nil {
		t.Errorf("")
	}
}
