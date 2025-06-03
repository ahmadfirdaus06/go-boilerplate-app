package utils

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/ahmadfirdaus06/go-boilerplate-app/app/externals"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/http/types"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/models"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/repo"
	appTypes "github.com/ahmadfirdaus06/go-boilerplate-app/app/types"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/utils"

	"github.com/gertd/go-pluralize"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func PrintRoutes(e *echo.Echo) {
	routes := e.Routes()

	// Filter out "echo_route_not_found"
	filtered := make([]*echo.Route, 0, len(routes))
	for _, route := range routes {
		if route.Method != "echo_route_not_found" {
			filtered = append(filtered, route)
		}
	}

	// Sort by route.Path
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Path < filtered[j].Path
	})

	// Print table with Method and Path only
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "METHOD\tPATH")

	for _, route := range filtered {
		fmt.Fprintf(w, "%s\t%s\n", route.Method, route.Path)
	}

	w.Flush()
}

// Generate CRUD resource routes, must pass type T, usually the model stuct of the intended data
func GenerateResourceRoutes[T any](resourceName string, config types.GenerateResourceRoutesConfig) {
	pluralize := pluralize.NewClient()
	resourceNameSingular := pluralize.Singular(resourceName)

	mongoExt, mongoExtError := externals.GetExternal[*externals.MongoDBExternal](config.Externals)

	if mongoExtError != nil {
		log.Fatalf("%v", mongoExtError)
		return
	}

	repo := &repo.UserRepo[T]{
		BaseRepo: &repo.BaseRepo[T]{
			DB: appTypes.AppDB{
				MongoDB: mongoExt.DB,
			},
			Collection: resourceName,
			UpdatedAt:  true,
			CreatedAt:  true,
		},
	}

	if config.Create.Enabled || config.GetAll.Enabled {
		routesWithoutId := config.Router.Group(fmt.Sprintf("/%s", resourceName))

		if config.Create.Enabled {
			if config.Create.Override != nil {
				routesWithoutId.POST("", func(c echo.Context) error {
					handler := config.Create.Override

					for _, middleware := range config.GetAll.Middlewares {
						handler = middleware(handler)
					}

					return handler(c)
				})
			} else {
				routesWithoutId.POST("", func(c echo.Context) error {
					handler := func(c echo.Context) error {
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

						if createdErr != nil {
							return echo.NewHTTPError(500, createdErr)
						}

						var output any

						if config.Create.OutputSchema != nil {
							output = &config.Create.OutputSchema
						}

						if err := utils.BindData(created, &output); err != nil {
							return err
						}

						return c.JSON(201, echo.Map{"data": output})
					}
					if len(config.Create.Middlewares) > 0 {
						for _, middleware := range config.Create.Middlewares {
							handler = middleware(handler)
						}
					}

					return handler(c)
				})
			}
		}

		if config.GetAll.Enabled {
			if config.GetAll.Override != nil {
				routesWithoutId.GET("", func(c echo.Context) error {
					handler := config.GetAll.Override

					for _, middleware := range config.GetAll.Middlewares {
						handler = middleware(handler)
					}

					return handler(c)
				})
			} else {
				routesWithoutId.GET("", func(c echo.Context) error {
					handler := func(c echo.Context) error {
						filters, sorts := ParseQueryParams(c.QueryParams())
						pageString := c.QueryParam("page")
						perPageString := c.QueryParam("per_page")

						var (
							page    = 1
							perPage = 10
						)

						if pageString != "" {
							if parsedPage, err := strconv.Atoi(pageString); err == nil {
								page = parsedPage
							} else {
								return echo.NewHTTPError(400, fmt.Sprintf("Invalid pagination param: %s", "page"))
							}
						}

						if perPageString != "" {
							if parsedPerPage, err := strconv.Atoi(perPageString); err == nil {
								perPage = parsedPerPage
							} else {
								return echo.NewHTTPError(400, fmt.Sprintf("Invalid pagination params: %s", "per_page"))
							}
						}

						all, getAllErr := repo.GetAll(true, &appTypes.GetAllFiltersAndSorts{QueryParamsFilters: filters, QueryParamsSortFields: sorts}, &appTypes.PaginationParams{
							Page:    page,
							PerPage: perPage,
						})

						if getAllErr != nil {
							return echo.NewHTTPError(500, getAllErr)
						}

						var outputResults struct {
							Records    []any `json:"records"`
							Page       int   `json:"page"`
							PerPage    int   `json:"per_page"`
							Total      int   `json:"total"`
							TotalPages int   `json:"total_pages"`
						}

						if err := utils.BindData(all, &outputResults); err != nil {
							return err
						}

						if config.GetAll.OutputSchema != nil {
							var records []any
							for _, record := range all.Records {
								output := &config.GetAll.OutputSchema
								if err := utils.BindData(record, output); err != nil {
									return err
								}

								records = append(records, output)
							}

							outputResults.Records = records
						}

						return c.JSON(200, echo.Map{"data": outputResults})
					}

					for _, middleware := range config.GetAll.Middlewares {
						handler = middleware(handler)
					}

					return handler(c)

				})
			}

		}

	}

	if config.GetById.Enabled || config.UpdateById.Enabled || config.DeleteById.Enabled {
		routesWithId := config.Router.Group(fmt.Sprintf("/%s/:%s", resourceName, resourceNameSingular))

		routesWithId.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				objectId, parseObjIdErr := bson.ObjectIDFromHex(c.Param(resourceNameSingular))

				if parseObjIdErr != nil {
					return echo.NewHTTPError(400, fmt.Sprintf("Invalid resource identifier: %s", c.Param(resourceNameSingular)))
				}

				resource, getByIdErr := repo.GetByID(objectId)

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
				routesWithId.GET("", func(c echo.Context) error {
					handler := config.GetById.Override

					for _, middleware := range config.GetAll.Middlewares {
						handler = middleware(handler)
					}

					return handler(c)
				})
			} else {
				routesWithId.GET("", func(c echo.Context) error {
					handler := func(c echo.Context) error {
						all, getAllErr := repo.GetByID(c.Param(resourceNameSingular))

						if getAllErr != nil {
							return echo.NewHTTPError(500, getAllErr)
						}

						return c.JSON(200, echo.Map{"data": all})
					}

					for _, middleware := range config.GetAll.Middlewares {
						handler = middleware(handler)
					}

					return handler(c)
				})

			}

		}

		if config.UpdateById.Enabled {
			if config.UpdateById.Override != nil {
				routesWithId.PUT("", func(c echo.Context) error {
					handler := config.UpdateById.Override

					for _, middleware := range config.GetAll.Middlewares {
						handler = middleware(handler)
					}

					return handler(c)
				})
			} else {
				routesWithId.PUT("", func(c echo.Context) error {
					handler := func(c echo.Context) error {
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

						return c.JSON(200, echo.Map{"data": updated})
					}

					for _, middleware := range config.GetAll.Middlewares {
						handler = middleware(handler)
					}

					return handler(c)
				})
			}
		}

		if config.DeleteById.Enabled {
			if config.DeleteById.Override != nil {
				routesWithId.DELETE("", func(c echo.Context) error {
					handler := config.DeleteById.Override

					for _, middleware := range config.GetAll.Middlewares {
						handler = middleware(handler)
					}

					return handler(c)
				})
			} else {
				routesWithId.DELETE("", func(c echo.Context) error {
					handler := func(c echo.Context) error {
						_, deleteByIdErr := repo.DeleteByID(c.Param(resourceNameSingular))

						if deleteByIdErr != nil {
							return echo.NewHTTPError(500, deleteByIdErr)
						}

						return c.JSON(204, echo.Map{})
					}

					for _, middleware := range config.GetAll.Middlewares {
						handler = middleware(handler)
					}

					return handler(c)
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

func GetAuthUser(c echo.Context) *models.User {
	return c.Get("auth").(*models.User)
}

func ParseQueryParams(query url.Values) ([]appTypes.QueryParamsFilter, []appTypes.QueryParamsSortField) {
	var filters []appTypes.QueryParamsFilter
	var sorts []appTypes.QueryParamsSortField

	// Parse filters
	for key, values := range query {
		if !strings.HasPrefix(key, "filter.") {
			continue
		}

		fieldParts := strings.Split(key[len("filter."):], ".")
		field := fieldParts[0]
		operator := appTypes.OpEq

		if len(fieldParts) > 1 {
			operator = appTypes.QueryParamsFilterOp(fieldParts[1])
		}

		for _, val := range values {
			filters = append(filters, appTypes.QueryParamsFilter{
				Field:    field,
				Operator: operator,
				Value:    val,
			})
		}
	}

	// Parse sort
	if sortQuery := query.Get("sort"); sortQuery != "" {
		for _, s := range strings.Split(sortQuery, ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			desc := false
			if strings.HasPrefix(s, "-") {
				desc = true
				s = s[1:]
			}
			sorts = append(sorts, appTypes.QueryParamsSortField{
				Field:      s,
				Descending: desc,
			})
		}
	}

	return filters, sorts
}
