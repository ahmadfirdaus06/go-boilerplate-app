package utils

import (
	"encoding/json"
	"fmt"
	"go-boilerplate-backend/internal/externals"
	"go-boilerplate-backend/internal/models"
	"go-boilerplate-backend/internal/repo"
	"log"
	"net/http"
	"unicode"

	"github.com/gertd/go-pluralize"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type ControllerConfig struct {
	Override     echo.HandlerFunc
	Enabled      bool
	InputSchema  any
	OutputSchema any
}

type GenerateResourceRoutesConfig struct {
	Router     *echo.Group
	GetAll     ControllerConfig
	Create     ControllerConfig
	GetById    ControllerConfig
	UpdateById ControllerConfig
	DeleteById ControllerConfig
	Externals  *externals.AllAppExternals
}

func GenerateResourceRoutes(resourceName string, config GenerateResourceRoutesConfig) {
	pluralize := pluralize.NewClient()
	resourceNameSingular := pluralize.Singular(resourceName)

	repo := &repo.UserRepo[interface{}]{
		BaseRepo: &repo.BaseRepo[interface{}]{
			DB:         config.Externals.JsonDB,
			Collection: resourceName,
			UpdatedAt:  true,
			CreatedAt:  true,
		},
	}

	if config.Create.Enabled || config.GetAll.Enabled {
		routesWithoutId := config.Router.Group(fmt.Sprintf("/%s", resourceName))

		if config.Create.Enabled {

			if config.Create.Override != nil {
				routesWithoutId.POST("", config.Create.Override)
			} else {
				routesWithoutId.POST("", func(c echo.Context) error {

					var inputs any

					if config.Create.InputSchema != nil {
						inputs = config.Create.InputSchema
						if err := c.Bind(&inputs); err != nil {
							return echo.NewHTTPError(http.StatusBadRequest, err.Error())
						}
						if err := c.Validate(inputs); err != nil {
							return err
						}
					} else {
						if err := c.Bind(&inputs); err != nil {
							return echo.NewHTTPError(http.StatusBadRequest, err.Error())
						}
					}

					created, createdErr := repo.Create(inputs)

					if config.Create.OutputSchema != nil {
						bytes, err := json.Marshal(created)
						if err != nil {
							log.Fatal("marshal failed:", err)
						}

						created = &config.Create.OutputSchema

						if err := json.Unmarshal(bytes, &created); err != nil {
							log.Fatal("unmarshal failed:", err)
						}
					}

					if createdErr != nil {
						return echo.NewHTTPError(500, createdErr)
					}

					return c.JSON(201, echo.Map{"data": created})
				})
			}
		}

		if config.GetAll.Enabled {
			if config.GetAll.Override != nil {
				routesWithoutId.GET("", config.GetAll.Override)
			} else {
				routesWithoutId.GET("", func(c echo.Context) error {
					all, getAllErr := repo.GetAll(nil)

					if getAllErr != nil {
						return echo.NewHTTPError(500, getAllErr)
					}

					return c.JSON(200, echo.Map{"data": all})
				})
			}

		}

	}

	if config.GetById.Enabled || config.UpdateById.Enabled || config.DeleteById.Enabled {
		routesWithId := config.Router.Group(fmt.Sprintf("/%s/:%s", resourceName, resourceNameSingular))

		routesWithId.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				resource, getByIdErr := repo.GetByID(c.Param(resourceNameSingular))

				if getByIdErr != nil {
					return echo.NewHTTPError(500, getByIdErr)
				}

				if resource == nil {
					return echo.NewHTTPError(404)
				}

				return next(c)
			}
		})

		if config.GetById.Enabled {
			if config.GetById.Override != nil {
				routesWithId.GET("", config.GetById.Override)
			} else {
				routesWithId.GET("", func(c echo.Context) error {
					all, getAllErr := repo.GetByID(c.Param(resourceNameSingular))

					if getAllErr != nil {
						return echo.NewHTTPError(500, getAllErr)
					}

					return c.JSON(200, echo.Map{"data": all})
				})
			}

		}

		if config.UpdateById.Enabled {
			if config.UpdateById.Override != nil {
				routesWithId.PUT("", config.UpdateById.Override)
			} else {
				routesWithId.PUT("", func(c echo.Context) error {

					var inputs any

					if config.UpdateById.InputSchema != nil {
						inputs = config.UpdateById.InputSchema
						if err := c.Bind(inputs); err != nil {
							return echo.NewHTTPError(http.StatusBadRequest, err.Error())
						}
						if err := c.Validate(inputs); err != nil {
							return err
						}
					} else {
						if err := c.Bind(&inputs); err != nil {
							return echo.NewHTTPError(http.StatusBadRequest, err.Error())
						}
					}

					updated, updatedErr := repo.UpdateByID(c.Param(resourceNameSingular), inputs)

					if updatedErr != nil {
						return echo.NewHTTPError(500, updatedErr)
					}

					return c.JSON(201, echo.Map{"data": updated})
				})
			}
		}

		if config.DeleteById.Enabled {
			if config.DeleteById.Override != nil {
				routesWithId.DELETE("", config.DeleteById.Override)
			} else {
				routesWithId.DELETE("", func(c echo.Context) error {

					_, deleteByIdErr := repo.DeleteByID(c.Param(resourceNameSingular))

					if deleteByIdErr != nil {
						return echo.NewHTTPError(500, deleteByIdErr)
					}

					return c.JSON(204, echo.Map{})
				})
			}
		}
	}
}

func ValidateInput(c echo.Context, schemaData interface{}) error {
	if err := c.Bind(&schemaData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(schemaData); err != nil {
		return err
	}

	return nil
}

func NormalizeFieldName(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func BindData(input any, outputSchema any) error {
	bytes, err := json.Marshal(input)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &outputSchema); err != nil {
		return err
	}

	return nil
}

func GenerateJWTtToken(claims *jwt.MapClaims, JWTSecret string) (string, error) {
	if JWTSecret == "" {
		return "", fmt.Errorf("NO SECRET PROVIDED")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(JWTSecret))
}

func VerifyJWTToken(tokenString string, JWTSecret string) (any, error) {
	if JWTSecret == "" {
		return "", fmt.Errorf("NO SECRET PROVIDED")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("CLAIMS DOES NOT EXIST")
}

func GetAuthUser(c echo.Context) *models.User {
	return c.Get("auth").(*models.User)
}
