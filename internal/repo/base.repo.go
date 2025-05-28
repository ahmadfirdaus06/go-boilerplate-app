package repo

import (
	"encoding/json"
	"go-boilerplate-backend/internal/externals"
)

type BaseRepo[T any] struct {
	DB         *externals.JsonDBExternal
	Collection string
	UpdatedAt  bool
	CreatedAt  bool
}

func bindData(input any, outputSchema any) error {
	bytes, err := json.Marshal(input)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &outputSchema); err != nil {
		return err
	}

	return nil
}

func (r *BaseRepo[T]) Create(data any) (*T, error) {
	var (
		typed  T
		output T
	)

	if bindErr := bindData(data, &typed); bindErr != nil {
		return nil, bindErr
	}

	created, err := r.DB.Create(&externals.JsonDBCreateConfig{Collection: r.Collection, UpdatedAt: r.UpdatedAt, CreatedAt: r.CreatedAt}, typed)
	if err != nil {
		return nil, err
	}

	if bindErr := bindData(created, &output); bindErr != nil {
		return nil, bindErr
	}

	return &output, nil
}

func (r *BaseRepo[T]) GetAll(pipeline []externals.JsonDBPipelineStage) ([]T, error) {
	results, err := r.DB.GetAll(r.Collection, pipeline)
	if err != nil {
		return nil, err
	}

	var typedResults []T
	for _, result := range results {
		var typed T

		if bindErr := bindData(result, &typed); bindErr != nil {
			return nil, bindErr
		}

		typedResults = append(typedResults, typed)
	}

	return typedResults, nil
}

func (r *BaseRepo[T]) GetByID(id string) (*T, error) {
	result, err := r.DB.GetByID(r.Collection, id)

	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	var typed T

	if bindErr := bindData(result, &typed); bindErr != nil {
		return nil, bindErr
	}

	return &typed, nil
}

func (r *BaseRepo[T]) UpdateByID(id string, data interface{}) (*T, error) {
	result, err := r.DB.UpdateByID(&externals.JsonDBUpdateByIdConfig{
		Collection: r.Collection,
		UpdatedAt:  r.UpdatedAt,
	}, id, data)

	if err != nil {
		return nil, err
	}

	var typed T

	if bindErr := bindData(result, &typed); bindErr != nil {
		return nil, bindErr
	}

	return &typed, nil
}

func (r *BaseRepo[T]) DeleteByID(id string) (bool, error) {
	deleted, err := r.DB.DeleteByID(r.Collection, id)
	if err != nil {
		return false, err
	}
	return deleted, nil
}
