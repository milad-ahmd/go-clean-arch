package domain

import (
	"context"
)

// BaseRepository defines the base repository interface
type BaseRepository[T any, ID any] interface {
	FindByID(ctx context.Context, id ID) (*T, error)
	FindAll(ctx context.Context, limit, offset int) ([]T, int, error)
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id ID) error
}

// BaseUseCase defines the base use case interface
type BaseUseCase[T any, ID any, C any, U any] interface {
	GetByID(ctx context.Context, id ID) (*T, error)
	List(ctx context.Context, limit, offset int) ([]T, int, error)
	Create(ctx context.Context, createDTO *C) (*T, error)
	Update(ctx context.Context, id ID, updateDTO *U) (*T, error)
	Delete(ctx context.Context, id ID) error
}

// Pagination represents pagination parameters
type Pagination struct {
	Page    int `json:"page" query:"page"`
	PerPage int `json:"per_page" query:"per_page"`
}

// GetOffset returns the offset for pagination
func (p *Pagination) GetOffset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 10
	}
	return (p.Page - 1) * p.PerPage
}

// GetLimit returns the limit for pagination
func (p *Pagination) GetLimit() int {
	if p.PerPage < 1 {
		p.PerPage = 10
	}
	return p.PerPage
}

// BaseEntity defines common fields for all entities
type BaseEntity struct {
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}
