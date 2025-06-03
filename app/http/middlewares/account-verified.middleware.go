package middlewares

import (
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/http/utils"

	"github.com/labstack/echo/v4"
)

func AccountVerified() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := utils.GetAuthUser(c)

			if user.EmailVerifiedAt == nil {
				return echo.NewHTTPError(403, "Please verify your account.")
			}

			return next(c)
		}
	}
}
