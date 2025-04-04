package domain

import (
	"context"
	"time"
)

// Role represents user role
type Role string

// Available roles
const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// User represents the user entity
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Password is not exposed in JSON
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRepository represents the user repository contract
type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
}

// UserUseCase represents the user use case contract
type UserUseCase interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, user *User) error
	ValidateToken(ctx context.Context, token string) (*User, error)
}
