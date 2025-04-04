package usecase

import (
	"context"
	"time"

	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// userUseCase implements domain.UserUseCase
type userUseCase struct {
	userRepo   domain.UserRepository
	jwtService domain.JWTService
	logger     logger.Logger
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(userRepo domain.UserRepository, jwtService domain.JWTService, logger logger.Logger) domain.UserUseCase {
	return &userUseCase{
		userRepo:   userRepo,
		jwtService: jwtService,
		logger:     logger,
	}
}

// GetByID gets a user by ID
func (u *userUseCase) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get user by ID", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}
	return user, nil
}

// Create creates a new user
func (u *userUseCase) Create(ctx context.Context, user *domain.User) error {
	// Check if user with the same email already exists
	existingUser, err := u.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return &domain.ConflictError{
			Entity: "User",
			Field:  "email",
			Value:  user.Email,
		}
	}

	// Check if user with the same username already exists
	existingUser, err = u.userRepo.GetByUsername(ctx, user.Username)
	if err == nil && existingUser != nil {
		return &domain.ConflictError{
			Entity: "User",
			Field:  "username",
			Value:  user.Username,
		}
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		u.logger.Error("Failed to hash password", zap.Error(err))
		return domain.ErrInternalServer
	}
	user.Password = string(hashedPassword)

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Create the user
	if err := u.userRepo.Create(ctx, user); err != nil {
		u.logger.Error("Failed to create user", zap.Error(err))
		return err
	}

	return nil
}

// Update updates a user
func (u *userUseCase) Update(ctx context.Context, user *domain.User) error {
	// Get the existing user
	existingUser, err := u.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		u.logger.Error("Failed to get user for update", zap.Int64("id", user.ID), zap.Error(err))
		return err
	}

	// Check if email is being changed and if it's already taken
	if user.Email != existingUser.Email {
		emailUser, err := u.userRepo.GetByEmail(ctx, user.Email)
		if err == nil && emailUser != nil && emailUser.ID != user.ID {
			return &domain.ConflictError{
				Entity: "User",
				Field:  "email",
				Value:  user.Email,
			}
		}
	}

	// Check if username is being changed and if it's already taken
	if user.Username != existingUser.Username {
		usernameUser, err := u.userRepo.GetByUsername(ctx, user.Username)
		if err == nil && usernameUser != nil && usernameUser.ID != user.ID {
			return &domain.ConflictError{
				Entity: "User",
				Field:  "username",
				Value:  user.Username,
			}
		}
	}

	// If password is provided, hash it
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			u.logger.Error("Failed to hash password for update", zap.Error(err))
			return domain.ErrInternalServer
		}
		user.Password = string(hashedPassword)
	} else {
		// Keep the existing password
		user.Password = existingUser.Password
	}

	// Update timestamp
	user.UpdatedAt = time.Now()
	user.CreatedAt = existingUser.CreatedAt

	// Update the user
	if err := u.userRepo.Update(ctx, user); err != nil {
		u.logger.Error("Failed to update user", zap.Int64("id", user.ID), zap.Error(err))
		return err
	}

	return nil
}

// Delete deletes a user
func (u *userUseCase) Delete(ctx context.Context, id int64) error {
	// Check if user exists
	_, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get user for deletion", zap.Int64("id", id), zap.Error(err))
		return err
	}

	// Delete the user
	if err := u.userRepo.Delete(ctx, id); err != nil {
		u.logger.Error("Failed to delete user", zap.Int64("id", id), zap.Error(err))
		return err
	}

	return nil
}

// List lists users with pagination
func (u *userUseCase) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	users, err := u.userRepo.List(ctx, limit, offset)
	if err != nil {
		u.logger.Error("Failed to list users", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, err
	}
	return users, nil
}

// Login authenticates a user and returns a JWT token
func (u *userUseCase) Login(ctx context.Context, email, password string) (string, error) {
	// Get the user by email
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		u.logger.Error("Failed to get user by email", zap.String("email", email), zap.Error(err))
		return "", domain.ErrUnauthorized
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		u.logger.Error("Invalid password", zap.String("email", email), zap.Error(err))
		return "", domain.ErrUnauthorized
	}

	// Generate token
	token, err := u.jwtService.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		u.logger.Error("Failed to generate token", zap.String("email", email), zap.Error(err))
		return "", domain.ErrInternalServer
	}

	return token, nil
}

// Register registers a new user
func (u *userUseCase) Register(ctx context.Context, user *domain.User) error {
	// Set default role if not provided
	if user.Role == "" {
		user.Role = domain.RoleUser
	}

	// Create the user
	return u.Create(ctx, user)
}

// ValidateToken validates a JWT token and returns the user
func (u *userUseCase) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	// Validate the token
	claims, err := u.jwtService.ValidateToken(token)
	if err != nil {
		u.logger.Error("Failed to validate token", zap.Error(err))
		return nil, domain.ErrUnauthorized
	}

	// Get the user by ID
	user, err := u.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		u.logger.Error("Failed to get user by ID from token", zap.Int64("id", claims.UserID), zap.Error(err))
		return nil, domain.ErrUnauthorized
	}

	return user, nil
}
