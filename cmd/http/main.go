package main

import (
	"fmt"
	"go-boilerplate-backend/internal/externals"
	"go-boilerplate-backend/internal/middlewares"
	"go-boilerplate-backend/internal/routes"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.HideBanner = true

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method} ${status} ${uri}\n",
	}))

	e.Use(middleware.Recover())

	validatorInstance := validator.New()

	validatorInstance.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tag := fld.Tag.Get("json")
		if tag == "-" {
			return ""
		}
		return tag
	})

	e.Validator = &middlewares.CustomValidator{Validator: validatorInstance}

	e.HTTPErrorHandler = middlewares.CustomHTTPErrorHandler

	externals, externalsErr := externals.RegisterExternals()

	if externalsErr != nil {
		e.Logger.Fatal("Failed to register externals:", externalsErr)
	}

	routes.InitRoutes(e, externals)

	// Print all routes
	for _, route := range e.Routes() {
		if route.Method != "echo_route_not_found" {
			fmt.Printf("%s %s\n", route.Method, route.Path)
		}
	}

	e.Logger.Fatal(e.Start(":1234"))
}
