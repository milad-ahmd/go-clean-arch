package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"go.uber.org/zap"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	userUseCase domain.UserUseCase
	logger      logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(r *mux.Router, userUseCase domain.UserUseCase, logger logger.Logger) {
	handler := &AuthHandler{
		userUseCase: userUseCase,
		logger:      logger,
	}

	r.HandleFunc("/auth/login", handler.Login).Methods("POST")
	r.HandleFunc("/auth/register", handler.Register).Methods("POST")
	r.HandleFunc("/auth/me", handler.Me).Methods("GET")
}

// Login handles user login
// @Summary Login user
// @Description Login user and get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Login Request"
// @Success 200 {object} domain.TokenResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginReq domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		h.logger.Error("Failed to decode login request", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate login request
	if loginReq.Email == "" || loginReq.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Login user
	token, err := h.userUseCase.Login(r.Context(), loginReq.Email, loginReq.Password)
	if err != nil {
		h.logger.Error("Failed to login user", zap.String("email", loginReq.Email), zap.Error(err))
		if err == domain.ErrUnauthorized {
			respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to login")
		return
	}

	respondWithJSON(w, http.StatusOK, domain.TokenResponse{Token: token})
}

// Register handles user registration
// @Summary Register user
// @Description Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Register Request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var registerReq domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&registerReq); err != nil {
		h.logger.Error("Failed to decode register request", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate register request
	if registerReq.Username == "" || registerReq.Email == "" || registerReq.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	// Create user
	user := &domain.User{
		Username: registerReq.Username,
		Email:    registerReq.Email,
		Password: registerReq.Password,
		Role:     domain.RoleUser, // Default role
	}

	if err := h.userUseCase.Register(r.Context(), user); err != nil {
		h.logger.Error("Failed to register user", zap.String("email", registerReq.Email), zap.Error(err))
		if _, ok := err.(*domain.ConflictError); ok {
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"id":       user.ID,
		"message":  "User registered successfully",
		"username": user.Username,
		"email":    user.Email,
	})
}

// Me handles getting the current user
// @Summary Get current user
// @Description Get the current user from JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.User
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Authorization header is required")
		return
	}

	// Check if the header has the Bearer prefix
	if !strings.HasPrefix(authHeader, "Bearer ") {
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
		return
	}

	// Extract the token
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate token and get user
	user, err := h.userUseCase.ValidateToken(r.Context(), token)
	if err != nil {
		h.logger.Error("Failed to validate token", zap.Error(err))
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}
