package externals_test

import (
	"fmt"
	"go-boilerplate-backend/internal/externals"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	dbInstance := &externals.JsonDBExternal{DBName: "db_test.json"}

	_, err := dbInstance.Connect()

	if err != nil {
		t.Fatalf("Failed to connect to JsonDBExternal: %v", err)
	}

	os.Remove(dbInstance.DBName)
}

func TestCreate(t *testing.T) {
	var dbName = "db_test.json"

	_ = os.Remove(dbName)

	dbInstance := &externals.JsonDBExternal{DBName: dbName}

	dbInstance, _ = dbInstance.Connect()

	someItem := map[string]interface{}{
		"name": "Test Item",
		"type": "example",
	}

	if _, createErr := dbInstance.Create(&externals.JsonDBCreateConfig{Collection: "items", UpdatedAt: true, CreatedAt: true}, someItem); createErr != nil {
		t.Fatalf("Failed to create item: %v", createErr)
	}

	items, _ := dbInstance.GetAll("items", nil)

	assert.Len(t, items, 1)

	_ = os.Remove(dbName)
}

func TestGetAll(t *testing.T) {
	var (
		dbName        = "db_test.json"
		numberOfItems = 5
	)

	_ = os.Remove(dbName)

	dbInstance := &externals.JsonDBExternal{DBName: dbName}

	dbInstance, _ = dbInstance.Connect()

	for i := 0; i < numberOfItems; i++ {
		someItem := map[string]interface{}{
			"name": fmt.Sprintf("Test Item %d", i),
			"type": "example",
		}
		dbInstance.Create(&externals.JsonDBCreateConfig{Collection: "items"}, someItem)
	}

	items, getAllErr := dbInstance.GetAll("items", nil)

	if getAllErr != nil {
		t.Fatalf("Failed to get all items: %v", getAllErr)
	}

	assert.Len(t, items, numberOfItems)

	_ = os.Remove(dbName)
}

func TestGetByID(t *testing.T) {
	var (
		dbName = "db_test.json"
	)

	_ = os.Remove(dbName)

	dbInstance := &externals.JsonDBExternal{DBName: dbName}

	dbInstance, _ = dbInstance.Connect()

	someItem := map[string]interface{}{
		"name": "Test Item",
		"type": "example",
	}

	created, _ := dbInstance.Create(&externals.JsonDBCreateConfig{Collection: "items"}, someItem)

	item, getByIDErr := dbInstance.GetByID("items", created.(map[string]interface{})["id"].(string))

	if getByIDErr != nil {
		t.Fatalf("Failed to get item by ID: %v", getByIDErr)
	}

	assert.NotNil(t, item)

	_ = os.Remove(dbName)
}

func TestUpdateByID(t *testing.T) {
	var (
		dbName      = "db_test.json"
		updatedName = "updated name"
	)

	_ = os.Remove(dbName)

	dbInstance := &externals.JsonDBExternal{DBName: dbName}

	dbInstance, _ = dbInstance.Connect()

	someItem := map[string]interface{}{
		"name": "Test Item",
		"type": "example",
	}

	created, _ := dbInstance.Create(&externals.JsonDBCreateConfig{Collection: "items"}, someItem)

	item, updateByIDErr := dbInstance.UpdateByID(&externals.JsonDBUpdateByIdConfig{
		Collection: "items",
	}, created.(map[string]interface{})["id"].(string), map[string]interface{}{"name": updatedName})

	if updateByIDErr != nil {
		t.Fatalf("Failed to update item by ID: %v", updateByIDErr)
	}

	assert.NotNil(t, item)
	assert.Equal(t, item.(map[string]interface{})["name"], updatedName)

	_ = os.Remove(dbName)
}

func TestDeleteByID(t *testing.T) {
	var (
		dbName = "db_test.json"
	)

	_ = os.Remove(dbName)

	dbInstance := &externals.JsonDBExternal{DBName: dbName}

	dbInstance, _ = dbInstance.Connect()

	someItem := map[string]interface{}{
		"name": "Test Item",
		"type": "example",
	}

	created, _ := dbInstance.Create(&externals.JsonDBCreateConfig{Collection: "items"}, someItem)

	deleted, deleteByIDErr := dbInstance.DeleteByID("items", created.(map[string]interface{})["id"].(string))

	if deleteByIDErr != nil {
		t.Fatalf("Failed to delete item by ID: %v", deleteByIDErr)
	}

	assert.True(t, deleted)

	_ = os.Remove(dbName)
}
