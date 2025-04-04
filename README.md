# Go Clean Architecture

A backend Go project implementing clean architecture principles. This project serves as a template and showcase for building maintainable, testable, and scalable Go applications.

## Architecture Overview

This project follows the clean architecture principles proposed by Robert C. Martin (Uncle Bob), with the following layers:

1. **Domain Layer**: Core business logic and entities
2. **Use Case Layer**: Application-specific business rules
3. **Interface/Delivery Layer**: Adapters for external interfaces (HTTP, DB, etc.)
4. **Infrastructure Layer**: External frameworks and tools

## Project Structure

```
.
├── cmd/
│   └── api/                  # Application entry points
│       └── full/             # Full version with all modules
├── internal/
│   ├── domain/               # Enterprise business rules (entities)
│   ├── usecase/              # Application business rules
│   ├── delivery/             # Interface adapters (controllers, presenters)
│   │   └── http/             # HTTP handlers
│   └── repository/           # Data access implementations
│       └── postgres/         # PostgreSQL implementations
└── pkg/
    ├── auth/                 # Authentication utilities
    ├── config/               # Configuration management
    ├── errors/               # Error handling
    ├── logger/               # Logging utilities
    ├── middleware/           # HTTP middleware
    ├── response/             # Standardized API responses
    └── swagger/              # Swagger documentation
```

## Features

- Clean architecture implementation
- RESTful API endpoints
- Authentication and Authorization
- Dependency injection
- Standardized error handling
- Centralized logging
- Configuration management with .env file support
- Database integration (PostgreSQL)
- Swagger documentation
- Unit and integration tests

## Modules

The project includes the following modules:

- **User**: User management and authentication
- **Category**: Product category management
- **Product**: Product management with category relationships
- **Order**: Order management with product and user relationships

## Getting Started

### Prerequisites

- Go 1.18 or higher
- PostgreSQL

### Installation

1. Clone the repository

```bash
git clone https://github.com/milad-ahmd/go-clean-arch.git
cd go-clean-arch
```

2. Install dependencies

```bash
go mod tidy
```

3. Configure environment variables

You can configure the application using environment variables or a `.env` file. Create a `.env` file in the root directory with the following variables:

```
# Server Configuration
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10
SERVER_WRITE_TIMEOUT=10
SERVER_IDLE_TIMEOUT=120

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=clean_arch
DB_SSL_MODE=disable

# Logger Configuration
LOG_LEVEL=info

# Authentication Configuration
JWT_SECRET=your-secret-key-change-in-production
```

4. Run the application

```bash
# Run with basic modules (user and authentication)
go run cmd/api/main.go

# Run with all modules (user, category, product, order)
go run cmd/api/full/main.go
```

## API Documentation

API documentation is available at `/swagger/index.html` when the server is running.

## API Endpoints

### Authentication

- `POST /auth/register`: Register a new user
- `POST /auth/login`: Login and get JWT token

### Users

- `GET /users`: List users
- `GET /users/{id}`: Get user by ID
- `PUT /users/{id}`: Update user
- `DELETE /users/{id}`: Delete user

### Categories

- `GET /categories`: List categories
- `GET /categories/{id}`: Get category by ID
- `POST /categories`: Create category
- `PUT /categories/{id}`: Update category
- `DELETE /categories/{id}`: Delete category
- `GET /categories/slug/{slug}`: Get category by slug

### Products

- `GET /products`: List products
- `GET /products/{id}`: Get product by ID
- `POST /products`: Create product
- `PUT /products/{id}`: Update product
- `DELETE /products/{id}`: Delete product
- `GET /products/sku/{sku}`: Get product by SKU
- `GET /products/category/{categoryID}`: Get products by category
- `GET /products/search`: Search products
- `PATCH /products/{id}/stock`: Update product stock

### Orders

- `GET /orders`: List orders
- `GET /orders/{id}`: Get order by ID
- `POST /orders`: Create order
- `PUT /orders/{id}`: Update order
- `DELETE /orders/{id}`: Delete order
- `PATCH /orders/{id}/status`: Update order status
- `GET /orders/user/{userID}`: Get orders by user
- `GET /orders/status/{status}`: Get orders by status

## License

This project is licensed under the MIT License - see the LICENSE file for details.
