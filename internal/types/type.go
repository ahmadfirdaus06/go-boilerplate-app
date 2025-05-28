package types

import (
	"github.com/labstack/echo/v4"
)

type CrudController struct {
	GetAll  echo.HandlerFunc
	GetByID echo.HandlerFunc
	Create  echo.HandlerFunc
	Update  echo.HandlerFunc
	Delete  echo.HandlerFunc
}

type CrudRouteOptions struct {
	Group        *echo.Group
	Controller   CrudController
	IDParamName  string
	UpdateMethod string
	Middleware   []echo.MiddlewareFunc
}

type AppExternals interface {
	HealthCheck() error
	Connect() (any, error)
}

type RepoTimestampConfig struct {
	CreatedAt bool
	UpdatedAt bool
}
