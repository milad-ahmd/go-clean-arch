package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	"github.com/milad-ahmd/go-clean-arch/pkg/errors"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"go.uber.org/zap"
)

type categoryRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *sql.DB, logger logger.Logger) domain.CategoryRepository {
	return &categoryRepository{
		db:     db,
		logger: logger,
	}
}

// FindByID finds a category by ID
func (r *categoryRepository) FindByID(ctx context.Context, id int64) (*domain.Category, error) {
	query := `
		SELECT id, name, description, slug, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	var category domain.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.Slug,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Category", id)
		}
		r.logger.Error("Failed to find category by ID", zap.Int64("id", id), zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	return &category, nil
}

// FindAll finds all categories with pagination
func (r *categoryRepository) FindAll(ctx context.Context, limit, offset int) ([]domain.Category, int, error) {
	query := `
		SELECT id, name, description, slug, created_at, updated_at
		FROM categories
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("Failed to find all categories", zap.Error(err))
		return nil, 0, errors.NewInternalError(err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var category domain.Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.Slug,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan category", zap.Error(err))
			return nil, 0, errors.NewInternalError(err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating category rows", zap.Error(err))
		return nil, 0, errors.NewInternalError(err)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM categories`
	err = r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to get total category count", zap.Error(err))
		return nil, 0, errors.NewInternalError(err)
	}

	return categories, total, nil
}

// Create creates a new category
func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	query := `
		INSERT INTO categories (name, description, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	now := time.Now().Unix()
	category.CreatedAt = now
	category.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx,
		query,
		category.Name,
		category.Description,
		category.Slug,
		category.CreatedAt,
		category.UpdatedAt,
	).Scan(&category.ID)

	if err != nil {
		r.logger.Error("Failed to create category", zap.Error(err))
		return errors.NewInternalError(err)
	}

	return nil
}

// Update updates a category
func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	query := `
		UPDATE categories
		SET name = $1, description = $2, slug = $3, updated_at = $4
		WHERE id = $5
	`

	category.UpdatedAt = time.Now().Unix()

	result, err := r.db.ExecContext(
		ctx,
		query,
		category.Name,
		category.Description,
		category.Slug,
		category.UpdatedAt,
		category.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update category", zap.Int64("id", category.ID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewInternalError(err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("Category", category.ID)
	}

	return nil
}

// Delete deletes a category
func (r *categoryRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM categories WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete category", zap.Int64("id", id), zap.Error(err))
		return errors.NewInternalError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewInternalError(err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("Category", id)
	}

	return nil
}

// FindBySlug finds a category by slug
func (r *categoryRepository) FindBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	query := `
		SELECT id, name, description, slug, created_at, updated_at
		FROM categories
		WHERE slug = $1
	`

	var category domain.Category
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.Slug,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Category", fmt.Sprintf("slug=%s", slug))
		}
		r.logger.Error("Failed to find category by slug", zap.String("slug", slug), zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	return &category, nil
}

// FindByName finds a category by name
func (r *categoryRepository) FindByName(ctx context.Context, name string) (*domain.Category, error) {
	query := `
		SELECT id, name, description, slug, created_at, updated_at
		FROM categories
		WHERE name = $1
	`

	var category domain.Category
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.Slug,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Category", fmt.Sprintf("name=%s", name))
		}
		r.logger.Error("Failed to find category by name", zap.String("name", name), zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	return &category, nil
}
