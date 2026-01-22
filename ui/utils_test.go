package ui

import "testing"

func TestMergeClasses(t *testing.T) {
	tests := []struct {
		name    string
		classes []string
		want    string
	}{
		{
			name:    "single class",
			classes: []string{"foo"},
			want:    "foo",
		},
		{
			name:    "multiple classes",
			classes: []string{"foo", "bar"},
			want:    "foo bar",
		},
		{
			name:    "empty strings filtered",
			classes: []string{"foo", "", "bar"},
			want:    "foo bar",
		},
		{
			name:    "whitespace trimmed",
			classes: []string{"  foo  ", "bar"},
			want:    "foo bar",
		},
		{
			name:    "all empty returns empty",
			classes: []string{"", ""},
			want:    "",
		},
		{
			name:    "no args returns empty",
			classes: []string{},
			want:    "",
		},
		{
			name:    "whitespace only strings filtered",
			classes: []string{"foo", "   ", "bar"},
			want:    "foo bar",
		},
		{
			name:    "multiple spaces in middle",
			classes: []string{"foo", "", "", "bar", "baz"},
			want:    "foo bar baz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeClasses(tt.classes...)
			if got != tt.want {
				t.Errorf("MergeClasses(%v) = %q, want %q", tt.classes, got, tt.want)
			}
		})
	}
}

func TestConditionalClass(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		class     string
		want      string
	}{
		{
			name:      "true returns class",
			condition: true,
			class:     "hidden",
			want:      "hidden",
		},
		{
			name:      "false returns empty",
			condition: false,
			class:     "hidden",
			want:      "",
		},
		{
			name:      "true with complex class",
			condition: true,
			class:     "bg-blue-500 hover:bg-blue-600",
			want:      "bg-blue-500 hover:bg-blue-600",
		},
		{
			name:      "false with complex class",
			condition: false,
			class:     "bg-blue-500 hover:bg-blue-600",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConditionalClass(tt.condition, tt.class)
			if got != tt.want {
				t.Errorf("ConditionalClass(%v, %q) = %q, want %q", tt.condition, tt.class, got, tt.want)
			}
		})
	}
}
