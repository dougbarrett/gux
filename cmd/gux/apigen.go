package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type InterfaceInfo struct {
	Name       string
	ClientName string
	BasePath   string
	Methods    []MethodInfo
}

// PathParam represents a path parameter with its name and type
type PathParam struct {
	Name   string
	Type   string // "int" or "string"
	IsInt  bool   // convenience field for templates
}

type MethodInfo struct {
	Name       string
	HTTPMethod string
	Path       string
	PathParams []PathParam
	HasBody    bool
	BodyParam  string
	BodyType   string
	ReturnType string
	IsPointer  bool
	IsSlice    bool
	HasReturn  bool
}

// GenerateAPI generates client and server code from a source file
func GenerateAPI(sourceFile, outputFile string) error {
	// Get the directory of the source file
	dir := filepath.Dir(sourceFile)
	if dir == "" {
		dir = "."
	}

	// Parse the source file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, sourceFile, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse file: %w", err)
	}

	// Find interfaces with @client annotation
	interfaces := findInterfaces(node)
	if len(interfaces) == 0 {
		return fmt.Errorf("no interfaces with @client annotation found")
	}

	// Generate client code
	clientCode, err := generateClientCode(interfaces)
	if err != nil {
		return fmt.Errorf("generate client: %w", err)
	}
	clientPath := filepath.Join(dir, outputFile)
	if err := os.WriteFile(clientPath, []byte(clientCode), 0644); err != nil {
		return fmt.Errorf("write client: %w", err)
	}
	fmt.Printf("    generated: %s\n", clientPath)

	// Generate server code
	serverCode, err := generateServerCode(interfaces)
	if err != nil {
		return fmt.Errorf("generate server: %w", err)
	}
	serverOutput := strings.Replace(outputFile, "_client_gen.go", "_server_gen.go", 1)
	serverPath := filepath.Join(dir, serverOutput)
	if err := os.WriteFile(serverPath, []byte(serverCode), 0644); err != nil {
		return fmt.Errorf("write server: %w", err)
	}
	fmt.Printf("    generated: %s\n", serverPath)

	return nil
}

func findInterfaces(node *ast.File) []InterfaceInfo {
	var interfaces []InterfaceInfo

	clientRegex := regexp.MustCompile(`@client\s+(\w+)`)
	basepathRegex := regexp.MustCompile(`@basepath\s+(\S+)`)
	routeRegex := regexp.MustCompile(`@route\s+(GET|POST|PUT|DELETE|PATCH)\s+(\S+)`)

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			ifaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}

			// Check for @client annotation in doc comments
			var clientName, basePath string
			if genDecl.Doc != nil {
				for _, comment := range genDecl.Doc.List {
					if match := clientRegex.FindStringSubmatch(comment.Text); match != nil {
						clientName = match[1]
					}
					if match := basepathRegex.FindStringSubmatch(comment.Text); match != nil {
						basePath = match[1]
					}
				}
			}

			if clientName == "" {
				continue
			}

			info := InterfaceInfo{
				Name:       typeSpec.Name.Name,
				ClientName: clientName,
				BasePath:   basePath,
			}

			// Parse methods
			for _, method := range ifaceType.Methods.List {
				if len(method.Names) == 0 {
					continue
				}

				funcType, ok := method.Type.(*ast.FuncType)
				if !ok {
					continue
				}

				methodInfo := MethodInfo{
					Name: method.Names[0].Name,
				}

				// Parse route annotation from comments
				if method.Doc != nil {
					for _, comment := range method.Doc.List {
						if match := routeRegex.FindStringSubmatch(comment.Text); match != nil {
							methodInfo.HTTPMethod = match[1]
							methodInfo.Path = match[2]
						}
					}
				}

				if methodInfo.HTTPMethod == "" {
					continue
				}

				// Extract path parameter names from the route path
				pathParamRegex := regexp.MustCompile(`\{(\w+)\}`)
				pathParamMatches := pathParamRegex.FindAllStringSubmatch(methodInfo.Path, -1)
				pathParamNames := make(map[string]bool)
				for _, match := range pathParamMatches {
					pathParamNames[match[1]] = true
				}

				// Parse function parameters (skip ctx, identify path params with types, and body param)
				if funcType.Params != nil {
					for i, param := range funcType.Params.List {
						if i == 0 {
							continue // Skip context
						}
						if len(param.Names) == 0 {
							continue
						}

						paramName := param.Names[0].Name
						paramType := exprToString(param.Type)

						if pathParamNames[paramName] {
							// This is a path parameter - store with its type
							isInt := paramType == "int"
							methodInfo.PathParams = append(methodInfo.PathParams, PathParam{
								Name:  paramName,
								Type:  paramType,
								IsInt: isInt,
							})
						} else {
							// Not a path param - must be body
							methodInfo.HasBody = true
							methodInfo.BodyParam = paramName
							methodInfo.BodyType = paramType
						}
					}
				}

				// Parse return type
				if funcType.Results != nil && len(funcType.Results.List) > 0 {
					firstResult := funcType.Results.List[0]
					returnType := exprToString(firstResult.Type)

					// If return is just "error", there's no data return
					if returnType != "error" {
						methodInfo.ReturnType = returnType
						methodInfo.HasReturn = true

						// Check if pointer or slice
						if _, ok := firstResult.Type.(*ast.StarExpr); ok {
							methodInfo.IsPointer = true
						}
						if _, ok := firstResult.Type.(*ast.ArrayType); ok {
							methodInfo.IsSlice = true
						}
					}
				}

				info.Methods = append(info.Methods, methodInfo)
			}

			interfaces = append(interfaces, info)
		}
	}

	return interfaces
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	default:
		return ""
	}
}

// GenerateClientSharedCode generates the shared client types and functions
func GenerateClientSharedCode() (string, error) {
	return `// Code generated by gux. DO NOT EDIT.
//go:build js && wasm

package api

import (
	"encoding/json"
	"fmt"

	"github.com/dougbarrett/gux/fetch"
)

// ClientOption configures a client
type ClientOption func(*clientConfig)

type clientConfig struct {
	baseURL      string
	basePath     string
	headers      map[string]string
	authProvider func() string
}

// WithBaseURL sets the base URL for API calls (e.g., "https://api.example.com")
func WithBaseURL(url string) ClientOption {
	return func(c *clientConfig) {
		c.baseURL = url
	}
}

// WithBasePath overrides the default API path prefix (e.g., "/api/v1/posts")
func WithBasePath(path string) ClientOption {
	return func(c *clientConfig) {
		c.basePath = path
	}
}

// WithHeader adds a header to all requests
func WithHeader(key, value string) ClientOption {
	return func(c *clientConfig) {
		if c.headers == nil {
			c.headers = make(map[string]string)
		}
		c.headers[key] = value
	}
}

// WithAuthProvider sets a function that provides the Authorization header value dynamically.
// The function is called on each request, allowing for token refresh scenarios.
// Example: WithAuthProvider(func() string { return "Bearer " + auth.GetToken() })
func WithAuthProvider(provider func() string) ClientOption {
	return func(c *clientConfig) {
		c.authProvider = provider
	}
}

func doRequest[T any](cfg *clientConfig, method, path string, body any) (T, error) {
	var result T

	url := cfg.baseURL + cfg.basePath + path

	var bodyStr string
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return result, fmt.Errorf("marshal request: %w", err)
		}
		bodyStr = string(data)
	}

	headers := make(map[string]string)
	for k, v := range cfg.headers {
		headers[k] = v
	}
	if cfg.authProvider != nil {
		if authValue := cfg.authProvider(); authValue != "" {
			headers["Authorization"] = authValue
		}
	}
	if body != nil {
		headers["Content-Type"] = "application/json"
	}

	resp, err := fetch.Fetch(url, &fetch.Options{
		Method:  method,
		Headers: headers,
		Body:    bodyStr,
	})
	if err != nil {
		return result, fmt.Errorf("fetch failed: %w", err)
	}

	if !resp.OK {
		return result, fmt.Errorf("unexpected status %d: %s", resp.Status, resp.StatusText)
	}

	// For DELETE or no-content responses
	if resp.Body == "" {
		return result, nil
	}

	if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
		return result, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}

func doRequestNoResponse(cfg *clientConfig, method, path string) error {
	url := cfg.baseURL + cfg.basePath + path

	headers := make(map[string]string)
	for k, v := range cfg.headers {
		headers[k] = v
	}
	if cfg.authProvider != nil {
		if authValue := cfg.authProvider(); authValue != "" {
			headers["Authorization"] = authValue
		}
	}

	resp, err := fetch.Fetch(url, &fetch.Options{
		Method:  method,
		Headers: headers,
	})
	if err != nil {
		return fmt.Errorf("fetch failed: %w", err)
	}

	if !resp.OK {
		return fmt.Errorf("unexpected status %d: %s", resp.Status, resp.StatusText)
	}

	return nil
}
`, nil
}

func generateClientCode(interfaces []InterfaceInfo) (string, error) {
	// Check if any method has path parameters (needs fmt import for Sprintf)
	needsFmt := false
	for _, iface := range interfaces {
		for _, method := range iface.Methods {
			if len(method.PathParams) > 0 {
				needsFmt = true
				break
			}
		}
		if needsFmt {
			break
		}
	}

	tmpl := `// Code generated by gux. DO NOT EDIT.
//go:build js && wasm

package api
{{if .NeedsFmt}}
import "fmt"
{{end}}
{{range $iface := .Interfaces}}
// {{$iface.ClientName}} is a client for {{$iface.Name}}
type {{$iface.ClientName}} struct {
	cfg *clientConfig
}

// New{{$iface.ClientName}} creates a new {{$iface.ClientName}}
func New{{$iface.ClientName}}(opts ...ClientOption) *{{$iface.ClientName}} {
	cfg := &clientConfig{
		baseURL:  "",
		basePath: "{{$iface.BasePath}}",
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return &{{$iface.ClientName}}{cfg: cfg}
}

{{range $method := $iface.Methods}}
// {{$method.Name}} {{if eq $method.HTTPMethod "GET"}}fetches{{else if eq $method.HTTPMethod "POST"}}creates{{else if eq $method.HTTPMethod "PUT"}}updates{{else if eq $method.HTTPMethod "DELETE"}}deletes{{else}}handles{{end}} data via {{$method.HTTPMethod}} {{$iface.BasePath}}{{$method.Path}}
{{- if $method.HasReturn}}
func (c *{{$iface.ClientName}}) {{$method.Name}}({{range $i, $p := $method.PathParams}}{{if $i}}, {{end}}{{$p.Name}} {{$p.Type}}{{end}}{{if and $method.PathParams $method.HasBody}}, {{end}}{{if $method.HasBody}}{{$method.BodyParam}} {{$method.BodyType}}{{end}}) ({{if $method.IsPointer}}*{{end}}{{if $method.IsSlice}}[]{{end}}{{$method.ReturnType | stripPrefix}}, error) {
	{{- if $method.IsPointer}}
	result, err := doRequest[{{$method.ReturnType}}](c.cfg, "{{$method.HTTPMethod}}", {{buildPath $method.Path $method.PathParams}}{{if $method.HasBody}}, {{$method.BodyParam}}{{else}}, nil{{end}})
	if err != nil {
		return nil, err
	}
	return &result, nil
	{{- else}}
	return doRequest[{{if $method.IsSlice}}[]{{end}}{{$method.ReturnType | stripPrefix}}](c.cfg, "{{$method.HTTPMethod}}", {{buildPath $method.Path $method.PathParams}}{{if $method.HasBody}}, {{$method.BodyParam}}{{else}}, nil{{end}})
	{{- end}}
}
{{- else}}
func (c *{{$iface.ClientName}}) {{$method.Name}}({{range $i, $p := $method.PathParams}}{{if $i}}, {{end}}{{$p.Name}} {{$p.Type}}{{end}}) error {
	return doRequestNoResponse(c.cfg, "{{$method.HTTPMethod}}", {{buildPath $method.Path $method.PathParams}})
}
{{- end}}
{{end}}
{{end}}`

	funcMap := template.FuncMap{
		"buildPath": func(path string, params []PathParam) string {
			if len(params) == 0 {
				return `"` + path + `"`
			}
			// Build a map of param name to type for lookup
			paramTypes := make(map[string]string)
			for _, p := range params {
				paramTypes[p.Name] = p.Type
			}
			// Replace each {param} with the appropriate format specifier
			re := regexp.MustCompile(`\{(\w+)\}`)
			result := re.ReplaceAllStringFunc(path, func(match string) string {
				paramName := match[1 : len(match)-1] // strip { and }
				if paramTypes[paramName] == "int" {
					return "%d"
				}
				return "%s"
			})
			// Build the parameter list
			var paramNames []string
			for _, p := range params {
				paramNames = append(paramNames, p.Name)
			}
			return `fmt.Sprintf("` + result + `", ` + strings.Join(paramNames, ", ") + `)`
		},
		"stripPrefix": func(s string) string {
			return strings.TrimPrefix(s, "[]")
		},
	}

	t := template.Must(template.New("client").Funcs(funcMap).Parse(tmpl))

	data := struct {
		Interfaces []InterfaceInfo
		NeedsFmt   bool
	}{
		Interfaces: interfaces,
		NeedsFmt:   needsFmt,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

func generateServerCode(interfaces []InterfaceInfo) (string, error) {
	tmpl := `// Code generated by gux. DO NOT EDIT.

package api

import (
	"encoding/json"
	"net/http"
{{- if .NeedsStrconv}}
	"strconv"
{{- end}}
{{- if .HasPathParams}}
	"strings"
{{- end}}

	gqapi "github.com/dougbarrett/gux/api"
)

{{range $iface := .Interfaces}}
// {{$iface.Name}}Handler wraps a {{$iface.Name}} implementation with HTTP handlers
type {{$iface.Name}}Handler struct {
	service    {{$iface.Name}}
	middleware []func(http.Handler) http.Handler
}

// New{{$iface.Name}}Handler creates a new HTTP handler for {{$iface.Name}}
func New{{$iface.Name}}Handler(service {{$iface.Name}}) *{{$iface.Name}}Handler {
	return &{{$iface.Name}}Handler{service: service}
}

// Use adds middleware to the handler chain
func (h *{{$iface.Name}}Handler) Use(mw ...func(http.Handler) http.Handler) {
	h.middleware = append(h.middleware, mw...)
}

// wrap applies middleware chain to a handler
func (h *{{$iface.Name}}Handler) wrap(handler http.HandlerFunc) http.Handler {
	var result http.Handler = handler
	for i := len(h.middleware) - 1; i >= 0; i-- {
		result = h.middleware[i](result)
	}
	return result
}

// RegisterRoutes registers all routes for {{$iface.Name}}
func (h *{{$iface.Name}}Handler) RegisterRoutes(mux *http.ServeMux) {
{{- range $method := $iface.Methods}}
	mux.Handle("{{$method.HTTPMethod}} {{$iface.BasePath}}{{$method.Path}}", h.wrap(h.handle{{$method.Name}}))
{{- end}}
}

{{range $method := $iface.Methods}}
func (h *{{$iface.Name}}Handler) handle{{$method.Name}}(w http.ResponseWriter, r *http.Request) {
{{- if $method.PathParams}}
	// Extract path parameters
	path := strings.TrimPrefix(r.URL.Path, "{{$iface.BasePath}}")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	_ = parts // avoid unused variable if no params extracted
{{- range $p := $method.PathParams}}
{{- if $p.IsInt}}
	{{$p.Name}}, err := strconv.Atoi(parts[{{pathParamIndex $method.Path $p.Name}}])
	if err != nil {
		gqapi.WriteError(w, gqapi.BadRequest("invalid {{$p.Name}}: must be an integer"))
		return
	}
{{- else}}
	{{$p.Name}} := parts[{{pathParamIndex $method.Path $p.Name}}]
{{- end}}
{{- end}}
{{- end}}
{{- if $method.HasBody}}
	var req {{$method.BodyType}}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		gqapi.WriteError(w, gqapi.BadRequest("invalid request body"))
		return
	}
{{- end}}

	{{if $method.HasReturn}}result, {{end}}err {{if or $method.HasReturn (not (hasIntPathParam $method.PathParams))}}:{{end}}= h.service.{{$method.Name}}(r.Context(){{range $method.PathParams}}, {{.Name}}{{end}}{{if $method.HasBody}}, req{{end}})
	if err != nil {
		gqapi.WriteError(w, err)
		return
	}

{{- if $method.HasReturn}}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
{{- else}}
	w.WriteHeader(http.StatusNoContent)
{{- end}}
}
{{end}}
{{end}}
`

	// Check if any interface has path parameters (needs strings import)
	// and if any have int path parameters (needs strconv import)
	needsStrconv := false
	hasPathParams := false
	for _, iface := range interfaces {
		for _, method := range iface.Methods {
			if len(method.PathParams) > 0 {
				hasPathParams = true
			}
			for _, p := range method.PathParams {
				if p.IsInt {
					needsStrconv = true
				}
			}
		}
	}

	funcMap := template.FuncMap{
		"methodName": func(method string) string {
			return strings.ToUpper(method[:1]) + strings.ToLower(method[1:])
		},
		"pathParamIndex": func(path, param string) int {
			// Find the index of the parameter in the path parts
			// e.g., "/{userId}/posts/{postId}" -> userId is at index 0, postId is at index 2
			parts := strings.Split(strings.Trim(path, "/"), "/")
			for i, part := range parts {
				if part == "{"+param+"}" {
					return i
				}
			}
			return 0
		},
		"hasIntPathParam": func(params []PathParam) bool {
			for _, p := range params {
				if p.IsInt {
					return true
				}
			}
			return false
		},
	}

	t := template.Must(template.New("server").Funcs(funcMap).Parse(tmpl))

	data := struct {
		Interfaces    []InterfaceInfo
		NeedsStrconv  bool
		HasPathParams bool
	}{
		Interfaces:    interfaces,
		NeedsStrconv:  needsStrconv,
		HasPathParams: hasPathParams,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

