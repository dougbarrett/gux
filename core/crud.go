package core

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/dougbarrett/gux/api"
)

// pluralize converts a singular word to its plural form.
// Handles common English pluralization rules.
func pluralize(word string) string {
	if word == "" {
		return word
	}

	lower := strings.ToLower(word)

	// Words ending in s, x, z, ch, sh → add "es"
	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z") || strings.HasSuffix(lower, "ch") ||
		strings.HasSuffix(lower, "sh") {
		return word + "es"
	}

	// Words ending in consonant + y → change y to ies
	if strings.HasSuffix(lower, "y") && len(word) > 1 {
		prev := lower[len(lower)-2]
		// Check if previous char is a consonant (not a, e, i, o, u)
		if prev != 'a' && prev != 'e' && prev != 'i' && prev != 'o' && prev != 'u' {
			return word[:len(word)-1] + "ies"
		}
	}

	// Default: add "s"
	return word + "s"
}

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
	Public         bool         // If true, no authentication required
	Roles          []string     // Roles required to access (any of these roles)
	AuditConfig    *AuditConfig // Audit logging config (nil = disabled)
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

// WithPublic marks this CRUD endpoint as public (no authentication required).
// By default, all CRUD endpoints require authentication.
func WithPublic() CRUDOption {
	return func(m *CRUDModel) {
		m.Public = true
	}
}

// WithRoles specifies required roles for this CRUD endpoint.
// User must have ANY of the specified roles (OR logic).
// CRUD endpoints are protected by default, so this just adds role restrictions.
func WithRoles(roles ...string) CRUDOption {
	return func(m *CRUDModel) {
		m.Roles = roles
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

// getUserFromRequest extracts the authenticated user from the request session.
// Returns nil if not authenticated or auth is not configured.
func (a *App) getUserFromRequest(r *http.Request) *SessionUser {
	if a.authConfig == nil || a.authConfig.SessionStore == nil {
		return nil
	}
	cookieName := a.authConfig.CookieName
	if cookieName == "" {
		cookieName = DefaultSessionCookieName
	}
	sessionID := getSessionIDFromCookie(r, cookieName)
	if sessionID == "" {
		return nil
	}
	user, _ := a.authConfig.SessionStore.Get(sessionID)
	return user
}

// checkCRUDAuth checks if the request is authorized for a CRUD model.
// Returns an error if not authorized, nil if authorized.
func (a *App) checkCRUDAuth(w http.ResponseWriter, r *http.Request, model CRUDModel) bool {
	// Public endpoints don't require auth
	if model.Public {
		return true
	}

	// If auth is not configured, allow access (backwards compatibility)
	if a.authConfig == nil || a.authConfig.SessionStore == nil {
		return true
	}

	user := a.getUserFromRequest(r)
	if user == nil {
		api.WriteError(w, api.Unauthorized("authentication required"))
		return false
	}

	// Check roles if specified
	if len(model.Roles) > 0 && !user.HasAnyRole(model.Roles...) {
		api.WriteError(w, api.Forbidden("insufficient permissions"))
		return false
	}

	return true
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
	path := strings.ToLower(pluralize(name)) // Proper pluralization (industry → industries)

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
	// Auto-migrate audit_entries table if any model uses audit logging
	if !a.auditMigrated {
		for _, model := range a.crudModels {
			if model.AuditConfig != nil && model.AuditConfig.Enabled {
				a.autoMigrateAudit()
				a.auditMigrated = true
				break
			}
		}
	}
	for _, model := range a.crudModels {
		a.registerModelHandlers(mux, model)
	}
}

func (a *App) registerModelHandlers(mux *http.ServeMux, model CRUDModel) {
	basePath := a.apiPrefix + "/crud/" + model.Path

	// List and Create: /__gux_api/crud/counters
	mux.HandleFunc(basePath, func(w http.ResponseWriter, r *http.Request) {
		// Auth check
		if !a.checkCRUDAuth(w, r, model) {
			return
		}

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
		// Auth check
		if !a.checkCRUDAuth(w, r, model) {
			return
		}

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

	// Apply query parameter filters (e.g., ?order_id=5&active=true)
	dbVal = applyQueryFilters(dbVal, r, model.ModelType)

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

			// Get the actual DTO field type (dereference if pointer)
			dtoFieldType := dtoField.Type
			dtoFieldIsPtr := dtoFieldType.Kind() == reflect.Ptr
			if dtoFieldIsPtr {
				dtoFieldType = dtoFieldType.Elem()
			}

			// First check if types are directly assignable (handles time.Time, etc.)
			if modelField.Type().AssignableTo(dtoField.Type) {
				dtoVal.Field(i).Set(modelField)
			} else if srcVal.Kind() == reflect.Struct && dtoFieldType.Kind() == reflect.Struct {
				// Handle struct-to-struct conversion (including pointer DTO fields)
				// Skip standard library types like time.Time which have unexported fields
				srcType := srcVal.Type()
				if srcType.PkgPath() != "" && !strings.HasPrefix(srcType.PkgPath(), "time") {
					nestedDTO := a.convertSingleToDTO(srcVal.Interface(), dtoFieldType)
					if dtoFieldIsPtr {
						// DTO field is a pointer - create a pointer to the converted value
						nestedPtr := reflect.New(dtoFieldType)
						nestedPtr.Elem().Set(reflect.ValueOf(nestedDTO))
						dtoVal.Field(i).Set(nestedPtr)
					} else {
						dtoVal.Field(i).Set(reflect.ValueOf(nestedDTO))
					}
				}
			} else if srcVal.Kind() == reflect.Slice && dtoFieldType.Kind() == reflect.Slice {
				// Handle slice conversion (e.g., []Post -> []PostDTO)
				srcElemType := srcVal.Type().Elem()
				dtoElemType := dtoFieldType.Elem()

				// Check if it's a slice of structs
				if srcElemType.Kind() == reflect.Struct && dtoElemType.Kind() == reflect.Struct {
					if srcElemType.PkgPath() != "" && !strings.HasPrefix(srcElemType.PkgPath(), "time") {
						convertedSlice := a.convertToDTO(srcVal.Interface(), dtoElemType)
						if dtoFieldIsPtr {
							slicePtr := reflect.New(dtoFieldType)
							slicePtr.Elem().Set(reflect.ValueOf(convertedSlice))
							dtoVal.Field(i).Set(slicePtr)
						} else {
							dtoVal.Field(i).Set(reflect.ValueOf(convertedSlice))
						}
					}
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

	// Audit logging for create
	if model.AuditConfig != nil && model.AuditConfig.Enabled {
		user := a.getUserFromRequest(r)
		entry := AuditEntry{
			Action:     AuditActionCreate,
			EntityType: model.Name,
			EntityID:   extractEntityID(item),
			IPAddress:  getClientIP(r),
			Changes:    snapshotForAudit(item, model.AuditConfig.IgnoreFields),
		}
		if user != nil {
			entry.UserID = user.ID
			entry.UserEmail = user.Email
		}
		go a.writeAuditLog(entry)
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
	var auditBeforeJSON []byte // snapshot for audit diff
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

		// Capture before-state for audit diff
		if model.AuditConfig != nil && model.AuditConfig.Enabled {
			auditBeforeJSON, _ = json.Marshal(reflect.ValueOf(existing).Elem().Interface())
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
		// Default behavior: fetch existing record first, then merge JSON fields.
		// This preserves CreatedAt and other fields not included in the request.
		item = reflect.New(model.ModelType).Interface()
		firstMethod := dbVal.MethodByName("First")
		if firstMethod.IsValid() {
			ret := firstMethod.Call([]reflect.Value{reflect.ValueOf(item), reflect.ValueOf(id)})
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

		// Capture before-state for audit diff
		if model.AuditConfig != nil && model.AuditConfig.Enabled {
			auditBeforeJSON, _ = json.Marshal(reflect.ValueOf(item).Elem().Interface())
		}

		// Decode JSON onto the existing record — only overwrites fields present in the request
		if err := json.NewDecoder(r.Body).Decode(item); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
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

	// Audit logging for update
	if model.AuditConfig != nil && model.AuditConfig.Enabled && auditBeforeJSON != nil {
		var beforeMap map[string]interface{}
		json.Unmarshal(auditBeforeJSON, &beforeMap)
		afterMap := structToMapAudit(item)
		changes := computeFieldDiffFromMaps(beforeMap, afterMap, model.AuditConfig.IgnoreFields)

		user := a.getUserFromRequest(r)
		entry := AuditEntry{
			Action:     AuditActionUpdate,
			EntityType: model.Name,
			EntityID:   id,
			IPAddress:  getClientIP(r),
			Changes:    changes,
		}
		if user != nil {
			entry.UserID = user.ID
			entry.UserEmail = user.Email
		}
		go a.writeAuditLog(entry)
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

	// Audit logging for delete
	if model.AuditConfig != nil && model.AuditConfig.Enabled {
		user := a.getUserFromRequest(r)
		entry := AuditEntry{
			Action:     AuditActionDelete,
			EntityType: model.Name,
			EntityID:   id,
			IPAddress:  getClientIP(r),
			Changes:    "{}",
		}
		if user != nil {
			entry.UserID = user.ID
			entry.UserEmail = user.Email
		}
		go a.writeAuditLog(entry)
	}

	w.WriteHeader(http.StatusNoContent)
}

// applyQueryFilters applies WHERE conditions to the database query based on
// URL query parameters. Parameters are matched against the model's struct fields
// using JSON tag names (snake_case). Only fields that exist on the model are filtered.
func applyQueryFilters(dbVal reflect.Value, r *http.Request, modelType reflect.Type) reflect.Value {
	query := r.URL.Query()
	if len(query) == 0 {
		return dbVal
	}

	// Build a map of json tag → field info for the model
	type fieldInfo struct {
		Name   string       // Go field name
		Type   reflect.Type // Field type
		Column string       // DB column name (snake_case)
	}
	tagMap := make(map[string]fieldInfo)
	for i := 0; i < modelType.NumField(); i++ {
		f := modelType.Field(i)
		// Handle embedded structs (like gorm.Model)
		if f.Anonymous {
			for j := 0; j < f.Type.NumField(); j++ {
				ef := f.Type.Field(j)
				jsonTag := ef.Tag.Get("json")
				if jsonTag == "" || jsonTag == "-" {
					jsonTag = toSnakeCase(ef.Name)
				} else {
					jsonTag = strings.Split(jsonTag, ",")[0]
				}
				tagMap[jsonTag] = fieldInfo{Name: ef.Name, Type: ef.Type, Column: toSnakeCase(ef.Name)}
			}
			continue
		}
		jsonTag := f.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		if jsonTag == "" {
			jsonTag = toSnakeCase(f.Name)
		} else {
			jsonTag = strings.Split(jsonTag, ",")[0]
		}
		tagMap[jsonTag] = fieldInfo{Name: f.Name, Type: f.Type, Column: toSnakeCase(f.Name)}
	}

	for param, values := range query {
		if len(values) == 0 || values[0] == "" {
			continue
		}
		fi, ok := tagMap[param]
		if !ok {
			continue // Skip unknown parameters
		}

		// Coerce the string value to the correct type
		var val interface{}
		rawVal := values[0]
		switch fi.Type.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if v, err := strconv.ParseUint(rawVal, 10, 64); err == nil {
				val = v
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v, err := strconv.ParseInt(rawVal, 10, 64); err == nil {
				val = v
			}
		case reflect.Float32, reflect.Float64:
			if v, err := strconv.ParseFloat(rawVal, 64); err == nil {
				val = v
			}
		case reflect.Bool:
			if v, err := strconv.ParseBool(rawVal); err == nil {
				val = v
			}
		case reflect.String:
			val = rawVal
		case reflect.Ptr:
			// Handle pointer types (e.g., *uint) — coerce the element type
			elemKind := fi.Type.Elem().Kind()
			switch elemKind {
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if v, err := strconv.ParseUint(rawVal, 10, 64); err == nil {
					val = v
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if v, err := strconv.ParseInt(rawVal, 10, 64); err == nil {
					val = v
				}
			case reflect.String:
				val = rawVal
			case reflect.Bool:
				if v, err := strconv.ParseBool(rawVal); err == nil {
					val = v
				}
			}
		}

		if val == nil {
			continue // Could not coerce value
		}

		// Chain .Where("column = ?", val)
		whereMethod := dbVal.MethodByName("Where")
		if whereMethod.IsValid() {
			condition := fi.Column + " = ?"
			ret := whereMethod.Call([]reflect.Value{
				reflect.ValueOf(condition),
				reflect.ValueOf(val),
			})
			if len(ret) > 0 {
				dbVal = ret[0]
			}
		}
	}

	return dbVal
}

// toSnakeCase converts a Go field name to snake_case for DB column names.
// Handles consecutive uppercase letters (abbreviations) correctly:
// OrderID → order_id, UserID → user_id, HTMLParser → html_parser
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := rune(s[i-1])
			if prev >= 'a' && prev <= 'z' {
				// Lowercase→Uppercase boundary: Order|ID
				result.WriteByte('_')
			} else if i+1 < len(s) {
				next := rune(s[i+1])
				if next >= 'a' && next <= 'z' {
					// Uppercase→Lowercase with preceding uppercase: HTM|L→P|arser
					result.WriteByte('_')
				}
			}
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
