package ui

import (
	"strings"
	"testing"
)

func TestBarChart_BasicRender(t *testing.T) {
	node := BarChart(BarChartProps{
		Title:  "Revenue",
		Labels: []string{"Q1", "Q2", "Q3"},
		Series: []ChartSeries{
			{Name: "Sales", Values: []float64{10, 20, 30}, Color: "blue"},
		},
	})
	html := renderHTML(node)

	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG element, got: %s", truncate(html))
	}
	if !strings.Contains(html, `role="img"`) {
		t.Errorf("expected role=img, got: %s", truncate(html))
	}
	if !strings.Contains(html, "<title>Revenue</title>") {
		t.Errorf("expected title element, got: %s", truncate(html))
	}
	if !strings.Contains(html, `aria-label="Revenue"`) {
		t.Errorf("expected aria-label, got: %s", truncate(html))
	}
}

func TestBarChart_HorizontalRender(t *testing.T) {
	node := BarChart(BarChartProps{
		Title:      "Leads by Source",
		Labels:     []string{"Website", "Referral"},
		Horizontal: true,
		Series: []ChartSeries{
			{Name: "Leads", Values: []float64{45, 30}, Color: "green"},
		},
	})
	html := renderHTML(node)

	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG element for horizontal bar, got: %s", truncate(html))
	}
}

func TestBarChart_MultipleSeries(t *testing.T) {
	node := BarChart(BarChartProps{
		Title:  "Comparison",
		Labels: []string{"A", "B"},
		Series: []ChartSeries{
			{Name: "2023", Values: []float64{10, 20}, Color: "blue"},
			{Name: "2024", Values: []float64{15, 25}, Color: "red"},
		},
	})
	html := renderHTML(node)

	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG for multi-series bar chart, got: %s", truncate(html))
	}
}

func TestBarChart_EmptySeries(t *testing.T) {
	node := BarChart(BarChartProps{
		Title: "Empty",
	})
	html := renderHTML(node)

	if strings.Contains(html, "<svg") {
		t.Errorf("empty series should not produce SVG, got: %s", truncate(html))
	}
}

func TestBarChart_CustomClass(t *testing.T) {
	node := BarChart(BarChartProps{
		Title:  "Test",
		Labels: []string{"A"},
		Series: []ChartSeries{
			{Name: "S", Values: []float64{10}},
		},
		Class: "my-custom-class",
	})
	html := renderHTML(node)

	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class, got: %s", truncate(html))
	}
}

func TestBarChart_DefaultColors(t *testing.T) {
	// Series without explicit Color should use default palette
	node := BarChart(BarChartProps{
		Title:  "Defaults",
		Labels: []string{"A"},
		Series: []ChartSeries{
			{Name: "S1", Values: []float64{10}},
			{Name: "S2", Values: []float64{20}},
		},
	})
	html := renderHTML(node)

	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG with default colors, got: %s", truncate(html))
	}
}

func TestBarChart_ShowValues(t *testing.T) {
	node := BarChart(BarChartProps{
		Title:      "Values",
		Labels:     []string{"A", "B"},
		ShowValues: true,
		Series: []ChartSeries{
			{Name: "S", Values: []float64{10, 20}},
		},
	})
	html := renderHTML(node)

	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG with values shown, got: %s", truncate(html))
	}
}

func TestBarChart_CustomDimensions(t *testing.T) {
	node := BarChart(BarChartProps{
		Title:  "Custom Size",
		Labels: []string{"A"},
		Width:  800,
		Height: 300,
		Series: []ChartSeries{
			{Name: "S", Values: []float64{10}},
		},
	})
	html := renderHTML(node)

	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG with custom dimensions, got: %s", truncate(html))
	}
}

func TestLineChart_BasicRender(t *testing.T) {
	node := LineChart(LineChartProps{
		Title:  "Trend",
		Labels: []string{"Jan", "Feb", "Mar"},
		Series: []ChartSeries{
			{Name: "Revenue", Values: []float64{1200, 1500, 1800}, Color: "blue"},
		},
	})
	html := renderHTML(node)

	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG element, got: %s", truncate(html))
	}
	if !strings.Contains(html, `role="img"`) {
		t.Errorf("expected role=img, got: %s", truncate(html))
	}
	if !strings.Contains(html, "<title>Trend</title>") {
		t.Errorf("expected title element, got: %s", truncate(html))
	}
}

func TestLineChart_MultipleSeries(t *testing.T) {
	node := LineChart(LineChartProps{
		Title:  "Revenue Breakdown",
		Labels: []string{"Jan", "Feb", "Mar"},
		Series: []ChartSeries{
			{Name: "Setup", Values: []float64{5000, 7200, 4800}, Color: "blue"},
			{Name: "Monthly", Values: []float64{1200, 1500, 1800}, Color: "green"},
		},
	})
	html := renderHTML(node)

	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG for multi-series line chart, got: %s", truncate(html))
	}
}

func TestLineChart_EmptySeries(t *testing.T) {
	node := LineChart(LineChartProps{
		Title: "Empty",
	})
	html := renderHTML(node)

	if strings.Contains(html, "<svg") {
		t.Errorf("empty series should not produce SVG, got: %s", truncate(html))
	}
}

func TestLineChart_ShowDots(t *testing.T) {
	node := LineChart(LineChartProps{
		Title:    "Dots",
		Labels:   []string{"A", "B", "C"},
		ShowDots: true,
		Series: []ChartSeries{
			{Name: "S", Values: []float64{1, 2, 3}},
		},
	})
	html := renderHTML(node)

	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG with dots, got: %s", truncate(html))
	}
}

func TestLineChart_CustomClass(t *testing.T) {
	node := LineChart(LineChartProps{
		Title:  "Test",
		Labels: []string{"A"},
		Series: []ChartSeries{
			{Name: "S", Values: []float64{10}},
		},
		Class: "mt-8",
	})
	html := renderHTML(node)

	if !strings.Contains(html, "mt-8") {
		t.Errorf("expected custom class, got: %s", truncate(html))
	}
}

func TestResolveChartColor_TailwindName(t *testing.T) {
	c := resolveChartColor("blue")
	if c.R == 0 && c.G == 0 && c.B == 0 {
		t.Error("expected non-zero color for 'blue'")
	}
}

func TestResolveChartColor_Hex(t *testing.T) {
	c := resolveChartColor("#ff0000")
	if c.R != 255 {
		t.Errorf("expected R=255 for #ff0000, got R=%d", c.R)
	}
}

func TestResolveChartColor_Empty(t *testing.T) {
	c := resolveChartColor("")
	if c.R != 0 || c.G != 0 || c.B != 0 {
		t.Errorf("expected zero color for empty string, got R=%d G=%d B=%d", c.R, c.G, c.B)
	}
}

func TestWrapperHasRoleFigure(t *testing.T) {
	node := BarChart(BarChartProps{
		Title:  "Test",
		Labels: []string{"A"},
		Series: []ChartSeries{
			{Name: "S", Values: []float64{10}},
		},
	})
	html := renderHTML(node)

	if !strings.Contains(html, `role="figure"`) {
		t.Errorf("expected role=figure on wrapper div, got: %s", truncate(html))
	}
}

func TestChartAccessibility_Desc(t *testing.T) {
	node := BarChart(BarChartProps{
		Title:  "Test Chart",
		Labels: []string{"A"},
		Series: []ChartSeries{
			{Name: "S", Values: []float64{10}},
		},
	})
	html := renderHTML(node)

	if !strings.Contains(html, "<desc>Chart: Test Chart</desc>") {
		t.Errorf("expected desc element, got: %s", truncate(html))
	}
}

// truncate returns the first 500 chars of a string for error messages.
func truncate(s string) string {
	if len(s) > 500 {
		return s[:500] + "..."
	}
	return s
}
