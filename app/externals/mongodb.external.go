package externals

import (
	"context"

	"github.com/ahmadfirdaus06/go-boilerplate-app/app/utils"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDBExternal struct {
	DB *mongo.Database
}

func NewMongoDBExternal() *MongoDBExternal {
	return &MongoDBExternal{}
}

func (me *MongoDBExternal) Connect() (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(utils.GetAppConfig("MONGODB_URI")))

	if err != nil {
		return nil, err
	}

	me.DB = client.Database(utils.GetAppConfig("MONGODB_DATABASE"))

	return client, err
}

func (me *MongoDBExternal) ConnectRaw() error {
	_, err := me.Connect()

	return err
}

func (me *MongoDBExternal) Healthcheck() error {
	return me.DB.Client().Ping(context.Background(), nil)
}

func (me *MongoDBExternal) SuccessMessage() string {
	return "MongoDB connected."
}

var _ BaseExternal = (*MongoDBExternal)(nil)
var _ External[*mongo.Client] = (*MongoDBExternal)(nil)
