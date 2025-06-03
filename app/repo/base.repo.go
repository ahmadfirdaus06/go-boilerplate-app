package repo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ahmadfirdaus06/go-boilerplate-backend/app/types"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type BaseRepo[T any] struct {
	DB         types.AppDB
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

func bindBson(input any) (bson.M, error) {
	bsonBytes, err := bson.Marshal(input)

	if err != nil {
		return nil, err
	}

	var updateMap bson.M
	err = bson.Unmarshal(bsonBytes, &updateMap)
	if err != nil {
		return nil, err
	}

	return updateMap, nil
}

func (r *BaseRepo[T]) Create(data any) (*T, error) {
	parsed, parsedErr := bindBson(data)

	if parsedErr != nil {
		return nil, parsedErr
	}

	now := time.Now()

	if r.CreatedAt && r.UpdatedAt {
		parsed["updatedAt"] = now
		parsed["createdAt"] = now
	} else {
		if r.CreatedAt {
			parsed["createdAt"] = now
		}

		if r.UpdatedAt {
			parsed["createdAt"] = now
		}
	}

	var (
		typed  T
		output T
	)

	if bindErr := bindData(parsed, &typed); bindErr != nil {
		return nil, bindErr
	}

	created, err := r.DB.MongoDB.Collection(r.Collection).InsertOne(context.TODO(), typed)
	if err != nil {
		return nil, err
	}

	createdObj, err := r.GetByID(created.InsertedID)

	if err != nil {
		return nil, err
	}

	if bindErr := bindData(createdObj, &output); bindErr != nil {
		return nil, bindErr
	}

	return &output, nil
}

func (r *BaseRepo[T]) GetAll(paginated bool, filtersAndSorts *types.GetAllFiltersAndSorts, paginationParams *types.PaginationParams) (*types.PaginatedRecords[T], error) {
	var (
		pipelineStages                  mongo.Pipeline
		pipelineStagesWithoutPagination mongo.Pipeline
		page                                  = 1
		perPage                               = 10
		skip                                  = 0
		total                           int32 = 0
	)

	if filtersAndSorts != nil {
		matchStages := bson.D{}
		for _, item := range filtersAndSorts.QueryParamsFilters {
			switch item.Operator {
			case types.OpLike:
				matchStages = append(matchStages, bson.E{Key: item.Field, Value: bson.D{{
					Key: "$regex", Value: item.Value,
				}, {Key: "$options", Value: "i"}}})
			case types.OpEq:
				matchStages = append(matchStages, bson.E{Key: item.Field, Value: item.Value})
			}
		}

		sortStages := bson.D{}
		for _, item := range filtersAndSorts.QueryParamsSortFields {
			if item.Field != "" {

				value := 1

				if item.Descending {
					value = -1
				}

				sortStages = append(sortStages, bson.E{Key: item.Field, Value: value})
			}
		}

		if len(matchStages) > 0 {
			pipelineStagesWithoutPagination = append(pipelineStagesWithoutPagination, bson.D{{Key: "$match", Value: matchStages}})
		}

		if len(sortStages) > 0 {
			pipelineStagesWithoutPagination = append(pipelineStagesWithoutPagination, bson.D{{Key: "$sort", Value: sortStages}})
		}

	}

	if paginated {
		if paginationParams.Page != 0 {
			page = paginationParams.Page
		}

		if paginationParams.PerPage != 0 {
			perPage = paginationParams.PerPage
		}

		skip = (page - 1) * perPage

		pipelineStages = append(pipelineStagesWithoutPagination, bson.D{{
			Key: "$skip", Value: skip,
		}}, bson.D{{Key: "$limit", Value: perPage}})
	} else {
		pipelineStages = pipelineStagesWithoutPagination
	}

	results, err := r.DB.MongoDB.Collection(r.Collection).Aggregate(context.TODO(), pipelineStages)
	if err != nil {
		return nil, err
	}

	if paginated {
		if countResults, err := r.DB.MongoDB.Collection(r.Collection).Aggregate(context.TODO(), append(pipelineStagesWithoutPagination, bson.D{{Key: "$count", Value: "count"}})); err != nil {
			return nil, err
		} else {
			var result bson.M

			if ok := countResults.Next(context.TODO()); !ok {
				total = 0
			} else {
				if err := countResults.Decode(&result); err != nil {
					return nil, err
				} else {
					total = result["count"].(int32)
				}
			}
		}
	}

	var typedResults []T
	for results.Next(context.TODO()) {
		var result T
		if err := results.Decode(&result); err != nil {
			return nil, err
		}
		typedResults = append(typedResults, result)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	if typedResults == nil {
		typedResults = []T{}
	}

	paginatedRecords := &types.PaginatedRecords[T]{
		Records:    typedResults,
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: int((int64(total) + int64(perPage) - 1) / int64(perPage)),
	}

	return paginatedRecords, nil
}

func (r *BaseRepo[T]) GetByID(id any) (*T, error) {
	if _, ok := id.(bson.ObjectID); !ok {
		if objectId, err := bson.ObjectIDFromHex(id.(string)); err != nil {
			return nil, err
		} else {
			id = objectId
		}
	}

	var result T

	if err := r.DB.MongoDB.Collection(r.Collection).FindOne(context.TODO(), bson.M{"_id": id}).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		} else {
			return nil, err
		}
	} else {
		return &result, nil
	}
}

func (r *BaseRepo[T]) UpdateByID(id any, data any) (*T, error) {
	if _, ok := id.(bson.ObjectID); !ok {
		if objectId, err := bson.ObjectIDFromHex(id.(string)); err != nil {
			return nil, err
		} else {
			id = objectId
		}
	}

	parsed, parsedErr := bindBson(data)

	if parsedErr != nil {
		return nil, parsedErr
	}

	if r.UpdatedAt {
		parsed["updatedAt"] = time.Now()
	}

	if _, err := r.DB.MongoDB.Collection(r.Collection).UpdateByID(context.TODO(), id, bson.D{{Key: "$set", Value: parsed}}); err != nil {
		return nil, err
	}

	updated, updateErr := r.GetByID(id)

	if updateErr != nil {
		return nil, updateErr
	}

	var typed T

	if bindErr := bindData(updated, &typed); bindErr != nil {
		return nil, bindErr
	}

	return &typed, nil
}

func (r *BaseRepo[T]) DeleteByID(id any) (bool, error) {
	if _, ok := id.(bson.ObjectID); !ok {
		if objectId, err := bson.ObjectIDFromHex(id.(string)); err != nil {
			return false, err
		} else {
			id = objectId
		}
	}

	filter := bson.M{"_id": id}

	result, err := r.DB.MongoDB.Collection(r.Collection).DeleteOne(context.TODO(), filter)

	if err != nil {
		return false, err
	}

	if result.DeletedCount > 0 {
		return true, nil
	} else {
		return false, nil
	}
}
