package usecase

import (
	"context"

	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	"github.com/milad-ahmd/go-clean-arch/pkg/errors"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"go.uber.org/zap"
)

type categoryUseCase struct {
	categoryRepo domain.CategoryRepository
	logger       logger.Logger
}

// NewCategoryUseCase creates a new category use case
func NewCategoryUseCase(categoryRepo domain.CategoryRepository, logger logger.Logger) domain.CategoryUseCase {
	return &categoryUseCase{
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

// GetByID gets a category by ID
func (u *categoryUseCase) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	category, err := u.categoryRepo.FindByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get category by ID", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}
	return category, nil
}

// List lists categories with pagination
func (u *categoryUseCase) List(ctx context.Context, limit, offset int) ([]domain.Category, int, error) {
	categories, total, err := u.categoryRepo.FindAll(ctx, limit, offset)
	if err != nil {
		u.logger.Error("Failed to list categories", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, 0, err
	}
	return categories, total, nil
}

// Create creates a new category
func (u *categoryUseCase) Create(ctx context.Context, createDTO *domain.CategoryCreateDTO) (*domain.Category, error) {
	// Check if category with the same name already exists
	existingCategory, err := u.categoryRepo.FindByName(ctx, createDTO.Name)
	if err == nil && existingCategory != nil {
		return nil, errors.NewConflictError("Category", "name", createDTO.Name)
	}

	// Check if category with the same slug already exists
	existingCategory, err = u.categoryRepo.FindBySlug(ctx, createDTO.Slug)
	if err == nil && existingCategory != nil {
		return nil, errors.NewConflictError("Category", "slug", createDTO.Slug)
	}

	// Create the category
	category := &domain.Category{
		Name:        createDTO.Name,
		Description: createDTO.Description,
		Slug:        createDTO.Slug,
	}

	if err := u.categoryRepo.Create(ctx, category); err != nil {
		u.logger.Error("Failed to create category", zap.Error(err))
		return nil, err
	}

	return category, nil
}

// Update updates a category
func (u *categoryUseCase) Update(ctx context.Context, id int64, updateDTO *domain.CategoryUpdateDTO) (*domain.Category, error) {
	// Get the existing category
	category, err := u.categoryRepo.FindByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get category for update", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	// Check if name is being updated and if it's already taken
	if updateDTO.Name != "" && updateDTO.Name != category.Name {
		existingCategory, err := u.categoryRepo.FindByName(ctx, updateDTO.Name)
		if err == nil && existingCategory != nil && existingCategory.ID != id {
			return nil, errors.NewConflictError("Category", "name", updateDTO.Name)
		}
		category.Name = updateDTO.Name
	}

	// Check if slug is being updated and if it's already taken
	if updateDTO.Slug != "" && updateDTO.Slug != category.Slug {
		existingCategory, err := u.categoryRepo.FindBySlug(ctx, updateDTO.Slug)
		if err == nil && existingCategory != nil && existingCategory.ID != id {
			return nil, errors.NewConflictError("Category", "slug", updateDTO.Slug)
		}
		category.Slug = updateDTO.Slug
	}

	// Update description if provided
	if updateDTO.Description != "" {
		category.Description = updateDTO.Description
	}

	// Update the category
	if err := u.categoryRepo.Update(ctx, category); err != nil {
		u.logger.Error("Failed to update category", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	return category, nil
}

// Delete deletes a category
func (u *categoryUseCase) Delete(ctx context.Context, id int64) error {
	if err := u.categoryRepo.Delete(ctx, id); err != nil {
		u.logger.Error("Failed to delete category", zap.Int64("id", id), zap.Error(err))
		return err
	}
	return nil
}

// GetBySlug gets a category by slug
func (u *categoryUseCase) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	category, err := u.categoryRepo.FindBySlug(ctx, slug)
	if err != nil {
		u.logger.Error("Failed to get category by slug", zap.String("slug", slug), zap.Error(err))
		return nil, err
	}
	return category, nil
}
