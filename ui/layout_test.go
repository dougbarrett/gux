package ui

import (
	"strings"
	"testing"

	"github.com/dougbarrett/gux/core"
)

func TestContainer_DefaultMaxWidth(t *testing.T) {
	html := renderHTML(Container(ContainerProps{}))

	if !strings.Contains(html, "max-w-7xl") {
		t.Error("Container should default to max-w-7xl")
	}
	if !strings.Contains(html, "mx-auto") {
		t.Error("Container should have mx-auto for centering")
	}
	if !strings.Contains(html, "px-4") {
		t.Error("Container should have responsive padding px-4")
	}
}

func TestContainer_CustomMaxWidth(t *testing.T) {
	tests := []struct {
		maxWidth string
		expected string
	}{
		{"sm", "max-w-sm"},
		{"md", "max-w-md"},
		{"lg", "max-w-lg"},
		{"xl", "max-w-xl"},
		{"2xl", "max-w-2xl"},
		{"7xl", "max-w-7xl"},
	}

	for _, tt := range tests {
		t.Run(tt.maxWidth, func(t *testing.T) {
			html := renderHTML(Container(ContainerProps{MaxWidth: tt.maxWidth}))
			if !strings.Contains(html, tt.expected) {
				t.Errorf("Container with MaxWidth=%q should have %q class", tt.maxWidth, tt.expected)
			}
		})
	}
}

func TestContainer_CustomClass(t *testing.T) {
	html := renderHTML(Container(ContainerProps{Class: "py-8"}))

	if !strings.Contains(html, "py-8") {
		t.Error("Container should include custom class")
	}
	if !strings.Contains(html, "mx-auto") {
		t.Error("Container should retain default mx-auto with custom class")
	}
}

func TestVStack_FlexCol(t *testing.T) {
	html := renderHTML(VStack(StackProps{}))

	if !strings.Contains(html, "flex") {
		t.Error("VStack should have flex class")
	}
	if !strings.Contains(html, "flex-col") {
		t.Error("VStack should have flex-col class")
	}
	if !strings.Contains(html, "gap-4") {
		t.Error("VStack should default to gap-4")
	}
}

func TestHStack_FlexRow(t *testing.T) {
	html := renderHTML(HStack(StackProps{}))

	if !strings.Contains(html, "flex") {
		t.Error("HStack should have flex class")
	}
	if !strings.Contains(html, "flex-row") {
		t.Error("HStack should have flex-row class")
	}
	if !strings.Contains(html, "gap-4") {
		t.Error("HStack should default to gap-4")
	}
}

func TestStack_Gap(t *testing.T) {
	tests := []struct {
		gap      string
		expected string
	}{
		{"0", "gap-0"},
		{"1", "gap-1"},
		{"2", "gap-2"},
		{"4", "gap-4"},
		{"6", "gap-6"},
		{"8", "gap-8"},
	}

	for _, tt := range tests {
		t.Run("VStack-"+tt.gap, func(t *testing.T) {
			html := renderHTML(VStack(StackProps{Gap: tt.gap}))
			if !strings.Contains(html, tt.expected) {
				t.Errorf("VStack with Gap=%q should have %q class", tt.gap, tt.expected)
			}
		})
		t.Run("HStack-"+tt.gap, func(t *testing.T) {
			html := renderHTML(HStack(StackProps{Gap: tt.gap}))
			if !strings.Contains(html, tt.expected) {
				t.Errorf("HStack with Gap=%q should have %q class", tt.gap, tt.expected)
			}
		})
	}
}

func TestStack_Alignment(t *testing.T) {
	tests := []struct {
		name     string
		align    string
		justify  string
		expected []string
	}{
		{"align-start", "start", "", []string{"items-start"}},
		{"align-center", "center", "", []string{"items-center"}},
		{"align-end", "end", "", []string{"items-end"}},
		{"align-stretch", "stretch", "", []string{"items-stretch"}},
		{"justify-start", "", "start", []string{"justify-start"}},
		{"justify-center", "", "center", []string{"justify-center"}},
		{"justify-end", "", "end", []string{"justify-end"}},
		{"justify-between", "", "between", []string{"justify-between"}},
		{"justify-around", "", "around", []string{"justify-around"}},
		{"combined", "center", "between", []string{"items-center", "justify-between"}},
	}

	for _, tt := range tests {
		t.Run("VStack-"+tt.name, func(t *testing.T) {
			html := renderHTML(VStack(StackProps{Align: tt.align, Justify: tt.justify}))
			for _, exp := range tt.expected {
				if !strings.Contains(html, exp) {
					t.Errorf("VStack should have %q class", exp)
				}
			}
		})
		t.Run("HStack-"+tt.name, func(t *testing.T) {
			html := renderHTML(HStack(StackProps{Align: tt.align, Justify: tt.justify}))
			for _, exp := range tt.expected {
				if !strings.Contains(html, exp) {
					t.Errorf("HStack should have %q class", exp)
				}
			}
		})
	}
}

func TestStack_Children(t *testing.T) {
	html := renderHTML(VStack(StackProps{
		Children: []core.Node{
			core.Text("Item 1"),
			core.Text("Item 2"),
		},
	}))

	if !strings.Contains(html, "Item 1") {
		t.Error("VStack should render first child")
	}
	if !strings.Contains(html, "Item 2") {
		t.Error("VStack should render second child")
	}
}

func TestGrid_Columns(t *testing.T) {
	tests := []struct {
		cols     string
		expected string
	}{
		{"1", "grid-cols-1"},
		{"2", "grid-cols-2"},
		{"3", "grid-cols-3"},
		{"4", "grid-cols-4"},
		{"6", "grid-cols-6"},
		{"12", "grid-cols-12"},
	}

	for _, tt := range tests {
		t.Run("cols-"+tt.cols, func(t *testing.T) {
			html := renderHTML(Grid(GridProps{Cols: tt.cols}))
			if !strings.Contains(html, tt.expected) {
				t.Errorf("Grid with Cols=%q should have %q class", tt.cols, tt.expected)
			}
		})
	}
}

func TestGrid_DefaultColumns(t *testing.T) {
	html := renderHTML(Grid(GridProps{}))

	if !strings.Contains(html, "grid-cols-3") {
		t.Error("Grid should default to grid-cols-3")
	}
	if !strings.Contains(html, "gap-4") {
		t.Error("Grid should default to gap-4")
	}
	if !strings.Contains(html, "grid") {
		t.Error("Grid should have grid class")
	}
}

func TestGrid_Gap(t *testing.T) {
	html := renderHTML(Grid(GridProps{Gap: "8"}))

	if !strings.Contains(html, "gap-8") {
		t.Error("Grid should use custom gap-8")
	}
}

func TestGrid_CustomClass(t *testing.T) {
	html := renderHTML(Grid(GridProps{Class: "my-grid"}))

	if !strings.Contains(html, "my-grid") {
		t.Error("Grid should include custom class")
	}
}

func TestDivider_Horizontal(t *testing.T) {
	html := renderHTML(Divider(DividerProps{}))

	if !strings.Contains(html, "h-px") {
		t.Error("Horizontal Divider should have h-px class")
	}
	if !strings.Contains(html, "w-full") {
		t.Error("Horizontal Divider should have w-full class")
	}
	if !strings.Contains(html, "bg-gray-200") {
		t.Error("Divider should have bg-gray-200 class")
	}
}

func TestDivider_Vertical(t *testing.T) {
	html := renderHTML(Divider(DividerProps{Orientation: "vertical"}))

	if !strings.Contains(html, "w-px") {
		t.Error("Vertical Divider should have w-px class")
	}
	if !strings.Contains(html, "h-full") {
		t.Error("Vertical Divider should have h-full class")
	}
}

func TestDivider_CustomClass(t *testing.T) {
	html := renderHTML(Divider(DividerProps{Class: "my-4"}))

	if !strings.Contains(html, "my-4") {
		t.Error("Divider should include custom class")
	}
	if !strings.Contains(html, "h-px") {
		t.Error("Divider should retain default h-px with custom class")
	}
}
