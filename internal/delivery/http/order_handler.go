package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	"github.com/milad-ahmd/go-clean-arch/pkg/errors"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"github.com/milad-ahmd/go-clean-arch/pkg/middleware"
	"github.com/milad-ahmd/go-clean-arch/pkg/response"
	"go.uber.org/zap"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	orderUseCase domain.OrderUseCase
	logger       logger.Logger
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(r *mux.Router, orderUseCase domain.OrderUseCase, logger logger.Logger) {
	handler := &OrderHandler{
		orderUseCase: orderUseCase,
		logger:       logger,
	}

	// Public routes
	r.HandleFunc("/orders/{id:[0-9]+}", handler.GetByID).Methods("GET")

	// Protected routes (require authentication)
	protected := r.PathPrefix("/orders").Subrouter()
	// Note: In a real application, you would need to pass the userUseCase to the Auth middleware
	// This is a placeholder that will need to be updated when integrating with the actual application
	// protected.Use(middleware.Auth(userUseCase, logger))
	protected.HandleFunc("", handler.Create).Methods("POST")
	protected.HandleFunc("", handler.List).Methods("GET")
	protected.HandleFunc("/{id:[0-9]+}", handler.Update).Methods("PUT")
	protected.HandleFunc("/{id:[0-9]+}", handler.Delete).Methods("DELETE")
	protected.HandleFunc("/{id:[0-9]+}/status", handler.UpdateStatus).Methods("PATCH")
	protected.HandleFunc("/user/{userID:[0-9]+}", handler.GetByUserID).Methods("GET")
	protected.HandleFunc("/status/{status}", handler.GetByStatus).Methods("GET")
}

// Create handles the creation of a new order
// @Summary Create order
// @Description Create a new order
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.OrderCreateDTO true "Order Create Request"
// @Success 201 {object} response.Response{data=domain.Order}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders [post]
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, "Unauthorized", errors.NewUnauthorizedError(""), http.StatusUnauthorized)
		return
	}

	var createDTO domain.OrderCreateDTO
	if err := json.NewDecoder(r.Body).Decode(&createDTO); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.Error(w, "Invalid request payload", errors.NewBadRequestError("Invalid request payload"), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Set the user ID from the authenticated user
	createDTO.UserID = user.ID

	order, err := h.orderUseCase.Create(r.Context(), &createDTO)
	if err != nil {
		h.logger.Error("Failed to create order", zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to create order", err, statusCode)
		return
	}

	response.Success(w, "Order created successfully", order, http.StatusCreated)
}

// GetByID handles getting an order by ID
// @Summary Get order by ID
// @Description Get an order by its ID
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.Response{data=domain.Order}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id} [get]
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse order ID", zap.Error(err))
		response.Error(w, "Invalid order ID", errors.NewBadRequestError("Invalid order ID"), http.StatusBadRequest)
		return
	}

	order, err := h.orderUseCase.GetOrderWithDetails(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get order", zap.Int64("id", id), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to get order", err, statusCode)
		return
	}

	// Check if the user is authorized to view this order
	user, ok := middleware.GetUserFromContext(r.Context())
	if ok && user.ID != order.UserID && user.Role != domain.RoleAdmin {
		response.Error(w, "Forbidden", errors.NewForbiddenError(""), http.StatusForbidden)
		return
	}

	response.Success(w, "Order retrieved successfully", order, http.StatusOK)
}

// Update handles updating an order
// @Summary Update order
// @Description Update an order by its ID
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body domain.OrderUpdateDTO true "Order Update Request"
// @Success 200 {object} response.Response{data=domain.Order}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id} [put]
func (h *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse order ID", zap.Error(err))
		response.Error(w, "Invalid order ID", errors.NewBadRequestError("Invalid order ID"), http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, "Unauthorized", errors.NewUnauthorizedError(""), http.StatusUnauthorized)
		return
	}

	// Check if the order exists and belongs to the user
	order, err := h.orderUseCase.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get order for update", zap.Int64("id", id), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to get order", err, statusCode)
		return
	}

	// Only the order owner or an admin can update the order
	if user.ID != order.UserID && user.Role != domain.RoleAdmin {
		response.Error(w, "Forbidden", errors.NewForbiddenError(""), http.StatusForbidden)
		return
	}

	var updateDTO domain.OrderUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&updateDTO); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.Error(w, "Invalid request payload", errors.NewBadRequestError("Invalid request payload"), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Only admins can change the order status
	if updateDTO.Status != "" && user.Role != domain.RoleAdmin {
		response.Error(w, "Forbidden", errors.NewForbiddenError("Only admins can change order status"), http.StatusForbidden)
		return
	}

	updatedOrder, err := h.orderUseCase.Update(r.Context(), id, &updateDTO)
	if err != nil {
		h.logger.Error("Failed to update order", zap.Int64("id", id), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to update order", err, statusCode)
		return
	}

	response.Success(w, "Order updated successfully", updatedOrder, http.StatusOK)
}

// Delete handles deleting an order
// @Summary Delete order
// @Description Delete an order by its ID
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id} [delete]
func (h *OrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse order ID", zap.Error(err))
		response.Error(w, "Invalid order ID", errors.NewBadRequestError("Invalid order ID"), http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, "Unauthorized", errors.NewUnauthorizedError(""), http.StatusUnauthorized)
		return
	}

	// Check if the order exists and belongs to the user
	order, err := h.orderUseCase.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get order for deletion", zap.Int64("id", id), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to get order", err, statusCode)
		return
	}

	// Only the order owner or an admin can delete the order
	if user.ID != order.UserID && user.Role != domain.RoleAdmin {
		response.Error(w, "Forbidden", errors.NewForbiddenError(""), http.StatusForbidden)
		return
	}

	if err := h.orderUseCase.Delete(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete order", zap.Int64("id", id), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to delete order", err, statusCode)
		return
	}

	response.Success(w, "Order deleted successfully", nil, http.StatusOK)
}

// List handles listing orders with pagination
// @Summary List orders
// @Description List orders with pagination
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} response.PaginatedResponse{data=[]domain.Order}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders [get]
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, "Unauthorized", errors.NewUnauthorizedError(""), http.StatusUnauthorized)
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

	offset := (page - 1) * perPage

	var orders []domain.Order
	var total int
	var err error

	// If user is admin, show all orders, otherwise show only user's orders
	if user.Role == domain.RoleAdmin {
		orders, total, err = h.orderUseCase.List(r.Context(), perPage, offset)
	} else {
		orders, total, err = h.orderUseCase.GetByUserID(r.Context(), user.ID, page, perPage)
	}

	if err != nil {
		h.logger.Error("Failed to list orders", zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to list orders", err, statusCode)
		return
	}

	response.Paginated(w, "Orders retrieved successfully", orders, page, perPage, total, http.StatusOK)
}

// UpdateStatus handles updating an order's status
// @Summary Update order status
// @Description Update an order's status
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body map[string]string true "Status Update Request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id}/status [patch]
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse order ID", zap.Error(err))
		response.Error(w, "Invalid order ID", errors.NewBadRequestError("Invalid order ID"), http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, "Unauthorized", errors.NewUnauthorizedError(""), http.StatusUnauthorized)
		return
	}

	// Only admins can update order status
	if user.Role != domain.RoleAdmin {
		response.Error(w, "Forbidden", errors.NewForbiddenError("Only admins can update order status"), http.StatusForbidden)
		return
	}

	var statusUpdate struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.Error(w, "Invalid request payload", errors.NewBadRequestError("Invalid request payload"), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate status
	if statusUpdate.Status != string(domain.OrderStatusPending) &&
		statusUpdate.Status != string(domain.OrderStatusProcessing) &&
		statusUpdate.Status != string(domain.OrderStatusCompleted) &&
		statusUpdate.Status != string(domain.OrderStatusCancelled) {
		response.Error(w, "Invalid status", errors.NewBadRequestError("Invalid status"), http.StatusBadRequest)
		return
	}

	if err := h.orderUseCase.UpdateStatus(r.Context(), id, domain.OrderStatus(statusUpdate.Status)); err != nil {
		h.logger.Error("Failed to update order status", zap.Int64("id", id), zap.String("status", statusUpdate.Status), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to update order status", err, statusCode)
		return
	}

	response.Success(w, "Order status updated successfully", nil, http.StatusOK)
}

// GetByUserID handles getting orders by user ID
// @Summary Get orders by user ID
// @Description Get orders by user ID with pagination
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userID path int true "User ID"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} response.PaginatedResponse{data=[]domain.Order}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/user/{userID} [get]
func (h *OrderHandler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseInt(vars["userID"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse user ID", zap.Error(err))
		response.Error(w, "Invalid user ID", errors.NewBadRequestError("Invalid user ID"), http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, "Unauthorized", errors.NewUnauthorizedError(""), http.StatusUnauthorized)
		return
	}

	// Users can only see their own orders, admins can see any user's orders
	if user.ID != userID && user.Role != domain.RoleAdmin {
		response.Error(w, "Forbidden", errors.NewForbiddenError(""), http.StatusForbidden)
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

	orders, total, err := h.orderUseCase.GetByUserID(r.Context(), userID, page, perPage)
	if err != nil {
		h.logger.Error("Failed to get orders by user ID", zap.Int64("userID", userID), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to get orders", err, statusCode)
		return
	}

	response.Paginated(w, "Orders retrieved successfully", orders, page, perPage, total, http.StatusOK)
}

// GetByStatus handles getting orders by status
// @Summary Get orders by status
// @Description Get orders by status with pagination
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status path string true "Order Status"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} response.PaginatedResponse{data=[]domain.Order}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/status/{status} [get]
func (h *OrderHandler) GetByStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	status := vars["status"]

	// Validate status
	if status != string(domain.OrderStatusPending) &&
		status != string(domain.OrderStatusProcessing) &&
		status != string(domain.OrderStatusCompleted) &&
		status != string(domain.OrderStatusCancelled) {
		response.Error(w, "Invalid status", errors.NewBadRequestError("Invalid status"), http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, "Unauthorized", errors.NewUnauthorizedError(""), http.StatusUnauthorized)
		return
	}

	// Only admins can see all orders by status
	if user.Role != domain.RoleAdmin {
		response.Error(w, "Forbidden", errors.NewForbiddenError(""), http.StatusForbidden)
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

	orders, total, err := h.orderUseCase.GetByStatus(r.Context(), domain.OrderStatus(status), page, perPage)
	if err != nil {
		h.logger.Error("Failed to get orders by status", zap.String("status", status), zap.Error(err))
		statusCode := errors.GetStatusCode(err)
		response.Error(w, "Failed to get orders", err, statusCode)
		return
	}

	response.Paginated(w, "Orders retrieved successfully", orders, page, perPage, total, http.StatusOK)
}
