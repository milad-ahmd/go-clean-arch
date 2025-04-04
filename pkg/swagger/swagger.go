package swagger

import (
	"net/http"

	"github.com/gorilla/mux"
)

// @title Go Clean Architecture API
// @version 1.0
// @description This is a sample server for a Go Clean Architecture API.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// SetupSwagger sets up Swagger documentation
func SetupSwagger(r *mux.Router) {
	// Serve the Swagger UI and spec
	fs := http.FileServer(http.Dir("./docs/swagger"))
	r.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", fs))
	r.Handle("/swagger.yaml", fs)
}
