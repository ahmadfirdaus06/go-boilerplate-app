package middlewares

import (
	"fmt"
	"net/http"

	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/http/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code

		if code < 500 {
			var message any
			switch m := he.Message.(type) {
			case string:
				message = m
			case error:
				message = m.Error()
			default:
				message = fmt.Sprintf("%v", m)
			}
			c.JSON(code, echo.Map{"message": message})
			return
		} else {
			c.Logger().Error(err)
			c.JSON(code, echo.Map{"message": http.StatusText(code)})
			return
		}
	}

	if err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			code = http.StatusUnprocessableEntity
			formatted := formatErrorsWithJSONTags(errs)
			c.JSON(code, echo.Map{"message": http.StatusText(code), "errors": formatted})
			return
		}
	}

	c.Logger().Error(err)
	c.JSON(code, echo.Map{"message": http.StatusText(code)})
}

func formatErrorsWithJSONTags(errs validator.ValidationErrors) map[string]string {
	errors := make(map[string]string)
	for _, err := range errs {
		field := err.Field() // this will now return the `json` tag value
		errors[field] = fmt.Sprintf("%s %s", field, validationMessage(err))
	}
	return errors
}

func validationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "is a required field."
	case "email":
		return "is not a valid email."
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s.", err.Param())
	case "eqfield":
		return fmt.Sprintf("does not match with %s.", utils.NormalizeFieldName(err.Param()))
	case "min":
		return fmt.Sprintf("must be a minimum of %s character(s).", err.Param())
	case "len":
		return fmt.Sprintf("must be of %s character(s).", err.Param())
	// Add more cases as needed
	default:
		return fmt.Sprintf("not valid (%s)", err.Tag())
	}
}
