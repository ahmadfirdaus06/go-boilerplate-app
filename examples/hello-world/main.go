package main

import (
	"github.com/ahmadfirdaus06/go-boilerplate-app/app"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/externals"

	"github.com/labstack/echo/v4"
)

func main() {
	// simple hello world app
	app.InitHttpApp(&app.HttpAppConfig{
		Routes: func(g *echo.Group, aae *externals.AllAppExternals) {
			// Test this endpoint using Postman, curl
			g.GET("/hello-world", func(c echo.Context) error {
				return c.String(200, "Hello World!")
			})
		}})
}
