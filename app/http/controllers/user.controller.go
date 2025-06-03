package controllers

import (
	"time"

	"github.com/ahmadfirdaus06/go-boilerplate-app/app/externals"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/http/utils"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/services"
	appUtils "github.com/ahmadfirdaus06/go-boilerplate-app/app/utils"

	"github.com/labstack/echo/v4"
)

func RegisterUserController(externals *externals.AllAppExternals) func(c echo.Context) error {
	return func(c echo.Context) error {
		var createUserInputs = new(struct {
			Email           string `json:"email" validate:"email,required"`
			Username        string `json:"username" validate:"required"`
			FirstName       string `json:"firstName" validate:"required"`
			LastName        string `json:"lastName" validate:"required"`
			Password        string `json:"password" validate:"required,min=8,eqfield=ConfirmPassword"`
			ConfirmPassword string `json:"confirmPassword" validate:"required,min=8,eqfield=Password"`
		})

		if err := utils.ValidateInput(c, createUserInputs); err != nil {
			return err
		}

		result, serviceErr := services.NewUserService(externals).RegisterUser(createUserInputs)

		if serviceErr != nil {
			return serviceErr
		}

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

		if bindErr := appUtils.BindData(result, &outputData); bindErr != nil {
			return bindErr
		}

		return c.JSON(201, echo.Map{"data": outputData})
	}
}
