package examples

import (
	"go-boilerplate-backend/internal/externals"
	"go-boilerplate-backend/internal/utils"

	"github.com/labstack/echo/v4"
)

func InitNoteRoutes(router *echo.Group, externals *externals.AllAppExternals) {
	utils.GenerateResourceRoutes("notes", utils.GenerateResourceRoutesConfig{
		Router:    router,
		Externals: externals,
		GetAll: utils.ControllerConfig{
			Enabled: true,
		},
		GetById: utils.ControllerConfig{
			Enabled: true,
		},
		Create: utils.ControllerConfig{
			Enabled: true,
		},
		UpdateById: utils.ControllerConfig{
			Enabled: true,
		},
		DeleteById: utils.ControllerConfig{
			Enabled: true,
		},
	})
}
