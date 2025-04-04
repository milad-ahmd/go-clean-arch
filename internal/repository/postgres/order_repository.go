package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	pkgerrors "github.com/milad-ahmd/go-clean-arch/pkg/errors"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"go.uber.org/zap"
)

type orderRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *sql.DB, logger logger.Logger) domain.OrderRepository {
	return &orderRepository{
		db:     db,
		logger: logger,
	}
}

// FindByID finds an order by ID
func (r *orderRepository) FindByID(ctx context.Context, id int64) (*domain.Order, error) {
	query := `
		SELECT o.id, o.user_id, o.status, o.total_amount, o.payment_method, o.created_at, o.updated_at,
			   u.id, u.username, u.email, u.role, u.created_at, u.updated_at
		FROM orders o
		LEFT JOIN users u ON o.user_id = u.id
		WHERE o.id = $1
	`

	var order domain.Order
	var user domain.User

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.TotalAmount,
		&order.PaymentMethod,
		&order.CreatedAt,
		&order.UpdatedAt,
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.NewNotFoundError("Order", id)
		}
		r.logger.Error("Failed to find order by ID", zap.Int64("id", id), zap.Error(err))
		return nil, pkgerrors.NewInternalError(err)
	}

	order.User = user

	// Get order items
	items, err := r.GetOrderItems(ctx, order.ID)
	if err != nil {
		r.logger.Error("Failed to get order items", zap.Int64("orderID", order.ID), zap.Error(err))
		return nil, err
	}
	order.Items = items

	// Get shipping info
	shippingInfo, err := r.GetShippingInfo(ctx, order.ID)
	if err != nil && !errors.Is(err, pkgerrors.ErrNotFound) {
		r.logger.Error("Failed to get shipping info", zap.Int64("orderID", order.ID), zap.Error(err))
		return nil, err
	}
	if shippingInfo != nil {
		order.ShippingInfo = *shippingInfo
	}

	return &order, nil
}

// FindAll finds all orders with pagination
func (r *orderRepository) FindAll(ctx context.Context, limit, offset int) ([]domain.Order, int, error) {
	query := `
		SELECT o.id, o.user_id, o.status, o.total_amount, o.payment_method, o.created_at, o.updated_at,
			   u.id, u.username, u.email, u.role, u.created_at, u.updated_at
		FROM orders o
		LEFT JOIN users u ON o.user_id = u.id
		ORDER BY o.id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("Failed to find all orders", zap.Error(err))
		return nil, 0, pkgerrors.NewInternalError(err)
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var order domain.Order
		var user domain.User

		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.TotalAmount,
			&order.PaymentMethod,
			&order.CreatedAt,
			&order.UpdatedAt,
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan order", zap.Error(err))
			return nil, 0, pkgerrors.NewInternalError(err)
		}

		order.User = user
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating order rows", zap.Error(err))
		return nil, 0, pkgerrors.NewInternalError(err)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM orders`
	err = r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to get total order count", zap.Error(err))
		return nil, 0, pkgerrors.NewInternalError(err)
	}

	return orders, total, nil
}

// Create creates a new order
func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
		INSERT INTO orders (user_id, status, total_amount, payment_method, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	now := time.Now().Unix()
	order.CreatedAt = now
	order.UpdatedAt = now

	err = tx.QueryRowContext(
		ctx,
		query,
		order.UserID,
		order.Status,
		order.TotalAmount,
		order.PaymentMethod,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(&order.ID)

	if err != nil {
		r.logger.Error("Failed to create order", zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	// Insert order items
	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		order.Items[i].CreatedAt = now
		order.Items[i].UpdatedAt = now

		itemQuery := `
			INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`

		err = tx.QueryRowContext(
			ctx,
			itemQuery,
			order.Items[i].OrderID,
			order.Items[i].ProductID,
			order.Items[i].Quantity,
			order.Items[i].Price,
			order.Items[i].CreatedAt,
			order.Items[i].UpdatedAt,
		).Scan(&order.Items[i].ID)

		if err != nil {
			r.logger.Error("Failed to create order item", zap.Error(err))
			return pkgerrors.NewInternalError(err)
		}

		// Update product stock
		stockQuery := `
			UPDATE products
			SET stock = stock - $1, updated_at = $2
			WHERE id = $3
			RETURNING stock
		`

		var newStock int
		err = tx.QueryRowContext(
			ctx,
			stockQuery,
			order.Items[i].Quantity,
			now,
			order.Items[i].ProductID,
		).Scan(&newStock)

		if err != nil {
			r.logger.Error("Failed to update product stock", zap.Error(err))
			return pkgerrors.NewInternalError(err)
		}

		if newStock < 0 {
			return pkgerrors.NewBadRequestError("Insufficient stock for product ID: " + fmt.Sprintf("%d", order.Items[i].ProductID))
		}
	}

	// Insert shipping info if provided
	if order.ShippingInfo.Address != "" {
		order.ShippingInfo.OrderID = order.ID
		order.ShippingInfo.CreatedAt = now
		order.ShippingInfo.UpdatedAt = now

		shippingQuery := `
			INSERT INTO shipping_info (order_id, address, city, state, country, postal_code, phone_number, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`

		err = tx.QueryRowContext(
			ctx,
			shippingQuery,
			order.ShippingInfo.OrderID,
			order.ShippingInfo.Address,
			order.ShippingInfo.City,
			order.ShippingInfo.State,
			order.ShippingInfo.Country,
			order.ShippingInfo.PostalCode,
			order.ShippingInfo.PhoneNumber,
			order.ShippingInfo.CreatedAt,
			order.ShippingInfo.UpdatedAt,
		).Scan(&order.ShippingInfo.ID)

		if err != nil {
			r.logger.Error("Failed to create shipping info", zap.Error(err))
			return pkgerrors.NewInternalError(err)
		}
	}

	if err = tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	return nil
}

// Update updates an order
func (r *orderRepository) Update(ctx context.Context, order *domain.Order) error {
	query := `
		UPDATE orders
		SET status = $1, payment_method = $2, updated_at = $3
		WHERE id = $4
	`

	order.UpdatedAt = time.Now().Unix()

	result, err := r.db.ExecContext(
		ctx,
		query,
		order.Status,
		order.PaymentMethod,
		order.UpdatedAt,
		order.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update order", zap.Int64("id", order.ID), zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	if rowsAffected == 0 {
		return pkgerrors.NewNotFoundError("Order", order.ID)
	}

	// Update shipping info if provided
	if order.ShippingInfo.Address != "" {
		shippingQuery := `
			UPDATE shipping_info
			SET address = $1, city = $2, state = $3, country = $4, postal_code = $5, phone_number = $6, updated_at = $7
			WHERE order_id = $8
		`

		_, err = r.db.ExecContext(
			ctx,
			shippingQuery,
			order.ShippingInfo.Address,
			order.ShippingInfo.City,
			order.ShippingInfo.State,
			order.ShippingInfo.Country,
			order.ShippingInfo.PostalCode,
			order.ShippingInfo.PhoneNumber,
			order.UpdatedAt,
			order.ID,
		)

		if err != nil {
			r.logger.Error("Failed to update shipping info", zap.Int64("orderID", order.ID), zap.Error(err))
			return pkgerrors.NewInternalError(err)
		}
	}

	return nil
}

// Delete deletes an order
func (r *orderRepository) Delete(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Delete shipping info
	_, err = tx.ExecContext(ctx, `DELETE FROM shipping_info WHERE order_id = $1`, id)
	if err != nil {
		r.logger.Error("Failed to delete shipping info", zap.Int64("orderID", id), zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	// Delete order items
	_, err = tx.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, id)
	if err != nil {
		r.logger.Error("Failed to delete order items", zap.Int64("orderID", id), zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	// Delete order
	result, err := tx.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		r.logger.Error("Failed to delete order", zap.Int64("id", id), zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	if rowsAffected == 0 {
		return pkgerrors.NewNotFoundError("Order", id)
	}

	if err = tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	return nil
}

// FindByUserID finds orders by user ID
func (r *orderRepository) FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.Order, int, error) {
	query := `
		SELECT o.id, o.user_id, o.status, o.total_amount, o.payment_method, o.created_at, o.updated_at,
			   u.id, u.username, u.email, u.role, u.created_at, u.updated_at
		FROM orders o
		LEFT JOIN users u ON o.user_id = u.id
		WHERE o.user_id = $1
		ORDER BY o.id
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to find orders by user ID", zap.Int64("userID", userID), zap.Error(err))
		return nil, 0, pkgerrors.NewInternalError(err)
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var order domain.Order
		var user domain.User

		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.TotalAmount,
			&order.PaymentMethod,
			&order.CreatedAt,
			&order.UpdatedAt,
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan order", zap.Error(err))
			return nil, 0, pkgerrors.NewInternalError(err)
		}

		order.User = user
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating order rows", zap.Error(err))
		return nil, 0, pkgerrors.NewInternalError(err)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM orders WHERE user_id = $1`
	err = r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to get total order count by user ID", zap.Int64("userID", userID), zap.Error(err))
		return nil, 0, pkgerrors.NewInternalError(err)
	}

	return orders, total, nil
}

// FindByStatus finds orders by status
func (r *orderRepository) FindByStatus(ctx context.Context, status domain.OrderStatus, limit, offset int) ([]domain.Order, int, error) {
	query := `
		SELECT o.id, o.user_id, o.status, o.total_amount, o.payment_method, o.created_at, o.updated_at,
			   u.id, u.username, u.email, u.role, u.created_at, u.updated_at
		FROM orders o
		LEFT JOIN users u ON o.user_id = u.id
		WHERE o.status = $1
		ORDER BY o.id
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		r.logger.Error("Failed to find orders by status", zap.String("status", string(status)), zap.Error(err))
		return nil, 0, pkgerrors.NewInternalError(err)
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var order domain.Order
		var user domain.User

		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.TotalAmount,
			&order.PaymentMethod,
			&order.CreatedAt,
			&order.UpdatedAt,
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan order", zap.Error(err))
			return nil, 0, pkgerrors.NewInternalError(err)
		}

		order.User = user
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating order rows", zap.Error(err))
		return nil, 0, pkgerrors.NewInternalError(err)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM orders WHERE status = $1`
	err = r.db.QueryRowContext(ctx, countQuery, status).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to get total order count by status", zap.String("status", string(status)), zap.Error(err))
		return nil, 0, pkgerrors.NewInternalError(err)
	}

	return orders, total, nil
}

// UpdateStatus updates an order's status
func (r *orderRepository) UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now().Unix()

	result, err := r.db.ExecContext(ctx, query, status, now, id)
	if err != nil {
		r.logger.Error("Failed to update order status", zap.Int64("id", id), zap.String("status", string(status)), zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	if rowsAffected == 0 {
		return pkgerrors.NewNotFoundError("Order", id)
	}

	return nil
}

// AddOrderItem adds an item to an order
func (r *orderRepository) AddOrderItem(ctx context.Context, item *domain.OrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Check if order exists
	var exists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1)`, item.OrderID).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check if order exists", zap.Int64("orderID", item.OrderID), zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	if !exists {
		return pkgerrors.NewNotFoundError("Order", item.OrderID)
	}

	// Check if product exists and has enough stock
	var stock int
	err = tx.QueryRowContext(ctx, `SELECT stock FROM products WHERE id = $1`, item.ProductID).Scan(&stock)
	if err != nil {
		if err == sql.ErrNoRows {
			return pkgerrors.NewNotFoundError("Product", item.ProductID)
		}
		r.logger.Error("Failed to get product stock", zap.Int64("productID", item.ProductID), zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	if stock < item.Quantity {
		return pkgerrors.NewBadRequestError("Insufficient stock")
	}

	// Insert order item
	now := time.Now().Unix()
	item.CreatedAt = now
	item.UpdatedAt = now

	query := `
		INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err = tx.QueryRowContext(
		ctx,
		query,
		item.OrderID,
		item.ProductID,
		item.Quantity,
		item.Price,
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&item.ID)

	if err != nil {
		r.logger.Error("Failed to add order item", zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	// Update product stock
	_, err = tx.ExecContext(
		ctx,
		`UPDATE products SET stock = stock - $1, updated_at = $2 WHERE id = $3`,
		item.Quantity,
		now,
		item.ProductID,
	)

	if err != nil {
		r.logger.Error("Failed to update product stock", zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	// Update order total amount
	_, err = tx.ExecContext(
		ctx,
		`UPDATE orders SET total_amount = total_amount + $1, updated_at = $2 WHERE id = $3`,
		item.Price*float64(item.Quantity),
		now,
		item.OrderID,
	)

	if err != nil {
		r.logger.Error("Failed to update order total amount", zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	if err = tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	return nil
}

// GetOrderItems gets all items for an order
func (r *orderRepository) GetOrderItems(ctx context.Context, orderID int64) ([]domain.OrderItem, error) {
	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.created_at, oi.updated_at,
			   p.id, p.name, p.description, p.price, p.sku, p.stock, p.category_id, p.created_at, p.updated_at
		FROM order_items oi
		LEFT JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		r.logger.Error("Failed to get order items", zap.Int64("orderID", orderID), zap.Error(err))
		return nil, pkgerrors.NewInternalError(err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		var product domain.Product

		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
			&item.CreatedAt,
			&item.UpdatedAt,
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.SKU,
			&product.Stock,
			&product.CategoryID,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan order item", zap.Error(err))
			return nil, pkgerrors.NewInternalError(err)
		}

		item.Product = product
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating order item rows", zap.Error(err))
		return nil, pkgerrors.NewInternalError(err)
	}

	return items, nil
}

// SaveShippingInfo saves shipping information for an order
func (r *orderRepository) SaveShippingInfo(ctx context.Context, info *domain.ShippingInfo) error {
	// Check if shipping info already exists for this order
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM shipping_info WHERE order_id = $1)`, info.OrderID).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check if shipping info exists", zap.Int64("orderID", info.OrderID), zap.Error(err))
		return pkgerrors.NewInternalError(err)
	}

	now := time.Now().Unix()
	info.UpdatedAt = now

	if exists {
		// Update existing shipping info
		query := `
			UPDATE shipping_info
			SET address = $1, city = $2, state = $3, country = $4, postal_code = $5, phone_number = $6, updated_at = $7
			WHERE order_id = $8
		`

		_, err = r.db.ExecContext(
			ctx,
			query,
			info.Address,
			info.City,
			info.State,
			info.Country,
			info.PostalCode,
			info.PhoneNumber,
			info.UpdatedAt,
			info.OrderID,
		)

		if err != nil {
			r.logger.Error("Failed to update shipping info", zap.Int64("orderID", info.OrderID), zap.Error(err))
			return pkgerrors.NewInternalError(err)
		}
	} else {
		// Insert new shipping info
		query := `
			INSERT INTO shipping_info (order_id, address, city, state, country, postal_code, phone_number, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`

		info.CreatedAt = now

		err = r.db.QueryRowContext(
			ctx,
			query,
			info.OrderID,
			info.Address,
			info.City,
			info.State,
			info.Country,
			info.PostalCode,
			info.PhoneNumber,
			info.CreatedAt,
			info.UpdatedAt,
		).Scan(&info.ID)

		if err != nil {
			r.logger.Error("Failed to create shipping info", zap.Error(err))
			return pkgerrors.NewInternalError(err)
		}
	}

	return nil
}

// GetShippingInfo gets shipping information for an order
func (r *orderRepository) GetShippingInfo(ctx context.Context, orderID int64) (*domain.ShippingInfo, error) {
	query := `
		SELECT id, order_id, address, city, state, country, postal_code, phone_number, created_at, updated_at
		FROM shipping_info
		WHERE order_id = $1
	`

	var info domain.ShippingInfo
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&info.ID,
		&info.OrderID,
		&info.Address,
		&info.City,
		&info.State,
		&info.Country,
		&info.PostalCode,
		&info.PhoneNumber,
		&info.CreatedAt,
		&info.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.NewNotFoundError("ShippingInfo", orderID)
		}
		r.logger.Error("Failed to get shipping info", zap.Int64("orderID", orderID), zap.Error(err))
		return nil, pkgerrors.NewInternalError(err)
	}

	return &info, nil
}
