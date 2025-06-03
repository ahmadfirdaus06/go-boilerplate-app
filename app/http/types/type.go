package types

import (
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/externals"

	"github.com/labstack/echo/v4"
)

type ControllerConfig struct {
	Override     echo.HandlerFunc      // Override predefined controller for CRUD resources, able to custom to your own logic, default is nil
	Enabled      bool                  // Toggle endpoint availability, must explicitly set to true, default is false
	InputSchema  any                   // Validation schema towards request body, refer github.com/go-playground/validator/v10, not applied by default
	OutputSchema any                   // Act as output filter to hide/show certain fields from controller to response body, not applied by default
	Middlewares  []echo.MiddlewareFunc // Applied multiple middleware(s) to current route, works with overriden controller too, default is empty/not applied
}

type GenerateResourceRoutesConfig struct {
	Router     *echo.Group                // Base echo.Group or any extended one
	GetAll     ControllerConfig           // Get all resource route e.g: GET /resources
	Create     ControllerConfig           // Create a single resource route e.g: POST /resources
	GetById    ControllerConfig           // Get single resource by id route e.g: GET /resources/:resourceId
	UpdateById ControllerConfig           // Update single resource properties by id route e.g: PUT /resources/:resourceId
	DeleteById ControllerConfig           // Delete single resource by id route e.g: DELETE /resources/:resourceId
	Externals  *externals.AllAppExternals // All app external must be pass here as dependency injection
}
