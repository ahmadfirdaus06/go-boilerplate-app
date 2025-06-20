package main

import (
	"log"
	"time"

	"github.com/ahmadfirdaus06/go-boilerplate-app/app"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/externals"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/http/types"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/http/utils"
	appUtils "github.com/ahmadfirdaus06/go-boilerplate-app/app/utils"

	"github.com/labstack/echo/v4"
)

func InitNoteRoutes(e *echo.Group, externals *externals.AllAppExternals) {
	type Note struct {
		Title       string     `bson:"title" json:"title"`
		Description string     `bson:"description" json:"description"`
		CreatedAt   *time.Time `bson:"createdAt" json:"createdAt"`
		UpdatedAt   *time.Time `bson:"updatedAt" json:"updatedAt"`
	}

	utils.GenerateResourceRoutes[Note]("notes", types.GenerateResourceRoutesConfig{
		Router:    e,
		Externals: externals,
		GetAll: types.ControllerConfig{
			Enabled: true,
		},
		GetById: types.ControllerConfig{
			Enabled: true,
		},
		Create: types.ControllerConfig{
			Enabled: true,
		},
		UpdateById: types.ControllerConfig{
			Enabled: true,
		},
		DeleteById: types.ControllerConfig{
			Enabled: true,
		},
	})
}

func main() {
	// required envs
	requiredEnvs := []string{"MONGODB_URI", "MONGODB_DATABASE"}

	// load and check missing envs
	appUtils.LoadAppEnv(requiredEnvs)

	var all []externals.BaseExternal
	mongoDbExternal := externals.MongoDBExternal{}
	// can declare more externals here, refer externals.MongoDBExternal{} implementation

	all = append(all, &mongoDbExternal)

	// register all external dependencies
	externals, externalsErr := externals.RegisterExternals(all)

	if externalsErr != nil {
		log.Fatalf("%v", externalsErr)
	}

	// basic app config including note crud routing and registered externals
	config := &app.HttpAppConfig{
		Routes:    InitNoteRoutes,
		Externals: externals,
	}

	app.InitHttpApp(config)
}
