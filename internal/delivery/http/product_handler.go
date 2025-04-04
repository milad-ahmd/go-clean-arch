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

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	productUseCase domain.ProductUseCase
	logger         logger.Logger
}

// NewProductHandler creates a new product handler
func NewProductHandler(r *mux.Router, productUseCase domain.ProductUseCase, logger logger.Logger) {
	handler := &ProductHandler{
		productUseCase: productUseCase,
		logger:         logger,
	}

	// Register routes
	r.HandleFunc("/products", handler.Create).Methods("POST")
	r.HandleFunc("/products", handler.List).Methods("GET")
	r.HandleFunc("/products/{id:[0-9]+}", handler.GetByID).Methods("GET")
	r.HandleFunc("/products/{id:[0-9]+}", handler.Update).Methods("PUT")
	r.HandleFunc("/products/{id:[0-9]+}", handler.Delete).Methods("DELETE")
	r.HandleFunc("/products/sku/{sku}", handler.GetBySKU).Methods("GET")
	r.HandleFunc("/products/category/{categoryID:[0-9]+}", handler.GetByCategory).Methods("GET")
	r.HandleFunc("/products/search", handler.Search).Methods("GET")
	r.HandleFunc("/products/{id:[0-9]+}/stock", handler.UpdateStock).Methods("PATCH")
}

// Create handles the creation of a new product
// @Summary Create product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Param request body domain.ProductCreateDTO true "Product Create Request"
// @Success 201 {object} response.Response{data=domain.Product}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /products [post]
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var createDTO domain.ProductCreateDTO
	if err := json.NewDecoder(r.Body).Decode(&createDTO); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.Error(w, "Invalid request payload", errors.NewBadRequestError("Invalid request payload"), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	product, err := h.productUseCase.Create(r.Context(), &createDTO)
	if err != nil {
		h.logger.Error("Failed to create product", zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to create product", err, statusCode)
		return
	}

	response.Success(w, "Product created successfully", product, http.StatusCreated)
}

// GetByID handles getting a product by ID
// @Summary Get product by ID
// @Description Get a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} response.Response{data=domain.Product}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /products/{id} [get]
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse product ID", zap.Error(err))
		response.Error(w, "Invalid product ID", errors.NewBadRequestError("Invalid product ID"), http.StatusBadRequest)
		return
	}

	product, err := h.productUseCase.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get product", zap.Int64("id", id), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to get product", err, statusCode)
		return
	}

	response.Success(w, "Product retrieved successfully", product, http.StatusOK)
}

// Update handles updating a product
// @Summary Update product
// @Description Update a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param request body domain.ProductUpdateDTO true "Product Update Request"
// @Success 200 {object} response.Response{data=domain.Product}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /products/{id} [put]
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse product ID", zap.Error(err))
		response.Error(w, "Invalid product ID", errors.NewBadRequestError("Invalid product ID"), http.StatusBadRequest)
		return
	}

	var updateDTO domain.ProductUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&updateDTO); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.Error(w, "Invalid request payload", errors.NewBadRequestError("Invalid request payload"), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	product, err := h.productUseCase.Update(r.Context(), id, &updateDTO)
	if err != nil {
		h.logger.Error("Failed to update product", zap.Int64("id", id), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to update product", err, statusCode)
		return
	}

	response.Success(w, "Product updated successfully", product, http.StatusOK)
}

// Delete handles deleting a product
// @Summary Delete product
// @Description Delete a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /products/{id} [delete]
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse product ID", zap.Error(err))
		response.Error(w, "Invalid product ID", errors.NewBadRequestError("Invalid product ID"), http.StatusBadRequest)
		return
	}

	if err := h.productUseCase.Delete(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete product", zap.Int64("id", id), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to delete product", err, statusCode)
		return
	}

	response.Success(w, "Product deleted successfully", nil, http.StatusOK)
}

// List handles listing products with pagination
// @Summary List products
// @Description List products with pagination
// @Tags products
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} response.PaginatedResponse{data=[]domain.Product}
// @Failure 500 {object} response.Response
// @Router /products [get]
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
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

	products, total, err := h.productUseCase.List(r.Context(), perPage, offset)
	if err != nil {
		h.logger.Error("Failed to list products", zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to list products", err, statusCode)
		return
	}

	response.Paginated(w, "Products retrieved successfully", products, page, perPage, total, http.StatusOK)
}

// GetBySKU handles getting a product by SKU
// @Summary Get product by SKU
// @Description Get a product by its SKU
// @Tags products
// @Accept json
// @Produce json
// @Param sku path string true "Product SKU"
// @Success 200 {object} response.Response{data=domain.Product}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /products/sku/{sku} [get]
func (h *ProductHandler) GetBySKU(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sku := vars["sku"]

	product, err := h.productUseCase.GetBySKU(r.Context(), sku)
	if err != nil {
		h.logger.Error("Failed to get product by SKU", zap.String("sku", sku), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to get product", err, statusCode)
		return
	}

	response.Success(w, "Product retrieved successfully", product, http.StatusOK)
}

// GetByCategory handles getting products by category ID
// @Summary Get products by category
// @Description Get products by category ID with pagination
// @Tags products
// @Accept json
// @Produce json
// @Param categoryID path int true "Category ID"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} response.PaginatedResponse{data=[]domain.Product}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /products/category/{categoryID} [get]
func (h *ProductHandler) GetByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.ParseInt(vars["categoryID"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse category ID", zap.Error(err))
		response.Error(w, "Invalid category ID", errors.NewBadRequestError("Invalid category ID"), http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	products, total, err := h.productUseCase.GetByCategory(r.Context(), categoryID, page, perPage)
	if err != nil {
		h.logger.Error("Failed to get products by category", zap.Int64("categoryID", categoryID), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to get products", err, statusCode)
		return
	}

	response.Paginated(w, "Products retrieved successfully", products, page, perPage, total, http.StatusOK)
}

// Search handles searching for products
// @Summary Search products
// @Description Search for products by name or description
// @Tags products
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} response.PaginatedResponse{data=[]domain.Product}
// @Failure 500 {object} response.Response
// @Router /products/search [get]
func (h *ProductHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		response.Error(w, "Search query is required", errors.NewBadRequestError("Search query is required"), http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	products, total, err := h.productUseCase.Search(r.Context(), query, page, perPage)
	if err != nil {
		h.logger.Error("Failed to search products", zap.String("query", query), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to search products", err, statusCode)
		return
	}

	response.Paginated(w, "Products retrieved successfully", products, page, perPage, total, http.StatusOK)
}

// UpdateStock handles updating a product's stock
// @Summary Update product stock
// @Description Update a product's stock quantity
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param request body map[string]int true "Stock Update Request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /products/{id}/stock [patch]
func (h *ProductHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse product ID", zap.Error(err))
		response.Error(w, "Invalid product ID", errors.NewBadRequestError("Invalid product ID"), http.StatusBadRequest)
		return
	}

	var stockUpdate struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&stockUpdate); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.Error(w, "Invalid request payload", errors.NewBadRequestError("Invalid request payload"), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := h.productUseCase.UpdateStock(r.Context(), id, stockUpdate.Quantity); err != nil {
		h.logger.Error("Failed to update product stock", zap.Int64("id", id), zap.Int("quantity", stockUpdate.Quantity), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to update product stock", err, statusCode)
		return
	}

	response.Success(w, "Product stock updated successfully", nil, http.StatusOK)
}
