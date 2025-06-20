package controllers

import (
	"net/http"
	"time"

	"github.com/ahmadfirdaus06/go-boilerplate-app/app/externals"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/http/utils"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/services"
	appUtils "github.com/ahmadfirdaus06/go-boilerplate-app/app/utils"

	"github.com/labstack/echo/v4"
)

func SendVerificationCode(externals *externals.AllAppExternals) func(c echo.Context) error {
	return func(c echo.Context) error {
		user := utils.GetAuthUser(c)

		if user.EmailVerifiedAt != nil {
			return echo.NewHTTPError(403, "Account already verified.")
		}

		codeExpiredAt, err := services.NewAuthService(externals).SendVerificationCode(user)

		if err != nil {
			return err
		}

		return c.JSON(200, echo.Map{"message": "Code sent. Please verify your account within 2 minutes.", "data": map[string]string{
			"emailVerificationCodeExpiredAt": codeExpiredAt.String(),
		}})
	}
}

func Login(externals *externals.AllAppExternals) func(c echo.Context) error {
	return func(c echo.Context) error {
		var inputs = new(struct {
			UsernameOrEmail string `json:"usernameOrEmail" validate:"required"`
			Pasword         string `json:"password" validate:"required"`
		})

		if err := utils.ValidateInput(c, inputs); err != nil {
			return err
		}

		tokenString, err := services.NewAuthService(externals).LoginUser(inputs.UsernameOrEmail, inputs.Pasword)

		if err != nil {
			return err
		}

		c.SetCookie(&http.Cookie{Name: "token", Value: tokenString, HttpOnly: true, Path: "/"})

		return c.JSON(200, echo.Map{
			"data": map[string]string{
				"token": tokenString,
			},
		})
	}
}

func GetAuthUser(externals *externals.AllAppExternals) func(c echo.Context) error {
	return func(c echo.Context) error {
		user := utils.GetAuthUser(c)

		var outputData struct {
			ID              string     `json:"_id"`
			Username        string     `json:"username"`
			FirstName       string     `json:"firstName"`
			LastName        string     `json:"lastName"`
			Email           string     `json:"email"`
			EmailVerifiedAt *time.Time `json:"emailVerifiedAt"`
			CreatedAt       *time.Time `json:"createdAt"`
			UpdatedAt       *time.Time `json:"updatedAt"`
		}

		bindErr := appUtils.BindData(user, &outputData)

		if bindErr != nil {
			return bindErr
		}

		return c.JSON(200, echo.Map{
			"data": outputData,
		})
	}
}

func VerifyAuthCode(externals *externals.AllAppExternals) func(c echo.Context) error {
	return func(c echo.Context) error {
		var inputs = new(struct {
			VerificationCode string `json:"verificationCode" validate:"required,len=6"`
		})

		if err := utils.ValidateInput(c, inputs); err != nil {
			return err
		}

		verified, err := services.NewAuthService(externals).VerifyCode(utils.GetAuthUser(c), inputs.VerificationCode)

		if err != nil {
			return err
		}

		if !verified {
			return echo.NewHTTPError(400, "Wrong verification code.")
		}

		return GetAuthUser(externals)(c)
	}
}
