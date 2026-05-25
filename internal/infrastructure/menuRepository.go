package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssklv/mixfood-menu-service/internal/domain"
)

var dishCols = []string{
	"id",
	"category_id",
	"name",
	"description",
	"price",
	"weight",
	"volume",
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
	return &menuRepository{db: db, psql: psql}
}

func (r *menuRepository) CreateDish(ctx context.Context, d *domain.Dish) error {
	sql, args, err := r.psql.
		Insert("dishes").
		Columns(
			"category_id", "name", "description", "price",
			"weight", "volume", "proteins", "fats", "carbs", "calories",
			"image_url",
		).
		Values(
			d.CategoryID, d.Name, d.Description, d.Price,
			d.Weight, d.Volume, d.Proteins, d.Fats, d.Carbs, d.Calories,
			d.ImageURL,
		).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()

	if err != nil {
		return err
	}
	return r.db.QueryRow(ctx, sql, args...).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
}

func (r *menuRepository) UpdateDish(ctx context.Context, input *domain.UpdateDishParams) (*domain.Dish, error) {
	builder := r.psql.Update("dishes")

	if input.CategoryID != nil {
		builder = builder.Set("category_id", *input.CategoryID)
	}
	if input.Name != nil {
		builder = builder.Set("name", *input.Name)
	}
	if input.Description != nil {
		builder = builder.Set("description", *input.Description)
	}
	if input.Price != nil {
		builder = builder.Set("price", *input.Price)
	}
	if input.Weight != nil {
		builder = builder.Set("weight", input.Weight)
	}
	if input.Volume != nil {
		builder = builder.Set("volume", input.Volume)
	}
	if input.Proteins != nil {
		builder = builder.Set("proteins", *input.Proteins)
	}
	if input.Fats != nil {
		builder = builder.Set("fats", *input.Fats)
	}
	if input.Carbs != nil {
		builder = builder.Set("carbs", *input.Carbs)
	}
	if input.Calories != nil {
		builder = builder.Set("calories", *input.Calories)
	}
	if input.ImageURL != nil {
		builder = builder.Set("image_url", *input.ImageURL)
	}

	builder = builder.Set("updated_at", sq.Expr("NOW()"))

	sql, args, err := builder.
		Where(sq.Eq{"id": input.ID}).
		Suffix("RETURNING " + strings.Join(dishCols, ", ")).
		ToSql()

	if err != nil {
		return nil, err
	}

	dish := &domain.Dish{}
	err = scanDish(r.db.QueryRow(ctx, sql, args...), dish)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrDishNotFound
	}
	return dish, err
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

func (r *menuRepository) GetDishByID(ctx context.Context, id int64) (*domain.Dish, error) {
	sql, args, err := r.psql.
		Select(dishCols...).
		From("dishes").
		Where(sq.Eq{"id": id}).
		ToSql()

	if err != nil {
		return nil, err
	}

	dish := &domain.Dish{}
	if err := scanDish(r.db.QueryRow(ctx, sql, args...), dish); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDishNotFound
		}
		return nil, err
	}
	return dish, nil
}

func (r *menuRepository) GetAllDishes(ctx context.Context) ([]*domain.Dish, error) {
	sql, args, err := r.psql.
		Select(dishCols...).
		From("dishes").
		OrderBy("id DESC").
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dishes []*domain.Dish
	for rows.Next() {
		dish := &domain.Dish{}
		if err := scanDish(rows, dish); err != nil {
			return nil, err
		}
		dishes = append(dishes, dish)
	}
	return dishes, nil
}

func scanDish(row pgx.Row, d *domain.Dish) error {
	return row.Scan(
		&d.ID,
		&d.CategoryID,
		&d.Name,
		&d.Description,
		&d.Price,
		&d.Weight,
		&d.Volume,
		&d.Proteins,
		&d.Fats,
		&d.Carbs,
		&d.Calories,
		&d.ImageURL,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
}
