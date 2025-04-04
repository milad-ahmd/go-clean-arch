package domain

import (
	"context"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	// OrderStatusPending represents a pending order
	OrderStatusPending OrderStatus = "pending"

	// OrderStatusProcessing represents a processing order
	OrderStatusProcessing OrderStatus = "processing"

	// OrderStatusCompleted represents a completed order
	OrderStatusCompleted OrderStatus = "completed"

	// OrderStatusCancelled represents a cancelled order
	OrderStatusCancelled OrderStatus = "cancelled"
)

// PaymentMethod represents the payment method
type PaymentMethod string

const (
	// PaymentMethodCreditCard represents a credit card payment
	PaymentMethodCreditCard PaymentMethod = "credit_card"

	// PaymentMethodPayPal represents a PayPal payment
	PaymentMethodPayPal PaymentMethod = "paypal"

	// PaymentMethodBankTransfer represents a bank transfer payment
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)

// OrderItem represents an item in an order
type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Product   Product `json:"product,omitempty"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	BaseEntity
}

// Order represents an order entity
type Order struct {
	ID            int64         `json:"id"`
	UserID        int64         `json:"user_id"`
	User          User          `json:"user,omitempty"`
	Status        OrderStatus   `json:"status"`
	TotalAmount   float64       `json:"total_amount"`
	Items         []OrderItem   `json:"items,omitempty"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	ShippingInfo  ShippingInfo  `json:"shipping_info,omitempty"`
	BaseEntity
}

// ShippingInfo represents shipping information
type ShippingInfo struct {
	ID          int64  `json:"id"`
	OrderID     int64  `json:"order_id"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	PostalCode  string `json:"postal_code"`
	PhoneNumber string `json:"phone_number"`
	BaseEntity
}

// OrderRepository defines the order repository interface
type OrderRepository interface {
	BaseRepository[Order, int64]
	FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]Order, int, error)
	FindByStatus(ctx context.Context, status OrderStatus, limit, offset int) ([]Order, int, error)
	UpdateStatus(ctx context.Context, id int64, status OrderStatus) error
	AddOrderItem(ctx context.Context, item *OrderItem) error
	GetOrderItems(ctx context.Context, orderID int64) ([]OrderItem, error)
	SaveShippingInfo(ctx context.Context, info *ShippingInfo) error
	GetShippingInfo(ctx context.Context, orderID int64) (*ShippingInfo, error)
}

// OrderItemCreateDTO represents the data for creating an order item
type OrderItemCreateDTO struct {
	ProductID int64   `json:"product_id" validate:"required,gt=0"`
	Quantity  int     `json:"quantity" validate:"required,gt=0"`
	Price     float64 `json:"price" validate:"required,gt=0"`
}

// ShippingInfoDTO represents the data for shipping information
type ShippingInfoDTO struct {
	Address     string `json:"address" validate:"required"`
	City        string `json:"city" validate:"required"`
	State       string `json:"state" validate:"required"`
	Country     string `json:"country" validate:"required"`
	PostalCode  string `json:"postal_code" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"required"`
}

// OrderCreateDTO represents the data for creating an order
type OrderCreateDTO struct {
	UserID        int64                `json:"user_id" validate:"required,gt=0"`
	Items         []OrderItemCreateDTO `json:"items" validate:"required,dive"`
	PaymentMethod PaymentMethod        `json:"payment_method" validate:"required,oneof=credit_card paypal bank_transfer"`
	ShippingInfo  ShippingInfoDTO      `json:"shipping_info" validate:"required"`
}

// OrderUpdateDTO represents the data for updating an order
type OrderUpdateDTO struct {
	Status        OrderStatus     `json:"status" validate:"omitempty,oneof=pending processing completed cancelled"`
	PaymentMethod PaymentMethod   `json:"payment_method" validate:"omitempty,oneof=credit_card paypal bank_transfer"`
	ShippingInfo  ShippingInfoDTO `json:"shipping_info" validate:"omitempty"`
}

// OrderUseCase defines the order use case interface
type OrderUseCase interface {
	BaseUseCase[Order, int64, OrderCreateDTO, OrderUpdateDTO]
	GetByUserID(ctx context.Context, userID int64, page, perPage int) ([]Order, int, error)
	GetByStatus(ctx context.Context, status OrderStatus, page, perPage int) ([]Order, int, error)
	UpdateStatus(ctx context.Context, id int64, status OrderStatus) error
	GetOrderWithDetails(ctx context.Context, id int64) (*Order, error)
}
