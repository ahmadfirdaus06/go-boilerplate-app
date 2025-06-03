package middlewares

import (
	"net/http"

	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/externals"
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/models"
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/repo"
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/services"
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/types"
	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/utils"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
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

func Auth(appExternals *externals.AllAppExternals) echo.MiddlewareFunc {
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
				ID    bson.ObjectID `json:"_id" bson:"_id"`
				Email string        `json:"email"`
			}

			if bindErr := utils.BindData(claims["user"], &user); bindErr != nil {
				c.Error(echo.NewHTTPError(400, "Invalid token contents."))
				resetTokenCookie(c)
				return
			}

			mongoExt, mongoExtErr := externals.GetExternal[*externals.MongoDBExternal](appExternals)

			if mongoExtErr != nil {
				c.Error(mongoExtErr)
				return
			}

			userDetails, getUserErr := repo.NewUserRepo[models.User](types.AppDB{MongoDB: mongoExt.DB}, "users").GetByID(user.ID)

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
