package externals

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type (
	JsonDBExternal struct {
		mu     sync.Mutex
		data   map[string][]map[string]interface{}
		DBName string
	}
)

func (a *JsonDBExternal) Connect() (*JsonDBExternal, error) {
	if a.DBName == "" {
		a.DBName = "db.json"
	}

	file, err := os.ReadFile(a.DBName)
	if err != nil {
		if os.IsNotExist(err) {
			a.data = make(map[string][]map[string]interface{})
			return a, a.SaveToFile()
		}
		return nil, fmt.Errorf("failed to read db file: %w", err)
	}

	if err := json.Unmarshal(file, &a.data); err != nil {
		return nil, fmt.Errorf("failed to parse db file: %w", err)
	}

	return a, nil
}

func (a *JsonDBExternal) HealthCheck() error {
	_, err := a.Connect()

	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

func (a *JsonDBExternal) SaveToFile() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	file, _ := json.MarshalIndent(a.data, "", "  ")
	return os.WriteFile(a.DBName, file, 0644)
}

func (a *JsonDBExternal) LoadFromFile() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	file, err := os.ReadFile(a.DBName)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &a.data)
}

type JsonDBCreateConfig struct {
	Collection string
	UpdatedAt  bool
	CreatedAt  bool
}

type JsonDBPipelineStage struct {
	Op    string
	Value map[string]interface{}
}

func (a *JsonDBExternal) Create(config *JsonDBCreateConfig, value any) (any, error) {
	if _, ok := a.data[config.Collection]; !ok {
		a.data[config.Collection] = []map[string]interface{}{}
	}

	// Convert value (struct or map) to map[string]interface{}
	var raw map[string]interface{}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	// Add UUID
	id := uuid.New()
	raw["id"] = id.String()

	if config.CreatedAt && config.UpdatedAt {
		now := time.Now()
		raw["createdAt"] = now
		raw["updatedAt"] = now
	} else {
		if config.CreatedAt {
			raw["createdAt"] = time.Now()
		}

		if config.UpdatedAt {
			raw["updatedAt"] = time.Now()
		}
	}

	a.data[config.Collection] = append(a.data[config.Collection], raw)

	if err := a.SaveToFile(); err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}

	return raw, nil
}

func matchStage(data []map[string]interface{}, filter map[string]interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	for _, entry := range data {
		if matchesFilter(entry, filter) {
			out = append(out, entry)
		}
	}
	return out
}

func matchesFilter(entry map[string]interface{}, filter map[string]interface{}) bool {
	for key, val := range filter {
		// Handle $or as a special case
		if key == "$or" {
			conditions, ok := val.([]interface{})
			if !ok {
				return false
			}
			match := false
			for _, cond := range conditions {
				if subFilter, ok := cond.(map[string]interface{}); ok {
					if matchesFilter(entry, subFilter) {
						match = true
						break
					}
				}
			}
			if !match {
				return false
			}
		} else {
			if !matchCondition(entry, key, val) {
				return false
			}
		}
	}
	return true
}

func matchCondition(entry map[string]interface{}, key string, condition interface{}) bool {
	entryVal, entryOk := entry[key]
	if !entryOk {
		return false
	}

	switch cond := condition.(type) {
	case map[string]interface{}:
		// Support $like
		if like, ok := cond["$like"]; ok {
			entryStr, ok := entryVal.(string)
			if !ok {
				return false
			}
			return strings.Contains(strings.ToLower(entryStr), strings.ToLower(fmt.Sprint(like)))
		}
		return false // unsupported nested condition
	default:
		// Direct equality
		return entryVal == cond
	}
}

func (a *JsonDBExternal) GetAll(collection string, pipeline []JsonDBPipelineStage) ([]map[string]interface{}, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, ok := a.data[collection]
	if !ok {
		return nil, nil
	}

	// If no pipeline provided, return all data
	if len(pipeline) == 0 {
		// Deep copy to avoid mutation
		results := make([]map[string]interface{}, len(data))
		for i, item := range data {
			copy := make(map[string]interface{})
			for k, v := range item {
				copy[k] = v
			}
			results[i] = copy
		}
		return results, nil
	}

	// Apply each pipeline stage
	result := data
	for _, stage := range pipeline {
		switch stage.Op {
		case "$match":
			result = matchStage(result, stage.Value)
		default:
			return nil, fmt.Errorf("unsupported stage: %s", stage.Op)
		}
	}

	return result, nil
}

func (a *JsonDBExternal) GetByID(collection, id string) (interface{}, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.data[collection]; !ok {
		return nil, nil
	}

	// Search for the entry with the given ID
	for _, value := range a.data[collection] {
		if valueID, ok := value["id"].(string); ok && valueID == id {
			return value, nil
		}
	}

	return nil, nil
}

type JsonDBUpdateByIdConfig struct {
	Collection string
	UpdatedAt  bool
}

func (a *JsonDBExternal) UpdateByID(config *JsonDBUpdateByIdConfig, id string, value any) (interface{}, error) {
	if _, ok := a.data[config.Collection]; !ok {
		return nil, nil
	}

	// Convert `value` (could be struct or map) to map[string]interface{}
	var updates map[string]interface{}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}
	if err := json.Unmarshal(data, &updates); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	for i, entry := range a.data[config.Collection] {
		if entryID, ok := entry["id"].(string); ok && entryID == id {
			for k, v := range updates {
				// Only overwrite provided fields
				if k == "id" || k == "createdAt" {
					continue // these fields should never be updated
				}
				entry[k] = v
			}

			if config.UpdatedAt {
				entry["updatedAt"] = time.Now()
			}

			a.data[config.Collection][i] = entry

			if err := a.SaveToFile(); err != nil {
				return nil, fmt.Errorf("write failed: %w", err)
			}
			return entry, nil
		}
	}

	return nil, nil
}

func (a *JsonDBExternal) DeleteByID(collection, id string) (bool, error) {
	if _, ok := a.data[collection]; !ok {
		return false, nil
	}

	for i, entry := range a.data[collection] {
		if entryID, ok := entry["id"].(string); ok && entryID == id {
			// Remove the entry from the collection
			a.data[collection] = append(a.data[collection][:i], a.data[collection][i+1:]...)
			if err := a.SaveToFile(); err != nil {
				return false, fmt.Errorf("write failed: %w", err)
			}
			return true, nil
		}

	}

	return false, nil
}
