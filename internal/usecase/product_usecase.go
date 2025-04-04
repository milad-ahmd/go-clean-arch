package usecase

import (
	"context"

	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	"github.com/milad-ahmd/go-clean-arch/pkg/errors"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"go.uber.org/zap"
)

type productUseCase struct {
	productRepo  domain.ProductRepository
	categoryRepo domain.CategoryRepository
	logger       logger.Logger
}

// NewProductUseCase creates a new product use case
func NewProductUseCase(productRepo domain.ProductRepository, categoryRepo domain.CategoryRepository, logger logger.Logger) domain.ProductUseCase {
	return &productUseCase{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

// GetByID gets a product by ID
func (u *productUseCase) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	product, err := u.productRepo.FindByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get product by ID", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}
	return product, nil
}

// List lists products with pagination
func (u *productUseCase) List(ctx context.Context, limit, offset int) ([]domain.Product, int, error) {
	products, total, err := u.productRepo.FindAll(ctx, limit, offset)
	if err != nil {
		u.logger.Error("Failed to list products", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, 0, err
	}
	return products, total, nil
}

// Create creates a new product
func (u *productUseCase) Create(ctx context.Context, createDTO *domain.ProductCreateDTO) (*domain.Product, error) {
	// Check if product with the same SKU already exists
	existingProduct, err := u.productRepo.FindBySKU(ctx, createDTO.SKU)
	if err == nil && existingProduct != nil {
		return nil, errors.NewConflictError("Product", "sku", createDTO.SKU)
	}

	// Check if category exists
	_, err = u.categoryRepo.FindByID(ctx, createDTO.CategoryID)
	if err != nil {
		u.logger.Error("Failed to find category for product creation", zap.Int64("categoryID", createDTO.CategoryID), zap.Error(err))
		return nil, errors.NewBadRequestError("Invalid category ID")
	}

	// Create the product
	product := &domain.Product{
		Name:        createDTO.Name,
		Description: createDTO.Description,
		Price:       createDTO.Price,
		SKU:         createDTO.SKU,
		Stock:       createDTO.Stock,
		CategoryID:  createDTO.CategoryID,
		Images:      createDTO.Images,
	}

	if err := u.productRepo.Create(ctx, product); err != nil {
		u.logger.Error("Failed to create product", zap.Error(err))
		return nil, err
	}

	// Get the complete product with category
	product, err = u.productRepo.FindByID(ctx, product.ID)
	if err != nil {
		u.logger.Error("Failed to get created product", zap.Int64("id", product.ID), zap.Error(err))
		return product, nil // Return the product anyway, even if we couldn't get the category
	}

	return product, nil
}

// Update updates a product
func (u *productUseCase) Update(ctx context.Context, id int64, updateDTO *domain.ProductUpdateDTO) (*domain.Product, error) {
	// Get the existing product
	product, err := u.productRepo.FindByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get product for update", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	// Check if SKU is being updated and if it's already taken
	if updateDTO.SKU != "" && updateDTO.SKU != product.SKU {
		existingProduct, err := u.productRepo.FindBySKU(ctx, updateDTO.SKU)
		if err == nil && existingProduct != nil && existingProduct.ID != id {
			return nil, errors.NewConflictError("Product", "sku", updateDTO.SKU)
		}
		product.SKU = updateDTO.SKU
	}

	// Check if category is being updated and if it exists
	if updateDTO.CategoryID != 0 && updateDTO.CategoryID != product.CategoryID {
		_, err = u.categoryRepo.FindByID(ctx, updateDTO.CategoryID)
		if err != nil {
			u.logger.Error("Failed to find category for product update", zap.Int64("categoryID", updateDTO.CategoryID), zap.Error(err))
			return nil, errors.NewBadRequestError("Invalid category ID")
		}
		product.CategoryID = updateDTO.CategoryID
	}

	// Update other fields if provided
	if updateDTO.Name != "" {
		product.Name = updateDTO.Name
	}
	if updateDTO.Description != "" {
		product.Description = updateDTO.Description
	}
	if updateDTO.Price > 0 {
		product.Price = updateDTO.Price
	}
	if updateDTO.Stock >= 0 {
		product.Stock = updateDTO.Stock
	}
	if updateDTO.Images != nil {
		product.Images = updateDTO.Images
	}

	// Update the product
	if err := u.productRepo.Update(ctx, product); err != nil {
		u.logger.Error("Failed to update product", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	// Get the updated product with category
	product, err = u.productRepo.FindByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get updated product", zap.Int64("id", id), zap.Error(err))
		return product, nil // Return the product anyway, even if we couldn't get the category
	}

	return product, nil
}

// Delete deletes a product
func (u *productUseCase) Delete(ctx context.Context, id int64) error {
	if err := u.productRepo.Delete(ctx, id); err != nil {
		u.logger.Error("Failed to delete product", zap.Int64("id", id), zap.Error(err))
		return err
	}
	return nil
}

// GetBySKU gets a product by SKU
func (u *productUseCase) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	product, err := u.productRepo.FindBySKU(ctx, sku)
	if err != nil {
		u.logger.Error("Failed to get product by SKU", zap.String("sku", sku), zap.Error(err))
		return nil, err
	}
	return product, nil
}

// GetByCategory gets products by category ID
func (u *productUseCase) GetByCategory(ctx context.Context, categoryID int64, page, perPage int) ([]domain.Product, int, error) {
	// Check if category exists
	_, err := u.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		u.logger.Error("Failed to find category", zap.Int64("categoryID", categoryID), zap.Error(err))
		return nil, 0, err
	}

	// Calculate offset
	offset := (page - 1) * perPage
	if offset < 0 {
		offset = 0
	}

	products, total, err := u.productRepo.FindByCategory(ctx, categoryID, perPage, offset)
	if err != nil {
		u.logger.Error("Failed to get products by category", zap.Int64("categoryID", categoryID), zap.Error(err))
		return nil, 0, err
	}

	return products, total, nil
}

// UpdateStock updates a product's stock
func (u *productUseCase) UpdateStock(ctx context.Context, id int64, quantity int) error {
	if err := u.productRepo.UpdateStock(ctx, id, quantity); err != nil {
		u.logger.Error("Failed to update product stock", zap.Int64("id", id), zap.Int("quantity", quantity), zap.Error(err))
		return err
	}
	return nil
}

// Search searches for products
func (u *productUseCase) Search(ctx context.Context, query string, page, perPage int) ([]domain.Product, int, error) {
	// Calculate offset
	offset := (page - 1) * perPage
	if offset < 0 {
		offset = 0
	}

	products, total, err := u.productRepo.SearchProducts(ctx, query, perPage, offset)
	if err != nil {
		u.logger.Error("Failed to search products", zap.String("query", query), zap.Error(err))
		return nil, 0, err
	}

	return products, total, nil
}
