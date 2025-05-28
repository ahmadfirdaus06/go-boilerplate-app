package routes

import (
	"go-boilerplate-backend/internal/controllers"
	"go-boilerplate-backend/internal/externals"
	"go-boilerplate-backend/internal/middlewares"

	"github.com/labstack/echo/v4"
)

func InitAuthRoute(router *echo.Group, externals *externals.AllAppExternals) {
	authRoute := router.Group("/auth")

	authRoute.POST("/login", controllers.Login(externals))
	authRoute.Use(middlewares.Auth(externals))
	authRoute.GET("/", controllers.GetAuthUser(externals))
	authRoute.POST("/verification/code/send", controllers.SendVerificationCode(externals))
}
