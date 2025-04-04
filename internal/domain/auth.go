package domain

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest represents the register request
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// TokenResponse represents the token response
type TokenResponse struct {
	Token string `json:"token"`
}

// JWTService represents the JWT service contract
type JWTService interface {
	GenerateToken(userID int64, username string, role Role) (string, error)
	ValidateToken(token string) (*JWTClaims, error)
}

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
}
