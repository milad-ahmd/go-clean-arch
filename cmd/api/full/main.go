package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/milad-ahmd/go-clean-arch/internal/delivery/http"
	"github.com/milad-ahmd/go-clean-arch/internal/repository/postgres"
	"github.com/milad-ahmd/go-clean-arch/internal/usecase"
	"github.com/milad-ahmd/go-clean-arch/pkg/auth"
	"github.com/milad-ahmd/go-clean-arch/pkg/config"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"github.com/milad-ahmd/go-clean-arch/pkg/swagger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	log := logger.NewLogger(cfg.Logger.Level)
	log.Info("Starting application")

	// Connect to database
	db, err := postgres.NewPostgresConnection(cfg, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Create tables
	if err := postgres.InitTables(db, log); err != nil {
		log.Fatal("Failed to create tables", zap.Error(err))
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db, log)
	categoryRepo := postgres.NewCategoryRepository(db, log)
	productRepo := postgres.NewProductRepository(db, log)
	orderRepo := postgres.NewOrderRepository(db, log)

	// Initialize services
	jwtService := auth.NewJWTService(cfg.Auth.JWTSecret, log)

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(userRepo, jwtService, log)
	categoryUseCase := usecase.NewCategoryUseCase(categoryRepo, log)
	productUseCase := usecase.NewProductUseCase(productRepo, categoryRepo, log)
	orderUseCase := usecase.NewOrderUseCase(orderRepo, productRepo, userRepo, log)

	// Initialize HTTP server
	server := http.NewServer(cfg, log)
	server.SetupMiddleware()

	// Register HTTP handlers
	http.NewUserHandler(server.Router(), userUseCase, log)
	http.NewAuthHandler(server.Router(), userUseCase, log)
	http.NewCategoryHandler(server.Router(), categoryUseCase, log)
	http.NewProductHandler(server.Router(), productUseCase, log)
	http.NewOrderHandler(server.Router(), orderUseCase, log)

	// Setup Swagger
	swagger.SetupSwagger(server.Router())

	// Start server in a goroutine
	go func() {
		log.Info("Server is running on port " + cfg.Server.Port)
		if err := server.Start(); err != nil {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down application")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Application stopped")
}
