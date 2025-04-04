package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	pkgerrors "github.com/milad-ahmd/go-clean-arch/pkg/errors"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"go.uber.org/zap"
)

type orderUseCase struct {
	orderRepo   domain.OrderRepository
	productRepo domain.ProductRepository
	userRepo    domain.UserRepository
	logger      logger.Logger
}

// NewOrderUseCase creates a new order use case
func NewOrderUseCase(orderRepo domain.OrderRepository, productRepo domain.ProductRepository, userRepo domain.UserRepository, logger logger.Logger) domain.OrderUseCase {
	return &orderUseCase{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// GetByID gets an order by ID
func (u *orderUseCase) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get order by ID", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}
	return order, nil
}

// List lists orders with pagination
func (u *orderUseCase) List(ctx context.Context, limit, offset int) ([]domain.Order, int, error) {
	orders, total, err := u.orderRepo.FindAll(ctx, limit, offset)
	if err != nil {
		u.logger.Error("Failed to list orders", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, 0, err
	}
	return orders, total, nil
}

// Create creates a new order
func (u *orderUseCase) Create(ctx context.Context, createDTO *domain.OrderCreateDTO) (*domain.Order, error) {
	// Check if user exists
	user, err := u.userRepo.GetByID(ctx, createDTO.UserID)
	if err != nil {
		u.logger.Error("Failed to find user for order creation", zap.Int64("userID", createDTO.UserID), zap.Error(err))
		return nil, pkgerrors.NewBadRequestError("Invalid user ID")
	}

	// Validate order items
	if len(createDTO.Items) == 0 {
		return nil, pkgerrors.NewBadRequestError("Order must have at least one item")
	}

	// Create order items and calculate total amount
	var orderItems []domain.OrderItem
	var totalAmount float64

	for _, itemDTO := range createDTO.Items {
		// Check if product exists and has enough stock
		product, err := u.productRepo.FindByID(ctx, itemDTO.ProductID)
		if err != nil {
			u.logger.Error("Failed to find product for order item", zap.Int64("productID", itemDTO.ProductID), zap.Error(err))
			return nil, pkgerrors.NewBadRequestError("Invalid product ID: " + fmt.Sprintf("%d", itemDTO.ProductID))
		}

		if product.Stock < itemDTO.Quantity {
			return nil, pkgerrors.NewBadRequestError("Insufficient stock for product: " + product.Name)
		}

		// Create order item
		orderItem := domain.OrderItem{
			ProductID: itemDTO.ProductID,
			Quantity:  itemDTO.Quantity,
			Price:     product.Price, // Use the current product price
			Product:   *product,
		}

		orderItems = append(orderItems, orderItem)
		totalAmount += product.Price * float64(itemDTO.Quantity)
	}

	// Create shipping info
	shippingInfo := domain.ShippingInfo{
		Address:     createDTO.ShippingInfo.Address,
		City:        createDTO.ShippingInfo.City,
		State:       createDTO.ShippingInfo.State,
		Country:     createDTO.ShippingInfo.Country,
		PostalCode:  createDTO.ShippingInfo.PostalCode,
		PhoneNumber: createDTO.ShippingInfo.PhoneNumber,
	}

	// Create the order
	order := &domain.Order{
		UserID:        createDTO.UserID,
		Status:        domain.OrderStatusPending,
		TotalAmount:   totalAmount,
		Items:         orderItems,
		PaymentMethod: createDTO.PaymentMethod,
		ShippingInfo:  shippingInfo,
		User:          *user,
	}

	if err := u.orderRepo.Create(ctx, order); err != nil {
		u.logger.Error("Failed to create order", zap.Error(err))
		return nil, err
	}

	return order, nil
}

// Update updates an order
func (u *orderUseCase) Update(ctx context.Context, id int64, updateDTO *domain.OrderUpdateDTO) (*domain.Order, error) {
	// Get the existing order
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get order for update", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	// Update status if provided
	if updateDTO.Status != "" {
		order.Status = updateDTO.Status
	}

	// Update payment method if provided
	if updateDTO.PaymentMethod != "" {
		order.PaymentMethod = updateDTO.PaymentMethod
	}

	// Update shipping info if provided
	if updateDTO.ShippingInfo.Address != "" {
		order.ShippingInfo.Address = updateDTO.ShippingInfo.Address
		order.ShippingInfo.City = updateDTO.ShippingInfo.City
		order.ShippingInfo.State = updateDTO.ShippingInfo.State
		order.ShippingInfo.Country = updateDTO.ShippingInfo.Country
		order.ShippingInfo.PostalCode = updateDTO.ShippingInfo.PostalCode
		order.ShippingInfo.PhoneNumber = updateDTO.ShippingInfo.PhoneNumber
	}

	// Update the order
	if err := u.orderRepo.Update(ctx, order); err != nil {
		u.logger.Error("Failed to update order", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	// Get the updated order
	order, err = u.orderRepo.FindByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get updated order", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	return order, nil
}

// Delete deletes an order
func (u *orderUseCase) Delete(ctx context.Context, id int64) error {
	if err := u.orderRepo.Delete(ctx, id); err != nil {
		u.logger.Error("Failed to delete order", zap.Int64("id", id), zap.Error(err))
		return err
	}
	return nil
}

// GetByUserID gets orders by user ID
func (u *orderUseCase) GetByUserID(ctx context.Context, userID int64, page, perPage int) ([]domain.Order, int, error) {
	// Check if user exists
	_, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		u.logger.Error("Failed to find user", zap.Int64("userID", userID), zap.Error(err))
		return nil, 0, err
	}

	// Calculate offset
	offset := (page - 1) * perPage
	if offset < 0 {
		offset = 0
	}

	orders, total, err := u.orderRepo.FindByUserID(ctx, userID, perPage, offset)
	if err != nil {
		u.logger.Error("Failed to get orders by user ID", zap.Int64("userID", userID), zap.Error(err))
		return nil, 0, err
	}

	return orders, total, nil
}

// GetByStatus gets orders by status
func (u *orderUseCase) GetByStatus(ctx context.Context, status domain.OrderStatus, page, perPage int) ([]domain.Order, int, error) {
	// Calculate offset
	offset := (page - 1) * perPage
	if offset < 0 {
		offset = 0
	}

	orders, total, err := u.orderRepo.FindByStatus(ctx, status, perPage, offset)
	if err != nil {
		u.logger.Error("Failed to get orders by status", zap.String("status", string(status)), zap.Error(err))
		return nil, 0, err
	}

	return orders, total, nil
}

// UpdateStatus updates an order's status
func (u *orderUseCase) UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error {
	if err := u.orderRepo.UpdateStatus(ctx, id, status); err != nil {
		u.logger.Error("Failed to update order status", zap.Int64("id", id), zap.String("status", string(status)), zap.Error(err))
		return err
	}
	return nil
}

// GetOrderWithDetails gets an order with all details
func (u *orderUseCase) GetOrderWithDetails(ctx context.Context, id int64) (*domain.Order, error) {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get order with details", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	// Get order items with product details
	items, err := u.orderRepo.GetOrderItems(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get order items", zap.Int64("orderID", id), zap.Error(err))
		return nil, err
	}
	order.Items = items

	// Get shipping info
	shippingInfo, err := u.orderRepo.GetShippingInfo(ctx, id)
	if err != nil && !errors.Is(err, pkgerrors.ErrNotFound) {
		u.logger.Error("Failed to get shipping info", zap.Int64("orderID", id), zap.Error(err))
		return nil, err
	}
	if shippingInfo != nil {
		order.ShippingInfo = *shippingInfo
	}

	return order, nil
}
