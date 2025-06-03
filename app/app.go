package app

import (
	"fmt"
	"reflect"

	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/externals"
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/http/middlewares"
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/http/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type HttpAppConfig struct {
	Externals        *externals.AllAppExternals                    // Registered app dependencies
	AppPort          string                                        // App port number config in string, default is port 1234
	APIBasePrefixUrl string                                        // Custom base api prefix, default /api
	Routes           func(*echo.Group, *externals.AllAppExternals) // Collection of echo.Echo routes
}

// Initialize http web server using echo.Echo
func InitHttpApp(config *HttpAppConfig) *echo.Echo {
	e := echo.New()

	e.HideBanner = true

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method} ${status} ${uri}\n",
	}))

	e.Use(middleware.Recover())

	validatorInstance := validator.New()

	// Make validator using json embedded struct as default input label
	validatorInstance.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tag := fld.Tag.Get("json")
		if tag == "-" {
			return ""
		}
		return tag
	})

	// Using default input validator  refer github.com/go-playground/validator/v10
	e.Validator = &middlewares.CustomValidator{Validator: validatorInstance}

	// Update enhanced error handler
	e.HTTPErrorHandler = middlewares.CustomHTTPErrorHandler

	// WIP enable websocket
	// e.GET("/ws", websocket.HandleWebSocket)
	// e.GET("/ws/:namespace", websocket.HandleWebSocket)

	var router *echo.Group

	// Configure app routes base prefix
	if config.APIBasePrefixUrl != "" {
		router = e.Group(config.APIBasePrefixUrl)
	} else {
		router = e.Group("/api")
	}

	// Registing app routes
	config.Routes(router, config.Externals)

	// Printing routes
	utils.PrintRoutes(e)

	if config.AppPort != "" {
		e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", config.AppPort)))
	} else {
		e.Logger.Fatal(e.Start(":1234"))
	}

	return e
}
