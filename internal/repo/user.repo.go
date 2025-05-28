package repo

import (
	"go-boilerplate-backend/internal/externals"
)

type UserRepo[T any] struct {
	*BaseRepo[T]
}

func NewUserRepo[T any](DB *externals.JsonDBExternal, collection string) *UserRepo[T] {
	return &UserRepo[T]{
		BaseRepo: &BaseRepo[T]{
			DB:         DB,
			Collection: collection,
			UpdatedAt:  true,
			CreatedAt:  true,
		},
	}
}
