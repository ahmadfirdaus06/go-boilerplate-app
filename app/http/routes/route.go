package routes

import (
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/externals"
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/http/controllers"
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/http/types"
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/http/utils"
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/models"

	"github.com/labstack/echo/v4"
)

type RouteInit func(*echo.Echo, *externals.AllAppExternals)

func InitRoutes(e *echo.Echo, externals *externals.AllAppExternals) {
	apiRoutes := e.Group("/api/v1")

	utils.GenerateResourceRoutes[models.User]("users", types.GenerateResourceRoutesConfig{
		Router: apiRoutes,
		Create: types.ControllerConfig{
			Enabled:  true,
			Override: controllers.RegisterUserController(externals),
		},
		GetAll: types.ControllerConfig{
			Enabled: true,
		},
		Externals: externals,
	})

	InitAuthRoute(apiRoutes, externals)
}
