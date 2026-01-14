package main

import (
	"bytes"
	"flag"
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

var (
	sourceFile = flag.String("source", "", "Source file containing interface")
	outputFile = flag.String("output", "", "Output file for generated client")
)

type InterfaceInfo struct {
	Name       string
	ClientName string
	BasePath   string
	Methods    []MethodInfo
}

type MethodInfo struct {
	Name       string
	HTTPMethod string
	Path       string
	PathParams []string
	HasBody    bool
	BodyParam  string
	BodyType   string
	ReturnType string
	IsPointer  bool
	IsSlice    bool
	HasReturn  bool
}

func main() {
	flag.Parse()

	if *sourceFile == "" || *outputFile == "" {
		fmt.Println("Usage: apigen -source=<file.go> -output=<output.go>")
		os.Exit(1)
	}

	// Get the directory of the source file
	dir := filepath.Dir(*sourceFile)
	if dir == "" {
		dir = "."
	}

	// Parse the source file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, *sourceFile, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		os.Exit(1)
	}

	// Find interfaces with @client annotation
	interfaces := findInterfaces(node)
	if len(interfaces) == 0 {
		fmt.Println("No interfaces with @client annotation found")
		os.Exit(1)
	}

	// Generate client code
	clientCode := generateClient(interfaces)
	clientPath := filepath.Join(dir, *outputFile)
	if err := os.WriteFile(clientPath, []byte(clientCode), 0644); err != nil {
		fmt.Printf("Error writing client output: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated client: %s\n", clientPath)

	// Generate server code
	serverCode := generateServer(interfaces)
	serverOutput := strings.TrimSuffix(*outputFile, "_gen.go") + "_server_gen.go"
	serverOutput = strings.TrimSuffix(serverOutput, ".go") + ".go"
	if strings.HasSuffix(*outputFile, "_client_gen.go") {
		serverOutput = strings.Replace(*outputFile, "_client_gen.go", "_server_gen.go", 1)
	} else {
		serverOutput = strings.TrimSuffix(*outputFile, ".go") + "_server.go"
	}
	serverPath := filepath.Join(dir, serverOutput)
	if err := os.WriteFile(serverPath, []byte(serverCode), 0644); err != nil {
		fmt.Printf("Error writing server output: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated server: %s\n", serverPath)
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

				// Extract path parameters
				pathParamRegex := regexp.MustCompile(`\{(\w+)\}`)
				matches := pathParamRegex.FindAllStringSubmatch(methodInfo.Path, -1)
				for _, match := range matches {
					methodInfo.PathParams = append(methodInfo.PathParams, match[1])
				}

				// Parse function parameters (skip ctx, identify body param)
				if funcType.Params != nil {
					for i, param := range funcType.Params.List {
						if i == 0 {
							continue // Skip context
						}
						if len(param.Names) == 0 {
							continue
						}

						paramName := param.Names[0].Name
						isPathParam := false
						for _, pp := range methodInfo.PathParams {
							if pp == paramName {
								isPathParam = true
								break
							}
						}

						if !isPathParam {
							methodInfo.HasBody = true
							methodInfo.BodyParam = paramName
							methodInfo.BodyType = exprToString(param.Type)
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

func generateClient(interfaces []InterfaceInfo) string {
	tmpl := `// Code generated by apigen. DO NOT EDIT.
//go:build js && wasm

package api

import (
	"encoding/json"
	"fmt"

	"gux/fetch"
)

// ClientOption configures a client
type ClientOption func(*clientConfig)

type clientConfig struct {
	baseURL  string
	basePath string
	headers  map[string]string
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

{{range $iface := .}}
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
func (c *{{$iface.ClientName}}) {{$method.Name}}({{range $method.PathParams}}{{.}} int, {{end}}{{if $method.HasBody}}{{$method.BodyParam}} {{$method.BodyType}}{{end}}) ({{if $method.IsPointer}}*{{end}}{{if $method.IsSlice}}[]{{end}}{{$method.ReturnType | stripPrefix}}, error) {
	{{- if $method.IsPointer}}
	result, err := doRequest[{{$method.ReturnType}}](c.cfg, "{{$method.HTTPMethod}}", {{buildPath $method.Path}}{{if $method.HasBody}}, {{$method.BodyParam}}{{else}}, nil{{end}})
	if err != nil {
		return nil, err
	}
	return &result, nil
	{{- else}}
	return doRequest[{{if $method.IsSlice}}[]{{end}}{{$method.ReturnType | stripPrefix}}](c.cfg, "{{$method.HTTPMethod}}", {{buildPath $method.Path}}{{if $method.HasBody}}, {{$method.BodyParam}}{{else}}, nil{{end}})
	{{- end}}
}
{{- else}}
func (c *{{$iface.ClientName}}) {{$method.Name}}({{range $method.PathParams}}{{.}} int{{end}}) error {
	return doRequestNoResponse(c.cfg, "{{$method.HTTPMethod}}", {{buildPath $method.Path}})
}
{{- end}}
{{end}}
{{end}}
`

	funcMap := template.FuncMap{
		"buildPath": func(path string) string {
			re := regexp.MustCompile(`\{(\w+)\}`)
			params := extractParams(path)
			if len(params) == 0 {
				return `"` + path + `"`
			}
			result := re.ReplaceAllString(path, "%d")
			return `fmt.Sprintf("` + result + `", ` + strings.Join(params, ", ") + `)`
		},
		"stripPrefix": func(s string) string {
			return strings.TrimPrefix(s, "[]")
		},
	}

	t := template.Must(template.New("client").Funcs(funcMap).Parse(tmpl))

	var buf bytes.Buffer
	if err := t.Execute(&buf, interfaces); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		os.Exit(1)
	}

	return buf.String()
}

func generateServer(interfaces []InterfaceInfo) string {
	tmpl := `// Code generated by apigen. DO NOT EDIT.

package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	gqapi "gux/api"
)

{{range $iface := .}}
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
	mux.Handle("{{$iface.BasePath}}", h.wrap(h.handleRoot))
	mux.Handle("{{$iface.BasePath}}/", h.wrap(h.handleWithID))
}

func (h *{{$iface.Name}}Handler) handleRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
{{- range $method := $iface.Methods}}
{{- if eq $method.Path "/"}}
	case http.Method{{$method.HTTPMethod | methodName}}:
		h.handle{{$method.Name}}(w, r)
{{- end}}
{{- end}}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *{{$iface.Name}}Handler) handleWithID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "{{$iface.BasePath}}/")

	// Handle trailing slash - delegate to root handler
	if path == "" {
		h.handleRoot(w, r)
		return
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		gqapi.WriteError(w, gqapi.BadRequest("invalid ID"))
		return
	}

	switch r.Method {
{{- range $method := $iface.Methods}}
{{- if ne $method.Path "/"}}
	case http.Method{{$method.HTTPMethod | methodName}}:
		h.handle{{$method.Name}}(w, r, id)
{{- end}}
{{- end}}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

{{range $method := $iface.Methods}}
func (h *{{$iface.Name}}Handler) handle{{$method.Name}}(w http.ResponseWriter, r *http.Request{{if $method.PathParams}}, {{range $i, $p := $method.PathParams}}{{if $i}}, {{end}}{{$p}} int{{end}}{{end}}) {
{{- if $method.HasBody}}
	var req {{$method.BodyType}}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		gqapi.WriteError(w, gqapi.BadRequest("invalid request body"))
		return
	}
{{- end}}

	{{if $method.HasReturn}}result, {{end}}err := h.service.{{$method.Name}}(r.Context(){{range $method.PathParams}}, {{.}}{{end}}{{if $method.HasBody}}, req{{end}})
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

	funcMap := template.FuncMap{
		"methodName": func(method string) string {
			// Convert GET -> Get, POST -> Post, etc.
			return strings.ToUpper(method[:1]) + strings.ToLower(method[1:])
		},
	}

	t := template.Must(template.New("server").Funcs(funcMap).Parse(tmpl))

	var buf bytes.Buffer
	if err := t.Execute(&buf, interfaces); err != nil {
		fmt.Printf("Error executing server template: %v\n", err)
		os.Exit(1)
	}

	return buf.String()
}

func extractParams(path string) []string {
	re := regexp.MustCompile(`\{(\w+)\}`)
	matches := re.FindAllStringSubmatch(path, -1)
	var params []string
	for _, m := range matches {
		params = append(params, m[1])
	}
	return params
}
