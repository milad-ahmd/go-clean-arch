package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	"github.com/milad-ahmd/go-clean-arch/pkg/errors"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"go.uber.org/zap"
)

type productRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *sql.DB, logger logger.Logger) domain.ProductRepository {
	return &productRepository{
		db:     db,
		logger: logger,
	}
}

// FindByID finds a product by ID
func (r *productRepository) FindByID(ctx context.Context, id int64) (*domain.Product, error) {
	query := `
		SELECT p.id, p.name, p.description, p.price, p.sku, p.stock, p.category_id, p.images, p.created_at, p.updated_at,
			   c.id, c.name, c.description, c.slug, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
	`

	var product domain.Product
	var category domain.Category
	var imagesJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.SKU,
		&product.Stock,
		&product.CategoryID,
		&imagesJSON,
		&product.CreatedAt,
		&product.UpdatedAt,
		&category.ID,
		&category.Name,
		&category.Description,
		&category.Slug,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Product", id)
		}
		r.logger.Error("Failed to find product by ID", zap.Int64("id", id), zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Parse images JSON
	if imagesJSON != nil {
		if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
			r.logger.Error("Failed to unmarshal product images", zap.Error(err))
			return nil, errors.NewInternalError(err)
		}
	}

	product.Category = category

	return &product, nil
}

// FindAll finds all products with pagination
func (r *productRepository) FindAll(ctx context.Context, limit, offset int) ([]domain.Product, int, error) {
	query := `
		SELECT p.id, p.name, p.description, p.price, p.sku, p.stock, p.category_id, p.images, p.created_at, p.updated_at,
			   c.id, c.name, c.description, c.slug, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		ORDER BY p.id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("Failed to find all products", zap.Error(err))
		return nil, 0, errors.NewInternalError(err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		var category domain.Category
		var imagesJSON []byte

		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.SKU,
			&product.Stock,
			&product.CategoryID,
			&imagesJSON,
			&product.CreatedAt,
			&product.UpdatedAt,
			&category.ID,
			&category.Name,
			&category.Description,
			&category.Slug,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan product", zap.Error(err))
			return nil, 0, errors.NewInternalError(err)
		}

		// Parse images JSON
		if imagesJSON != nil {
			if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
				r.logger.Error("Failed to unmarshal product images", zap.Error(err))
				return nil, 0, errors.NewInternalError(err)
			}
		}

		product.Category = category
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating product rows", zap.Error(err))
		return nil, 0, errors.NewInternalError(err)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM products`
	err = r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to get total product count", zap.Error(err))
		return nil, 0, errors.NewInternalError(err)
	}

	return products, total, nil
}

// Create creates a new product
func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (name, description, price, sku, stock, category_id, images, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	now := time.Now().Unix()
	product.CreatedAt = now
	product.UpdatedAt = now

	// Convert images to JSON
	imagesJSON, err := json.Marshal(product.Images)
	if err != nil {
		r.logger.Error("Failed to marshal product images", zap.Error(err))
		return errors.NewInternalError(err)
	}

	err = r.db.QueryRowContext(
		ctx,
		query,
		product.Name,
		product.Description,
		product.Price,
		product.SKU,
		product.Stock,
		product.CategoryID,
		imagesJSON,
		product.CreatedAt,
		product.UpdatedAt,
	).Scan(&product.ID)

	if err != nil {
		r.logger.Error("Failed to create product", zap.Error(err))
		return errors.NewInternalError(err)
	}

	return nil
}

// Update updates a product
func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, sku = $4, stock = $5, category_id = $6, images = $7, updated_at = $8
		WHERE id = $9
	`

	product.UpdatedAt = time.Now().Unix()

	// Convert images to JSON
	imagesJSON, err := json.Marshal(product.Images)
	if err != nil {
		r.logger.Error("Failed to marshal product images", zap.Error(err))
		return errors.NewInternalError(err)
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		product.Name,
		product.Description,
		product.Price,
		product.SKU,
		product.Stock,
		product.CategoryID,
		imagesJSON,
		product.UpdatedAt,
		product.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update product", zap.Int64("id", product.ID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewInternalError(err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("Product", product.ID)
	}

	return nil
}

// Delete deletes a product
func (r *productRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete product", zap.Int64("id", id), zap.Error(err))
		return errors.NewInternalError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewInternalError(err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("Product", id)
	}

	return nil
}

// FindBySKU finds a product by SKU
func (r *productRepository) FindBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	query := `
		SELECT p.id, p.name, p.description, p.price, p.sku, p.stock, p.category_id, p.images, p.created_at, p.updated_at,
			   c.id, c.name, c.description, c.slug, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.sku = $1
	`

	var product domain.Product
	var category domain.Category
	var imagesJSON []byte

	err := r.db.QueryRowContext(ctx, query, sku).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.SKU,
		&product.Stock,
		&product.CategoryID,
		&imagesJSON,
		&product.CreatedAt,
		&product.UpdatedAt,
		&category.ID,
		&category.Name,
		&category.Description,
		&category.Slug,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Product", fmt.Sprintf("sku=%s", sku))
		}
		r.logger.Error("Failed to find product by SKU", zap.String("sku", sku), zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Parse images JSON
	if imagesJSON != nil {
		if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
			r.logger.Error("Failed to unmarshal product images", zap.Error(err))
			return nil, errors.NewInternalError(err)
		}
	}

	product.Category = category

	return &product, nil
}

// FindByCategory finds products by category ID
func (r *productRepository) FindByCategory(ctx context.Context, categoryID int64, limit, offset int) ([]domain.Product, int, error) {
	query := `
		SELECT p.id, p.name, p.description, p.price, p.sku, p.stock, p.category_id, p.images, p.created_at, p.updated_at,
			   c.id, c.name, c.description, c.slug, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.category_id = $1
		ORDER BY p.id
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, categoryID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to find products by category", zap.Int64("categoryID", categoryID), zap.Error(err))
		return nil, 0, errors.NewInternalError(err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		var category domain.Category
		var imagesJSON []byte

		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.SKU,
			&product.Stock,
			&product.CategoryID,
			&imagesJSON,
			&product.CreatedAt,
			&product.UpdatedAt,
			&category.ID,
			&category.Name,
			&category.Description,
			&category.Slug,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan product", zap.Error(err))
			return nil, 0, errors.NewInternalError(err)
		}

		// Parse images JSON
		if imagesJSON != nil {
			if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
				r.logger.Error("Failed to unmarshal product images", zap.Error(err))
				return nil, 0, errors.NewInternalError(err)
			}
		}

		product.Category = category
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating product rows", zap.Error(err))
		return nil, 0, errors.NewInternalError(err)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM products WHERE category_id = $1`
	err = r.db.QueryRowContext(ctx, countQuery, categoryID).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to get total product count by category", zap.Int64("categoryID", categoryID), zap.Error(err))
		return nil, 0, errors.NewInternalError(err)
	}

	return products, total, nil
}

// UpdateStock updates a product's stock
func (r *productRepository) UpdateStock(ctx context.Context, id int64, quantity int) error {
	query := `
		UPDATE products
		SET stock = stock + $1, updated_at = $2
		WHERE id = $3
		RETURNING stock
	`

	now := time.Now().Unix()
	var newStock int

	err := r.db.QueryRowContext(ctx, query, quantity, now, id).Scan(&newStock)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NewNotFoundError("Product", id)
		}
		r.logger.Error("Failed to update product stock", zap.Int64("id", id), zap.Int("quantity", quantity), zap.Error(err))
		return errors.NewInternalError(err)
	}

	// Check if stock is negative
	if newStock < 0 {
		// Revert the change
		_, err := r.db.ExecContext(ctx, `UPDATE products SET stock = stock - $1 WHERE id = $2`, quantity, id)
		if err != nil {
			r.logger.Error("Failed to revert stock update", zap.Int64("id", id), zap.Error(err))
		}
		return errors.NewBadRequestError("Insufficient stock")
	}

	return nil
}

// SearchProducts searches for products by name or description
func (r *productRepository) SearchProducts(ctx context.Context, query string, limit, offset int) ([]domain.Product, int, error) {
	sqlQuery := `
		SELECT p.id, p.name, p.description, p.price, p.sku, p.stock, p.category_id, p.images, p.created_at, p.updated_at,
			   c.id, c.name, c.description, c.slug, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.name ILIKE $1 OR p.description ILIKE $1
		ORDER BY p.id
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sqlQuery, searchPattern, limit, offset)
	if err != nil {
		r.logger.Error("Failed to search products", zap.String("query", query), zap.Error(err))
		return nil, 0, errors.NewInternalError(err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		var category domain.Category
		var imagesJSON []byte

		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.SKU,
			&product.Stock,
			&product.CategoryID,
			&imagesJSON,
			&product.CreatedAt,
			&product.UpdatedAt,
			&category.ID,
			&category.Name,
			&category.Description,
			&category.Slug,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan product", zap.Error(err))
			return nil, 0, errors.NewInternalError(err)
		}

		// Parse images JSON
		if imagesJSON != nil {
			if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
				r.logger.Error("Failed to unmarshal product images", zap.Error(err))
				return nil, 0, errors.NewInternalError(err)
			}
		}

		product.Category = category
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating product rows", zap.Error(err))
		return nil, 0, errors.NewInternalError(err)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM products WHERE name ILIKE $1 OR description ILIKE $1`
	err = r.db.QueryRowContext(ctx, countQuery, searchPattern).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to get total product count for search", zap.String("query", query), zap.Error(err))
		return nil, 0, errors.NewInternalError(err)
	}

	return products, total, nil
}
