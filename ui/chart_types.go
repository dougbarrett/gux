package ui

import (
	"fmt"
	"html"
	"strings"

	"github.com/dougbarrett/gux/core"
)

// ChartSeries represents a single data series in a chart.
type ChartSeries struct {
	Name   string    // Series name (shown in legend)
	Values []float64 // Data values
	Color  string    // Tailwind color name ("blue") or hex ("#3b82f6")
}

// BarChartProps configures a BarChart component.
type BarChartProps struct {
	Title      string        // Chart title (also used for accessibility)
	Labels     []string      // X-axis labels (category names)
	Series     []ChartSeries // Data series
	Horizontal bool          // Render as horizontal bar chart
	ShowValues bool          // Display value labels on bars
	Width      int           // SVG width in px (default: 600)
	Height     int           // SVG height in px (default: 400)
	ShowLegend bool          // Show legend (default: true when multiple series)
	Class      string        // Additional CSS classes on wrapper div
}

// LineChartProps configures a LineChart component.
type LineChartProps struct {
	Title    string        // Chart title (also used for accessibility)
	Labels   []string      // X-axis labels
	Series   []ChartSeries // Data series
	Width    int           // SVG width in px (default: 600)
	Height   int           // SVG height in px (default: 400)
	ShowGrid bool          // Show horizontal grid lines
	ShowDots bool          // Show data point markers
	Class    string        // Additional CSS classes on wrapper div
}

// AreaChartProps configures an AreaChart component.
type AreaChartProps struct {
	Title      string        // Chart title (also used for accessibility)
	Labels     []string      // X-axis labels
	Series     []ChartSeries // Data series
	Stacked    bool          // Stack series (cumulative sum, each layer fills to the one below)
	Width      int           // SVG width in px (default: 600)
	Height     int           // SVG height in px (default: 400)
	ShowDots   bool          // Show data point markers
	Opacity    uint8         // Fill opacity 0-255 (default: 128)
	ShowLegend bool          // Show legend (default: true when multiple series)
	Class      string        // Additional CSS classes on wrapper div
}

// resolveChartDimensions returns width and height with defaults applied.
func resolveChartDimensions(width, height int) (int, int) {
	if width <= 0 {
		width = 600
	}
	if height <= 0 {
		height = 400
	}
	return width, height
}

// wrapChartSVG adds accessibility attributes to the SVG and wraps it in a responsive div.
func wrapChartSVG(svg string, title string, class string) core.Node {
	svg = injectChartAccessibility(svg, title)

	wrapperClass := MergeClasses("w-full overflow-x-auto", class)

	return core.Div(core.Attrs{
		Class: wrapperClass,
		Extra: map[string]string{
			"role":       "figure",
			"aria-label": title,
		},
	}, core.RawHTML(svg))
}

// injectChartAccessibility adds role, aria-label, <title>, and <desc> to the SVG element.
func injectChartAccessibility(svg string, title string) string {
	if title == "" {
		return svg
	}

	escapedTitle := html.EscapeString(title)

	// Add role="img" and aria-label to the <svg> tag
	svg = strings.Replace(svg,
		"<svg ",
		`<svg role="img" aria-label="`+escapedTitle+`" `,
		1)

	// Inject <title> and <desc> as first children inside <svg>
	titleTag := "<title>" + escapedTitle + "</title>"
	descTag := "<desc>Chart: " + escapedTitle + "</desc>"

	idx := strings.Index(svg, ">")
	if idx >= 0 {
		svg = svg[:idx+1] + titleTag + descTag + svg[idx+1:]
	}

	return svg
}

// chartError renders an error message when chart generation fails.
func chartError(title string, err error) core.Node {
	msg := fmt.Sprintf("Chart error: %v", err)
	if title != "" {
		msg = fmt.Sprintf("Chart %q error: %v", title, err)
	}
	println("[gux] " + msg)
	return core.Div(core.Class("p-4 text-red-500 dark:text-red-400 text-sm border border-red-200 dark:border-red-800 rounded"),
		core.Text(msg),
	)
}
