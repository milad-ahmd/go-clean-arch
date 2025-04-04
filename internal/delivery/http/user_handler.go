package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"go.uber.org/zap"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	userUseCase domain.UserUseCase
	logger      logger.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(r *mux.Router, userUseCase domain.UserUseCase, logger logger.Logger) {
	handler := &UserHandler{
		userUseCase: userUseCase,
		logger:      logger,
	}

	r.HandleFunc("/users", handler.Create).Methods("POST")
	r.HandleFunc("/users", handler.List).Methods("GET")
	r.HandleFunc("/users/{id:[0-9]+}", handler.GetByID).Methods("GET")
	r.HandleFunc("/users/{id:[0-9]+}", handler.Update).Methods("PUT")
	r.HandleFunc("/users/{id:[0-9]+}", handler.Delete).Methods("DELETE")
}

// Create handles the creation of a new user
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate user input
	if user.Username == "" || user.Email == "" || user.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	if err := h.userUseCase.Create(r.Context(), &user); err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		var conflictErr *domain.ConflictError
		if errors.As(err, &conflictErr) {
			respondWithError(w, http.StatusConflict, conflictErr.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"id":       user.ID,
		"message":  "User created successfully",
		"username": user.Username,
		"email":    user.Email,
	})
}

// GetByID handles getting a user by ID
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse user ID", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userUseCase.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Int64("id", id), zap.Error(err))
		var notFoundErr *domain.NotFoundError
		if errors.As(err, &notFoundErr) {
			respondWithError(w, http.StatusNotFound, notFoundErr.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// Update handles updating a user
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse user ID for update", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.logger.Error("Failed to decode request body for update", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Set the ID from the URL
	user.ID = id

	if err := h.userUseCase.Update(r.Context(), &user); err != nil {
		h.logger.Error("Failed to update user", zap.Int64("id", id), zap.Error(err))
		var notFoundErr *domain.NotFoundError
		if errors.As(err, &notFoundErr) {
			respondWithError(w, http.StatusNotFound, notFoundErr.Error())
			return
		}
		var conflictErr *domain.ConflictError
		if errors.As(err, &conflictErr) {
			respondWithError(w, http.StatusConflict, conflictErr.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "User updated successfully",
	})
}

// Delete handles deleting a user
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse user ID for deletion", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.userUseCase.Delete(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete user", zap.Int64("id", id), zap.Error(err))
		var notFoundErr *domain.NotFoundError
		if errors.As(err, &notFoundErr) {
			respondWithError(w, http.StatusNotFound, notFoundErr.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "User deleted successfully",
	})
}

// List handles listing users with pagination
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default limit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	users, err := h.userUseCase.List(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}

// respondWithError responds with an error message
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON responds with JSON
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
