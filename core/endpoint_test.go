package core

import (
	"testing"
)

func TestConvertColonParams(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
	}{
		{
			name:  "no params",
			input: "GET /api/users",
			want:  "GET /api/users",
		},
		{
			name:  "single param",
			input: "GET /api/users/:id",
			want:  "GET /api/users/{id}",
		},
		{
			name:  "multiple params",
			input: "GET /api/users/:id/posts/:postId",
			want:  "GET /api/users/{id}/posts/{postId}",
		},
		{
			name:  "slug param",
			input: "GET /api/promotions/:slug",
			want:  "GET /api/promotions/{slug}",
		},
		{
			name:  "POST with param",
			input: "POST /api/users/:id",
			want:  "POST /api/users/{id}",
		},
		{
			name:  "DELETE with param",
			input: "DELETE /api/users/:id",
			want:  "DELETE /api/users/{id}",
		},
		{
			name:  "no method prefix",
			input: "/api/users/:id",
			want:  "/api/users/{id}",
		},
		{
			name:  "already bracket syntax",
			input: "GET /api/users/{id}",
			want:  "GET /api/users/{id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertColonParams(tt.input)
			if got != tt.want {
				t.Errorf("convertColonParams(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractParamNames(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			name:    "no params",
			pattern: "/api/users",
			want:    nil,
		},
		{
			name:    "single param",
			pattern: "/api/users/:id",
			want:    []string{"id"},
		},
		{
			name:    "multiple params",
			pattern: "/api/users/:id/posts/:postId",
			want:    []string{"id", "postId"},
		},
		{
			name:    "slug param",
			pattern: "/api/promotions/:slug",
			want:    []string{"slug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractParamNames(tt.pattern)
			if len(got) != len(tt.want) {
				t.Fatalf("extractParamNames(%q) returned %d names, want %d", tt.pattern, len(got), len(tt.want))
			}
			for i, name := range got {
				if name != tt.want[i] {
					t.Errorf("extractParamNames(%q)[%d] = %q, want %q", tt.pattern, i, name, tt.want[i])
				}
			}
		})
	}
}

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    map[string]string
	}{
		{
			name:    "no params",
			pattern: "/api/users",
			path:    "/api/users",
			want:    map[string]string{},
		},
		{
			name:    "single numeric param",
			pattern: "/api/users/:id",
			path:    "/api/users/123",
			want:    map[string]string{"id": "123"},
		},
		{
			name:    "single string param",
			pattern: "/api/promotions/:slug",
			path:    "/api/promotions/my-promo",
			want:    map[string]string{"slug": "my-promo"},
		},
		{
			name:    "multiple params",
			pattern: "/api/users/:id/posts/:postId",
			path:    "/api/users/42/posts/99",
			want:    map[string]string{"id": "42", "postId": "99"},
		},
		{
			name:    "mismatched lengths returns empty",
			pattern: "/api/users/:id",
			path:    "/api/users",
			want:    map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPathParams(tt.pattern, tt.path)
			if len(got) != len(tt.want) {
				t.Fatalf("extractPathParams(%q, %q) returned %d params, want %d", tt.pattern, tt.path, len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("extractPathParams(%q, %q)[%q] = %q, want %q", tt.pattern, tt.path, k, got[k], v)
				}
			}
		})
	}
}
