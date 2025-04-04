package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"go.uber.org/zap"
)

// userRepository implements domain.UserRepository
type userRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB, logger logger.Logger) domain.UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// GetByID gets a user by ID
func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, username, email, password, role, created_at, updated_at FROM users WHERE id = $1`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &domain.NotFoundError{
				Entity: "User",
				ID:     id,
			}
		}
		r.logger.Error("Failed to get user by ID", zap.Int64("id", id), zap.Error(err))
		return nil, domain.ErrInternalServer
	}

	return &user, nil
}

// GetByEmail gets a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, username, email, password, role, created_at, updated_at FROM users WHERE email = $1`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &domain.NotFoundError{
				Entity: "User",
				ID:     email,
			}
		}
		r.logger.Error("Failed to get user by email", zap.String("email", email), zap.Error(err))
		return nil, domain.ErrInternalServer
	}

	return &user, nil
}

// GetByUsername gets a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, username, email, password, role, created_at, updated_at FROM users WHERE username = $1`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &domain.NotFoundError{
				Entity: "User",
				ID:     username,
			}
		}
		r.logger.Error("Failed to get user by username", zap.String("username", username), zap.Error(err))
		return nil, domain.ErrInternalServer
	}

	return &user, nil
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (username, email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		r.logger.Error("Failed to create user", zap.Error(err))
		return domain.ErrInternalServer
	}

	return nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, password = $3, role = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update user", zap.Int64("id", user.ID), zap.Error(err))
		return domain.ErrInternalServer
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return domain.ErrInternalServer
	}

	if rowsAffected == 0 {
		return &domain.NotFoundError{
			Entity: "User",
			ID:     user.ID,
		}
	}

	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete user", zap.Int64("id", id), zap.Error(err))
		return domain.ErrInternalServer
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return domain.ErrInternalServer
	}

	if rowsAffected == 0 {
		return &domain.NotFoundError{
			Entity: "User",
			ID:     id,
		}
	}

	return nil
}

// List lists users with pagination
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	query := `
		SELECT id, username, email, password, role, created_at, updated_at
		FROM users
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("Failed to list users", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, domain.ErrInternalServer
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan user", zap.Error(err))
			return nil, domain.ErrInternalServer
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating user rows", zap.Error(err))
		return nil, domain.ErrInternalServer
	}

	return users, nil
}
