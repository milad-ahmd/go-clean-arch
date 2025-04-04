package domain

import (
	"context"
)

// Category represents a product category
type Category struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
	BaseEntity
}

// CategoryRepository defines the category repository interface
type CategoryRepository interface {
	BaseRepository[Category, int64]
	FindBySlug(ctx context.Context, slug string) (*Category, error)
	FindByName(ctx context.Context, name string) (*Category, error)
}

// CategoryCreateDTO represents the data for creating a category
type CategoryCreateDTO struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
	Slug        string `json:"slug" validate:"required,min=3,max=100,alphanum"`
}

// CategoryUpdateDTO represents the data for updating a category
type CategoryUpdateDTO struct {
	Name        string `json:"name" validate:"omitempty,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
	Slug        string `json:"slug" validate:"omitempty,min=3,max=100,alphanum"`
}

// CategoryUseCase defines the category use case interface
type CategoryUseCase interface {
	BaseUseCase[Category, int64, CategoryCreateDTO, CategoryUpdateDTO]
	GetBySlug(ctx context.Context, slug string) (*Category, error)
}
