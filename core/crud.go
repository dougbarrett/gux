package core

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// DTOMapper is implemented by DTOs that can map from a model.
type DTOMapper interface {
	FromModel(model interface{}) interface{}
}

// CRUDOption configures a CRUD registration.
type CRUDOption func(*CRUDModel)

// CreateHook is called before creating a model, allowing transformation of input data.
// The hook receives the decoded JSON as a map and returns the model to create.
type CreateHook func(data map[string]interface{}) (interface{}, error)

// UpdateHook is called before updating a model.
// The hook receives the existing model and decoded JSON, returns the model to save.
type UpdateHook func(existing interface{}, data map[string]interface{}) (interface{}, error)

// CRUDModel represents a registered CRUD model
type CRUDModel struct {
	Name           string       // e.g., "Counter"
	Path           string       // e.g., "counters"
	ModelType      reflect.Type // The struct type
	SliceType      reflect.Type // Slice of the struct type
	ListDTO        reflect.Type // Optional DTO for list responses
	DetailDTO      reflect.Type // Optional DTO for single item responses
	ListPreloads   []string     // Preloads for list queries
	DetailPreloads []string     // Preloads for detail queries
	OnCreate       CreateHook   // Optional hook for custom create logic
	OnUpdate       UpdateHook   // Optional hook for custom update logic
}

// WithListDTO sets a DTO type for list responses.
// The DTO must implement DTOMapper interface.
func WithListDTO(dto interface{}, preloads ...string) CRUDOption {
	return func(m *CRUDModel) {
		t := reflect.TypeOf(dto)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		m.ListDTO = t
		m.ListPreloads = preloads
	}
}

// WithDetailDTO sets a DTO type for single item responses.
// The DTO must implement DTOMapper interface.
func WithDetailDTO(dto interface{}, preloads ...string) CRUDOption {
	return func(m *CRUDModel) {
		t := reflect.TypeOf(dto)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		m.DetailDTO = t
		m.DetailPreloads = preloads
	}
}

// WithDTO sets the same DTO type for both list and detail responses.
func WithDTO(dto interface{}, preloads ...string) CRUDOption {
	return func(m *CRUDModel) {
		t := reflect.TypeOf(dto)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		m.ListDTO = t
		m.DetailDTO = t
		m.ListPreloads = preloads
		m.DetailPreloads = preloads
	}
}

// WithCreateHook sets a custom hook for create operations.
// Use this for password hashing, validation, or other pre-processing.
func WithCreateHook(hook CreateHook) CRUDOption {
	return func(m *CRUDModel) {
		m.OnCreate = hook
	}
}

// WithUpdateHook sets a custom hook for update operations.
func WithUpdateHook(hook UpdateHook) CRUDOption {
	return func(m *CRUDModel) {
		m.OnUpdate = hook
	}
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
// With DTOs: app.CRUD(models.User{}, core.WithListDTO(dto.UserList{}), core.WithDetailDTO(dto.UserDetail{}, "Posts"))
// Creates: GET/POST /__gux_api/crud/counters, GET/PUT/DELETE /__gux_api/crud/counters/:id
func (a *App) CRUD(model interface{}, opts ...CRUDOption) *App {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.Name()
	path := strings.ToLower(name) + "s" // Simple pluralization

	m := CRUDModel{
		Name:      name,
		Path:      path,
		ModelType: t,
		SliceType: reflect.SliceOf(t),
	}

	// Apply options
	for _, opt := range opts {
		opt(&m)
	}

	a.crudModels = append(a.crudModels, m)

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

	// Start with the db value
	dbVal := reflect.ValueOf(a.db)

	// Apply preloads for list queries
	for _, preload := range model.ListPreloads {
		preloadMethod := dbVal.MethodByName("Preload")
		if preloadMethod.IsValid() {
			ret := preloadMethod.Call([]reflect.Value{reflect.ValueOf(preload)})
			if len(ret) > 0 {
				dbVal = ret[0]
			}
		}
	}

	// Use reflection to call db.Find(results)
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

	// Convert to DTO if configured
	output := reflect.ValueOf(results).Elem().Interface()
	if model.ListDTO != nil {
		output = a.convertToDTO(output, model.ListDTO)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func (a *App) handleGet(w http.ResponseWriter, r *http.Request, model CRUDModel, id uint) {
	if a.db == nil {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	// Create instance to hold result
	result := reflect.New(model.ModelType).Interface()

	// Start with the db value
	dbVal := reflect.ValueOf(a.db)

	// Apply preloads for detail queries
	for _, preload := range model.DetailPreloads {
		preloadMethod := dbVal.MethodByName("Preload")
		if preloadMethod.IsValid() {
			ret := preloadMethod.Call([]reflect.Value{reflect.ValueOf(preload)})
			if len(ret) > 0 {
				dbVal = ret[0]
			}
		}
	}

	// Call db.First(result, id)
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

	// Convert to DTO if configured
	output := reflect.ValueOf(result).Elem().Interface()
	if model.DetailDTO != nil {
		output = a.convertSingleToDTO(output, model.DetailDTO)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

// convertToDTO converts a slice of models to a slice of DTOs.
func (a *App) convertToDTO(models interface{}, dtoType reflect.Type) interface{} {
	modelsVal := reflect.ValueOf(models)
	if modelsVal.Kind() != reflect.Slice {
		return models
	}

	// Create slice of DTOs
	dtoSlice := reflect.MakeSlice(reflect.SliceOf(dtoType), 0, modelsVal.Len())

	for i := 0; i < modelsVal.Len(); i++ {
		model := modelsVal.Index(i).Interface()
		dto := a.convertSingleToDTO(model, dtoType)
		dtoSlice = reflect.Append(dtoSlice, reflect.ValueOf(dto))
	}

	return dtoSlice.Interface()
}

// convertSingleToDTO converts a single model to a DTO.
func (a *App) convertSingleToDTO(model interface{}, dtoType reflect.Type) interface{} {
	// Create new DTO instance
	dtoPtr := reflect.New(dtoType)
	dto := dtoPtr.Interface()

	// Check if DTO implements DTOMapper
	if mapper, ok := dto.(DTOMapper); ok {
		return mapper.FromModel(model)
	}

	// Auto-map matching fields by name and type, or by gux tag
	modelVal := reflect.ValueOf(model)
	if modelVal.Kind() == reflect.Ptr {
		modelVal = modelVal.Elem()
	}
	dtoVal := dtoPtr.Elem()

	for i := 0; i < dtoType.NumField(); i++ {
		dtoField := dtoType.Field(i)

		// Check for gux tag to map from a different field name
		// Format: gux:"ModelName.FieldName" or gux:"FieldName"
		var modelFieldName string
		if guxTag := dtoField.Tag.Get("gux"); guxTag != "" {
			// Extract field name after the dot (e.g., "Post.User" -> "User")
			if dotIdx := strings.LastIndex(guxTag, "."); dotIdx >= 0 {
				modelFieldName = guxTag[dotIdx+1:]
			} else {
				modelFieldName = guxTag
			}
		} else {
			modelFieldName = dtoField.Name
		}

		modelField := modelVal.FieldByName(modelFieldName)

		if modelField.IsValid() {
			// Handle pointer fields - dereference if model field is pointer
			srcVal := modelField
			if srcVal.Kind() == reflect.Ptr && !srcVal.IsNil() {
				srcVal = srcVal.Elem()
			}

			// First check if types are directly assignable (handles time.Time, etc.)
			if modelField.Type().AssignableTo(dtoField.Type) {
				dtoVal.Field(i).Set(modelField)
			} else if srcVal.Kind() == reflect.Struct && dtoField.Type.Kind() == reflect.Struct {
				// Only recursively convert if the types are different (custom structs)
				// Skip standard library types like time.Time which have unexported fields
				srcType := srcVal.Type()
				if srcType.PkgPath() != "" && !strings.HasPrefix(srcType.PkgPath(), "time") {
					nestedDTO := a.convertSingleToDTO(srcVal.Interface(), dtoField.Type)
					dtoVal.Field(i).Set(reflect.ValueOf(nestedDTO))
				}
			}
		}
	}

	return dtoVal.Interface()
}

func (a *App) handleCreate(w http.ResponseWriter, r *http.Request, model CRUDModel) {
	if a.db == nil {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	var item interface{}

	if model.OnCreate != nil {
		// Use custom hook - decode as map first
		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		var err error
		item, err = model.OnCreate(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Default behavior - decode directly into model
		item = reflect.New(model.ModelType).Interface()
		if err := json.NewDecoder(r.Body).Decode(item); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
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
				// Get actual error message
				errInterface := errVal[0].Interface()
				if err, ok := errInterface.(error); ok {
					http.Error(w, "Create failed: "+err.Error(), http.StatusInternalServerError)
				} else {
					http.Error(w, "Create failed", http.StatusInternalServerError)
				}
				return
			}
		}
	}

	// Convert response to DTO if configured
	output := reflect.ValueOf(item).Elem().Interface()
	if model.DetailDTO != nil {
		output = a.convertSingleToDTO(output, model.DetailDTO)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(output)
}

func (a *App) handleUpdate(w http.ResponseWriter, r *http.Request, model CRUDModel, id uint) {
	if a.db == nil {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	var item interface{}
	dbVal := reflect.ValueOf(a.db)

	if model.OnUpdate != nil {
		// Fetch existing model first
		existing := reflect.New(model.ModelType).Interface()
		firstMethod := dbVal.MethodByName("First")
		if firstMethod.IsValid() {
			ret := firstMethod.Call([]reflect.Value{reflect.ValueOf(existing), reflect.ValueOf(id)})
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
		}

		// Decode JSON as map and call hook
		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		var err error
		item, err = model.OnUpdate(existing, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Default behavior
		item = reflect.New(model.ModelType).Interface()
		if err := json.NewDecoder(r.Body).Decode(item); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		// Set ID on the item
		itemVal := reflect.ValueOf(item).Elem()
		idField := itemVal.FieldByName("ID")
		if idField.IsValid() && idField.CanSet() {
			idField.SetUint(uint64(id))
		}
	}

	// Call db.Save(item)
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

	// Convert response to DTO if configured
	output := reflect.ValueOf(item).Elem().Interface()
	if model.DetailDTO != nil {
		output = a.convertSingleToDTO(output, model.DetailDTO)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
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
