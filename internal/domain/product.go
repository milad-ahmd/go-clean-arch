package domain

import (
	"context"
)

// Product represents a product entity
type Product struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	SKU         string   `json:"sku"`
	Stock       int      `json:"stock"`
	CategoryID  int64    `json:"category_id"`
	Category    Category `json:"category,omitempty"`
	Images      []string `json:"images,omitempty"`
	BaseEntity
}

// ProductRepository defines the product repository interface
type ProductRepository interface {
	BaseRepository[Product, int64]
	FindBySKU(ctx context.Context, sku string) (*Product, error)
	FindByCategory(ctx context.Context, categoryID int64, limit, offset int) ([]Product, int, error)
	UpdateStock(ctx context.Context, id int64, quantity int) error
	SearchProducts(ctx context.Context, query string, limit, offset int) ([]Product, int, error)
}

// ProductCreateDTO represents the data for creating a product
type ProductCreateDTO struct {
	Name        string   `json:"name" validate:"required,min=3,max=100"`
	Description string   `json:"description" validate:"max=1000"`
	Price       float64  `json:"price" validate:"required,gt=0"`
	SKU         string   `json:"sku" validate:"required,min=3,max=50"`
	Stock       int      `json:"stock" validate:"required,gte=0"`
	CategoryID  int64    `json:"category_id" validate:"required,gt=0"`
	Images      []string `json:"images" validate:"dive,url"`
}

// ProductUpdateDTO represents the data for updating a product
type ProductUpdateDTO struct {
	Name        string   `json:"name" validate:"omitempty,min=3,max=100"`
	Description string   `json:"description" validate:"max=1000"`
	Price       float64  `json:"price" validate:"omitempty,gt=0"`
	SKU         string   `json:"sku" validate:"omitempty,min=3,max=50"`
	Stock       int      `json:"stock" validate:"omitempty,gte=0"`
	CategoryID  int64    `json:"category_id" validate:"omitempty,gt=0"`
	Images      []string `json:"images" validate:"omitempty,dive,url"`
}

// ProductUseCase defines the product use case interface
type ProductUseCase interface {
	BaseUseCase[Product, int64, ProductCreateDTO, ProductUpdateDTO]
	GetBySKU(ctx context.Context, sku string) (*Product, error)
	GetByCategory(ctx context.Context, categoryID int64, page, perPage int) ([]Product, int, error)
	UpdateStock(ctx context.Context, id int64, quantity int) error
	Search(ctx context.Context, query string, page, perPage int) ([]Product, int, error)
}
