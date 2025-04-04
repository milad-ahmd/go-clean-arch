package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	"go.uber.org/zap/zapcore"
)

// mockUserRepository is a mock implementation of domain.UserRepository
type mockUserRepository struct {
	users map[int64]*domain.User
}

// newMockUserRepository creates a new mock user repository
func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[int64]*domain.User),
	}
}

// GetByID gets a user by ID
func (m *mockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, &domain.NotFoundError{
			Entity: "User",
			ID:     id,
		}
	}
	return user, nil
}

// GetByEmail gets a user by email
func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, &domain.NotFoundError{
		Entity: "User",
		ID:     email,
	}
}

// GetByUsername gets a user by username
func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, &domain.NotFoundError{
		Entity: "User",
		ID:     username,
	}
}

// Create creates a new user
func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	user.ID = int64(len(m.users) + 1)
	m.users[user.ID] = user
	return nil
}

// Update updates a user
func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	_, ok := m.users[user.ID]
	if !ok {
		return &domain.NotFoundError{
			Entity: "User",
			ID:     user.ID,
		}
	}
	m.users[user.ID] = user
	return nil
}

// Delete deletes a user
func (m *mockUserRepository) Delete(ctx context.Context, id int64) error {
	_, ok := m.users[id]
	if !ok {
		return &domain.NotFoundError{
			Entity: "User",
			ID:     id,
		}
	}
	delete(m.users, id)
	return nil
}

// List lists users with pagination
func (m *mockUserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	count := 0
	for _, user := range m.users {
		if count >= offset && count < offset+limit {
			users = append(users, user)
		}
		count++
	}
	return users, nil
}

// mockLogger is a mock implementation of logger.Logger
type mockLogger struct{}

// Debug logs a debug message
func (m *mockLogger) Debug(_ string, _ ...zapcore.Field) {}

// Info logs an info message
func (m *mockLogger) Info(_ string, _ ...zapcore.Field) {}

// Warn logs a warning message
func (m *mockLogger) Warn(_ string, _ ...zapcore.Field) {}

// Error logs an error message
func (m *mockLogger) Error(_ string, _ ...zapcore.Field) {}

// Fatal logs a fatal message
func (m *mockLogger) Fatal(_ string, _ ...zapcore.Field) {}

// mockJWTService is a mock implementation of domain.JWTService
type mockJWTService struct{}

// GenerateToken generates a mock token
func (m *mockJWTService) GenerateToken(_ int64, _ string, _ domain.Role) (string, error) {
	return "mock-token", nil
}

// ValidateToken validates a mock token
func (m *mockJWTService) ValidateToken(_ string) (*domain.JWTClaims, error) {
	return &domain.JWTClaims{
		UserID:   1,
		Username: "testuser",
		Role:     domain.RoleUser,
	}, nil
}

// TestUserUseCase_GetByID tests the GetByID method
func TestUserUseCase_GetByID(t *testing.T) {
	// Create a mock repository
	repo := newMockUserRepository()

	// Create a mock logger
	logger := &mockLogger{}

	// Create a mock JWT service
	jwtService := &mockJWTService{}

	// Create a user use case
	useCase := NewUserUseCase(repo, jwtService, logger)

	// Create a test user
	user := &domain.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add the user to the repository
	repo.users[user.ID] = user

	// Test getting a user by ID
	ctx := context.Background()
	result, err := useCase.GetByID(ctx, user.ID)

	// Check the result
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Error("Expected a user, got nil")
		return
	}

	if result.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, result.ID)
	}

	if result.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, result.Username)
	}

	if result.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, result.Email)
	}

	// Test getting a non-existent user
	result, err = useCase.GetByID(ctx, 999)

	// Check the result
	if err == nil {
		t.Error("Expected an error, got nil")
	}

	var notFoundErr *domain.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("Expected a NotFoundError, got %T", err)
	}

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

// TestUserUseCase_Create tests the Create method
func TestUserUseCase_Create(t *testing.T) {
	// Create a mock repository
	repo := newMockUserRepository()

	// Create a mock logger
	logger := &mockLogger{}

	// Create a mock JWT service
	jwtService := &mockJWTService{}

	// Create a user use case
	useCase := NewUserUseCase(repo, jwtService, logger)

	// Create a test user
	user := &domain.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password",
	}

	// Test creating a user
	ctx := context.Background()
	err := useCase.Create(ctx, user)

	// Check the result
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected a non-zero ID")
	}

	if user.CreatedAt.IsZero() {
		t.Error("Expected a non-zero CreatedAt")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("Expected a non-zero UpdatedAt")
	}

	// Test creating a user with a duplicate email
	duplicateUser := &domain.User{
		Username: "anotheruser",
		Email:    "test@example.com", // Same email as the first user
		Password: "password",
	}

	err = useCase.Create(ctx, duplicateUser)

	// Check the result
	if err == nil {
		t.Error("Expected an error, got nil")
	}

	var conflictErr *domain.ConflictError
	if !errors.As(err, &conflictErr) {
		t.Errorf("Expected a ConflictError, got %T", err)
	}
}
