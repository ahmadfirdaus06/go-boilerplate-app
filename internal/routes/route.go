package routes

import (
	"go-boilerplate-backend/internal/controllers"
	"go-boilerplate-backend/internal/externals"
	"go-boilerplate-backend/internal/utils"

	"github.com/labstack/echo/v4"
)

func InitRoutes(e *echo.Echo, externals *externals.AllAppExternals) {
	apiRoutes := e.Group("/api/v1")

	utils.GenerateResourceRoutes("users", utils.GenerateResourceRoutesConfig{
		Router: apiRoutes,
		Create: utils.ControllerConfig{
			Enabled:  true,
			Override: controllers.RegisterUserController(externals),
		},
		Externals: externals,
	})

	InitAuthRoute(apiRoutes, externals)
}
