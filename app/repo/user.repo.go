package repo

import (
	"context"

	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/types"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserRepo[T any] struct {
	*BaseRepo[T]
}

func NewUserRepo[T any](DB types.AppDB, collection string) *UserRepo[T] {
	return &UserRepo[T]{
		BaseRepo: &BaseRepo[T]{
			DB:         DB,
			Collection: collection,
			UpdatedAt:  true,
			CreatedAt:  true,
		},
	}
}

func (up *UserRepo[T]) GetUserByUsernameOrEmail(usernameOrEmail string) (*T, error) {
	filters := bson.D{{
		Key: "$or", Value: bson.A{
			bson.D{{Key: "username", Value: usernameOrEmail}}, bson.D{{Key: "email", Value: usernameOrEmail}},
		},
	}}

	var result T

	if err := up.DB.MongoDB.Collection(up.Collection).FindOne(context.TODO(), filters).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &result, nil
}
