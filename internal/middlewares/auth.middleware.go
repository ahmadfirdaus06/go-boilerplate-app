package middlewares

import (
	"go-boilerplate-backend/internal/externals"
	"go-boilerplate-backend/internal/models"
	"go-boilerplate-backend/internal/repo"
	"go-boilerplate-backend/internal/services"
	"go-boilerplate-backend/internal/utils"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func resetTokenCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/", // should match the original cookie path
		MaxAge:   -1,  // tell the browser to delete it
		HttpOnly: true,
	})
}

func Auth(externals *externals.AllAppExternals) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:  []byte(services.JwtSecret),
		TokenLookup: "header:Authorization:Bearer,cookie:token",
		SuccessHandler: func(c echo.Context) {
			token, ok := c.Get("user").(*jwt.Token)

			if !ok {
				c.Error(echo.NewHTTPError(400, "Invalid token."))
				resetTokenCookie(c)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				c.Error(echo.NewHTTPError(400, "Invalid token contents."))
				resetTokenCookie(c)
				return
			}

			var user struct {
				ID    string `json:"id"`
				Email string `json:"email"`
			}

			if bindErr := utils.BindData(claims["user"], &user); bindErr != nil {
				c.Error(echo.NewHTTPError(400, "Invalid token contents."))
				resetTokenCookie(c)
				return
			}

			userDetails, getUserErr := repo.NewUserRepo[models.User](externals.JsonDB, "users").GetByID(user.ID)

			if getUserErr != nil {
				c.Error(getUserErr)
				resetTokenCookie(c)
				return
			}

			if userDetails == nil {
				c.Error(echo.NewHTTPError(401, "Account does not exist."))
				resetTokenCookie(c)
				return
			}

			c.Set("auth", userDetails)
		},
	})
}
