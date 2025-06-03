package routes

import (
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/externals"
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/http/controllers"
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/http/middlewares"

	"github.com/labstack/echo/v4"
)

func InitAuthRoute(router *echo.Group, externals *externals.AllAppExternals) {
	authRoute := router.Group("/auth")

	authRoute.POST("/login", controllers.Login(externals))
	authRoute.Use(middlewares.Auth(externals))
	authRoute.GET("", controllers.GetAuthUser(externals))
	authRoute.POST("/verification/code/send", controllers.SendVerificationCode(externals))
	authRoute.POST("/verification/code/verify", controllers.VerifyAuthCode(externals))
}
