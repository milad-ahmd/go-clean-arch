package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	"github.com/milad-ahmd/go-clean-arch/pkg/errors"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"github.com/milad-ahmd/go-clean-arch/pkg/response"
	"go.uber.org/zap"
)

// CategoryHandler handles HTTP requests for categories
type CategoryHandler struct {
	categoryUseCase domain.CategoryUseCase
	logger          logger.Logger
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(r *mux.Router, categoryUseCase domain.CategoryUseCase, logger logger.Logger) {
	handler := &CategoryHandler{
		categoryUseCase: categoryUseCase,
		logger:          logger,
	}

	// Register routes
	r.HandleFunc("/categories", handler.Create).Methods("POST")
	r.HandleFunc("/categories", handler.List).Methods("GET")
	r.HandleFunc("/categories/{id:[0-9]+}", handler.GetByID).Methods("GET")
	r.HandleFunc("/categories/{id:[0-9]+}", handler.Update).Methods("PUT")
	r.HandleFunc("/categories/{id:[0-9]+}", handler.Delete).Methods("DELETE")
	r.HandleFunc("/categories/slug/{slug}", handler.GetBySlug).Methods("GET")
}

// Create handles the creation of a new category
// @Summary Create category
// @Description Create a new category
// @Tags categories
// @Accept json
// @Produce json
// @Param request body domain.CategoryCreateDTO true "Category Create Request"
// @Success 201 {object} response.Response{data=domain.Category}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /categories [post]
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var createDTO domain.CategoryCreateDTO
	if err := json.NewDecoder(r.Body).Decode(&createDTO); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.Error(w, "Invalid request payload", errors.NewBadRequestError("Invalid request payload"), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	category, err := h.categoryUseCase.Create(r.Context(), &createDTO)
	if err != nil {
		h.logger.Error("Failed to create category", zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to create category", err, statusCode)
		return
	}

	response.Success(w, "Category created successfully", category, http.StatusCreated)
}

// GetByID handles getting a category by ID
// @Summary Get category by ID
// @Description Get a category by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.Response{data=domain.Category}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /categories/{id} [get]
func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse category ID", zap.Error(err))
		response.Error(w, "Invalid category ID", errors.NewBadRequestError("Invalid category ID"), http.StatusBadRequest)
		return
	}

	category, err := h.categoryUseCase.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get category", zap.Int64("id", id), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to get category", err, statusCode)
		return
	}

	response.Success(w, "Category retrieved successfully", category, http.StatusOK)
}

// Update handles updating a category
// @Summary Update category
// @Description Update a category by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param request body domain.CategoryUpdateDTO true "Category Update Request"
// @Success 200 {object} response.Response{data=domain.Category}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /categories/{id} [put]
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse category ID", zap.Error(err))
		response.Error(w, "Invalid category ID", errors.NewBadRequestError("Invalid category ID"), http.StatusBadRequest)
		return
	}

	var updateDTO domain.CategoryUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&updateDTO); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.Error(w, "Invalid request payload", errors.NewBadRequestError("Invalid request payload"), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	category, err := h.categoryUseCase.Update(r.Context(), id, &updateDTO)
	if err != nil {
		h.logger.Error("Failed to update category", zap.Int64("id", id), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to update category", err, statusCode)
		return
	}

	response.Success(w, "Category updated successfully", category, http.StatusOK)
}

// Delete handles deleting a category
// @Summary Delete category
// @Description Delete a category by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /categories/{id} [delete]
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse category ID", zap.Error(err))
		response.Error(w, "Invalid category ID", errors.NewBadRequestError("Invalid category ID"), http.StatusBadRequest)
		return
	}

	if err := h.categoryUseCase.Delete(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete category", zap.Int64("id", id), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to delete category", err, statusCode)
		return
	}

	response.Success(w, "Category deleted successfully", nil, http.StatusOK)
}

// List handles listing categories with pagination
// @Summary List categories
// @Description List categories with pagination
// @Tags categories
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} response.PaginatedResponse{data=[]domain.Category}
// @Failure 500 {object} response.Response
// @Router /categories [get]
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	offset := (page - 1) * perPage

	categories, total, err := h.categoryUseCase.List(r.Context(), perPage, offset)
	if err != nil {
		h.logger.Error("Failed to list categories", zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to list categories", err, statusCode)
		return
	}

	response.Paginated(w, "Categories retrieved successfully", categories, page, perPage, total, http.StatusOK)
}

// GetBySlug handles getting a category by slug
// @Summary Get category by slug
// @Description Get a category by its slug
// @Tags categories
// @Accept json
// @Produce json
// @Param slug path string true "Category Slug"
// @Success 200 {object} response.Response{data=domain.Category}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /categories/slug/{slug} [get]
func (h *CategoryHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	category, err := h.categoryUseCase.GetBySlug(r.Context(), slug)
	if err != nil {
		h.logger.Error("Failed to get category by slug", zap.String("slug", slug), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to get category", err, statusCode)
		return
	}

	response.Success(w, "Category retrieved successfully", category, http.StatusOK)
}
