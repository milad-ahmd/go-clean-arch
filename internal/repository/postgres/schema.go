package postgres

import (
	"database/sql"

	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"go.uber.org/zap"
)

// CreateTables creates all necessary tables if they don't exist
func CreateTables(db *sql.DB, logger logger.Logger) error {
	// Create users table
	usersTable := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password VARCHAR(100) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'user',
			created_at BIGINT NOT NULL,
			updated_at BIGINT NOT NULL
		);
	`

	// Create categories table
	categoriesTable := `
		CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) UNIQUE NOT NULL,
			description TEXT,
			slug VARCHAR(100) UNIQUE NOT NULL,
			created_at BIGINT NOT NULL,
			updated_at BIGINT NOT NULL
		);
	`

	// Create products table
	productsTable := `
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			price DECIMAL(10, 2) NOT NULL,
			sku VARCHAR(50) UNIQUE NOT NULL,
			stock INT NOT NULL DEFAULT 0,
			category_id INT NOT NULL REFERENCES categories(id),
			images JSONB DEFAULT '[]',
			created_at BIGINT NOT NULL,
			updated_at BIGINT NOT NULL
		);
	`

	// Create orders table
	ordersTable := `
		CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(id),
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			total_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
			payment_method VARCHAR(20) NOT NULL,
			created_at BIGINT NOT NULL,
			updated_at BIGINT NOT NULL
		);
	`

	// Create order_items table
	orderItemsTable := `
		CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			product_id INT NOT NULL REFERENCES products(id),
			quantity INT NOT NULL,
			price DECIMAL(10, 2) NOT NULL,
			created_at BIGINT NOT NULL,
			updated_at BIGINT NOT NULL
		);
	`

	// Create shipping_info table
	shippingInfoTable := `
		CREATE TABLE IF NOT EXISTS shipping_info (
			id SERIAL PRIMARY KEY,
			order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			address TEXT NOT NULL,
			city VARCHAR(100) NOT NULL,
			state VARCHAR(100) NOT NULL,
			country VARCHAR(100) NOT NULL,
			postal_code VARCHAR(20) NOT NULL,
			phone_number VARCHAR(20) NOT NULL,
			created_at BIGINT NOT NULL,
			updated_at BIGINT NOT NULL,
			UNIQUE(order_id)
		);
	`

	// Execute all table creation queries
	tables := []string{
		usersTable,
		categoriesTable,
		productsTable,
		ordersTable,
		orderItemsTable,
		shippingInfoTable,
	}

	for _, table := range tables {
		_, err := db.Exec(table)
		if err != nil {
			logger.Error("Failed to create table", zap.Error(err))
			return err
		}
	}

	logger.Info("Database tables created successfully")
	return nil
}
