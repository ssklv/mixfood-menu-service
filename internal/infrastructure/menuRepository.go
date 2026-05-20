package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	//"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssklv/mixfood-menu-service/internal/domain"
)

var dishCols = []string{
	"id",
	"name",
	"category",
	"description",
	"price",
	"weight",
	"proteins",
	"fats",
	"carbs",
	"calories",
	"image_url",
	"created_at",
	"updated_at",
}

type menuRepository struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewMenuRepository(db *pgxpool.Pool, psql sq.StatementBuilderType) *menuRepository {
	return &menuRepository{
		db:   db,
		psql: psql,
	}
}

func (r *menuRepository) CreateDish(ctx context.Context, dish *domain.Dish) error {
	sql, args, err := r.psql.
		Insert("dishes").
		Columns(dishCols[1:]...).
		Values(
			dish.Name,
			dish.Category,
			dish.Description,
			dish.Price,
			dish.Weight,
			dish.Proteins,
			dish.Fats,
			dish.Carbs,
			dish.Calories,
			dish.ImageURL,
			dish.CreatedAt,
			dish.UpdatedAt,
		).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return fmt.Errorf("build create dish query: %w", err)
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(&dish.ID)
	if err != nil {
		return fmt.Errorf("execute create dish: %w", err)
	}
	return nil
}

func (r *menuRepository) GetDishByID(ctx context.Context, id int64) (*domain.Dish, error) {
	sql, args, err := r.psql.
		Select(dishCols...).
		From("dishes").
		Where(sq.Eq{"id": id}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("get dish by id query: %w", err)
	}

	dishRes := &domain.Dish{}
	row := r.db.QueryRow(ctx, sql, args...)

	if err := scanDish(row, dishRes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDishNotFound
		}
		return nil, fmt.Errorf("get dish by id: %w", err)
	}
	return dishRes, nil
}

func (r *menuRepository) GetAllDishes(ctx context.Context) ([]*domain.Dish, error) {
	sql, args, err := r.psql.
		Select(dishCols...).
		From("dishes").
		OrderBy("id DESC").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("get all dishes query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("get all dishes: %w", err)
	}
	defer rows.Close()

	var dishes []*domain.Dish
	for rows.Next() {
		dish := &domain.Dish{}
		if err := scanDish(rows, dish); err != nil {
			return nil, fmt.Errorf("scan dish from list: %w", err)
		}
		dishes = append(dishes, dish)
	}
	return dishes, nil
}

func (r *menuRepository) UpdateDish(ctx context.Context, input *domain.UpdateDishParams) (*domain.Dish, error) {
	setCount := 0
	builder := r.psql.Update("dishes")

	if input.Name != nil {
		builder = builder.Set("name", *input.Name)
		setCount++
	}
	if input.Category != nil {
		builder = builder.Set("category", *input.Category)
		setCount++
	}
	if input.Description != nil {
		builder = builder.Set("description", *input.Description)
		setCount++
	}
	if input.Price != nil {
		builder = builder.Set("price", *input.Price)
		setCount++
	}
	if input.Weight != nil {
		builder = builder.Set("weight", *input.Weight)
		setCount++
	}
	if input.Proteins != nil {
		builder = builder.Set("proteins", *input.Proteins)
		setCount++
	}
	if input.Fats != nil {
		builder = builder.Set("fats", *input.Fats)
		setCount++
	}
	if input.Carbs != nil {
		builder = builder.Set("carbs", *input.Carbs)
		setCount++
	}
	if input.Calories != nil {
		builder = builder.Set("calories", *input.Calories)
		setCount++
	}
	if input.ImageURL != nil {
		builder = builder.Set("image_url", *input.ImageURL)
		setCount++
	}

	if setCount == 0 {
		return nil, ErrNoChanges
	}

	builder = builder.Set("updated_at", sq.Expr("NOW()"))

	sql, args, err := builder.
		Where(sq.Eq{"id": input.ID}).
		Suffix("RETURNING " + strings.Join(dishCols, ", ")).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build update dish query: %w", err)
	}

	dish := &domain.Dish{}
	row := r.db.QueryRow(ctx, sql, args...)

	if err := scanDish(row, dish); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDishNotFound
		}
		return nil, fmt.Errorf("update dish execute: %w", err)
	}
	return dish, nil
}

func (r *menuRepository) DeleteDish(ctx context.Context, id int64) error {
	sql, args, err := r.psql.
		Delete("dishes").
		Where(sq.Eq{"id": id}).
		ToSql()

	if err != nil {
		return fmt.Errorf("build delete query: %w", err)
	}

	res, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("delete dish: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrDishNotFound
	}

	return nil
}

func scanDish(row pgx.Row, dish *domain.Dish) error {
	return row.Scan(
		&dish.ID,
		&dish.Name,
		&dish.Category,
		&dish.Description,
		&dish.Price,
		&dish.Weight,
		&dish.Proteins,
		&dish.Fats,
		&dish.Carbs,
		&dish.Calories,
		&dish.ImageURL,
		&dish.CreatedAt,
		&dish.UpdatedAt,
	)
}
