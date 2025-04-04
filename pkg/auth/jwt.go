package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"go.uber.org/zap"
)

// JWTService implements domain.JWTService
type jwtService struct {
	secretKey string
	logger    logger.Logger
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, logger logger.Logger) domain.JWTService {
	return &jwtService{
		secretKey: secretKey,
		logger:    logger,
	}
}

// GenerateToken generates a new JWT token
func (s *jwtService) GenerateToken(userID int64, username string, role domain.Role) (string, error) {
	// Create the claims
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		s.logger.Error("Failed to sign token", zap.Error(err))
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token
func (s *jwtService) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		s.logger.Error("Failed to parse token", zap.Error(err))
		return nil, err
	}

	// Validate the token
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Get the claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Extract the claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid user_id claim")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid username claim")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid role claim")
	}

	// Create the JWT claims
	jwtClaims := &domain.JWTClaims{
		UserID:   int64(userID),
		Username: username,
		Role:     domain.Role(role),
	}

	return jwtClaims, nil
}
