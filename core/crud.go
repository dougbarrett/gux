package core

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// CRUDModel represents a registered CRUD model
type CRUDModel struct {
	Name       string       // e.g., "Counter"
	Path       string       // e.g., "counters"
	ModelType  reflect.Type // The struct type
	SliceType  reflect.Type // Slice of the struct type
}

// DB interface for database operations (compatible with GORM)
type DB interface {
	Find(dest interface{}, conds ...interface{}) DB
	First(dest interface{}, conds ...interface{}) DB
	Create(value interface{}) DB
	Save(value interface{}) DB
	Delete(value interface{}, conds ...interface{}) DB
	Error() error
}

// GormWrapper wraps a GORM DB to implement our DB interface
type GormWrapper struct {
	db    interface{}
	error error
}

// SetDB sets the database connection on the app.
func (a *App) SetDB(db interface{}) {
	a.db = db
}

// GetDB returns the database connection.
func (a *App) GetDB() interface{} {
	return a.db
}

// CRUD registers CRUD endpoints for a model.
// Usage: app.CRUD(models.Counter{})
// Creates: GET/POST /__gux_api/crud/counters, GET/PUT/DELETE /__gux_api/crud/counters/:id
func (a *App) CRUD(model interface{}) *App {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.Name()
	path := strings.ToLower(name) + "s" // Simple pluralization

	a.crudModels = append(a.crudModels, CRUDModel{
		Name:      name,
		Path:      path,
		ModelType: t,
		SliceType: reflect.SliceOf(t),
	})

	return a
}

// registerCRUDHandlers registers HTTP handlers for all CRUD models
func (a *App) registerCRUDHandlers(mux *http.ServeMux) {
	for _, model := range a.crudModels {
		a.registerModelHandlers(mux, model)
	}
}

func (a *App) registerModelHandlers(mux *http.ServeMux, model CRUDModel) {
	basePath := a.apiPrefix + "/crud/" + model.Path

	// List and Create: /__gux_api/crud/counters
	mux.HandleFunc(basePath, func(w http.ResponseWriter, r *http.Request) {
		// Check for trailing path (ID)
		path := r.URL.Path
		if strings.HasPrefix(path, basePath+"/") {
			// Has ID - route to single item handlers
			idStr := strings.TrimPrefix(path, basePath+"/")
			a.handleSingleItem(w, r, model, idStr)
			return
		}

		switch r.Method {
		case http.MethodGet:
			a.handleList(w, r, model)
		case http.MethodPost:
			a.handleCreate(w, r, model)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Single item: /__gux_api/crud/counters/
	mux.HandleFunc(basePath+"/", func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, basePath+"/")
		if idStr == "" {
			http.Error(w, "ID required", http.StatusBadRequest)
			return
		}
		a.handleSingleItem(w, r, model, idStr)
	})
}

func (a *App) handleSingleItem(w http.ResponseWriter, r *http.Request, model CRUDModel, idStr string) {
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		a.handleGet(w, r, model, uint(id))
	case http.MethodPut:
		a.handleUpdate(w, r, model, uint(id))
	case http.MethodDelete:
		a.handleDelete(w, r, model, uint(id))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *App) handleList(w http.ResponseWriter, r *http.Request, model CRUDModel) {
	if a.db == nil {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	// Create slice to hold results
	results := reflect.New(model.SliceType).Interface()

	// Use reflection to call db.Find(results)
	dbVal := reflect.ValueOf(a.db)
	findMethod := dbVal.MethodByName("Find")
	if !findMethod.IsValid() {
		http.Error(w, "Database does not support Find", http.StatusInternalServerError)
		return
	}

	ret := findMethod.Call([]reflect.Value{reflect.ValueOf(results)})
	if len(ret) > 0 {
		// Check for error (GORM returns *gorm.DB, check Error field)
		errField := ret[0].MethodByName("Error")
		if errField.IsValid() {
			errVal := errField.Call(nil)
			if len(errVal) > 0 && !errVal[0].IsNil() {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reflect.ValueOf(results).Elem().Interface())
}

func (a *App) handleGet(w http.ResponseWriter, r *http.Request, model CRUDModel, id uint) {
	if a.db == nil {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	// Create instance to hold result
	result := reflect.New(model.ModelType).Interface()

	// Call db.First(result, id)
	dbVal := reflect.ValueOf(a.db)
	firstMethod := dbVal.MethodByName("First")
	if !firstMethod.IsValid() {
		http.Error(w, "Database does not support First", http.StatusInternalServerError)
		return
	}

	ret := firstMethod.Call([]reflect.Value{reflect.ValueOf(result), reflect.ValueOf(id)})
	if len(ret) > 0 {
		errField := ret[0].MethodByName("Error")
		if errField.IsValid() {
			errVal := errField.Call(nil)
			if len(errVal) > 0 && !errVal[0].IsNil() {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reflect.ValueOf(result).Elem().Interface())
}

func (a *App) handleCreate(w http.ResponseWriter, r *http.Request, model CRUDModel) {
	if a.db == nil {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	// Create instance and decode JSON body
	item := reflect.New(model.ModelType).Interface()
	if err := json.NewDecoder(r.Body).Decode(item); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Call db.Create(item)
	dbVal := reflect.ValueOf(a.db)
	createMethod := dbVal.MethodByName("Create")
	if !createMethod.IsValid() {
		http.Error(w, "Database does not support Create", http.StatusInternalServerError)
		return
	}

	ret := createMethod.Call([]reflect.Value{reflect.ValueOf(item)})
	if len(ret) > 0 {
		errField := ret[0].MethodByName("Error")
		if errField.IsValid() {
			errVal := errField.Call(nil)
			if len(errVal) > 0 && !errVal[0].IsNil() {
				http.Error(w, "Create failed", http.StatusInternalServerError)
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reflect.ValueOf(item).Elem().Interface())
}

func (a *App) handleUpdate(w http.ResponseWriter, r *http.Request, model CRUDModel, id uint) {
	if a.db == nil {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	// Create instance and decode JSON body
	item := reflect.New(model.ModelType).Interface()
	if err := json.NewDecoder(r.Body).Decode(item); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Set ID on the item (assumes ID field exists)
	itemVal := reflect.ValueOf(item).Elem()
	idField := itemVal.FieldByName("ID")
	if idField.IsValid() && idField.CanSet() {
		idField.SetUint(uint64(id))
	}

	// Call db.Save(item)
	dbVal := reflect.ValueOf(a.db)
	saveMethod := dbVal.MethodByName("Save")
	if !saveMethod.IsValid() {
		http.Error(w, "Database does not support Save", http.StatusInternalServerError)
		return
	}

	ret := saveMethod.Call([]reflect.Value{reflect.ValueOf(item)})
	if len(ret) > 0 {
		errField := ret[0].MethodByName("Error")
		if errField.IsValid() {
			errVal := errField.Call(nil)
			if len(errVal) > 0 && !errVal[0].IsNil() {
				http.Error(w, "Update failed", http.StatusInternalServerError)
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reflect.ValueOf(item).Elem().Interface())
}

func (a *App) handleDelete(w http.ResponseWriter, r *http.Request, model CRUDModel, id uint) {
	if a.db == nil {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	// Create instance with ID
	item := reflect.New(model.ModelType).Interface()
	itemVal := reflect.ValueOf(item).Elem()
	idField := itemVal.FieldByName("ID")
	if idField.IsValid() && idField.CanSet() {
		idField.SetUint(uint64(id))
	}

	// Call db.Delete(item)
	dbVal := reflect.ValueOf(a.db)
	deleteMethod := dbVal.MethodByName("Delete")
	if !deleteMethod.IsValid() {
		http.Error(w, "Database does not support Delete", http.StatusInternalServerError)
		return
	}

	ret := deleteMethod.Call([]reflect.Value{reflect.ValueOf(item)})
	if len(ret) > 0 {
		errField := ret[0].MethodByName("Error")
		if errField.IsValid() {
			errVal := errField.Call(nil)
			if len(errVal) > 0 && !errVal[0].IsNil() {
				http.Error(w, "Delete failed", http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
