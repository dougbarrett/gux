package ui

import (
	"fmt"
	"html"
	"strings"

	"github.com/dougbarrett/gux/core"
	"github.com/go-analyze/charts"
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

// BarChart renders a bar chart as inline SVG using the go-analyze/charts library.
//
// Example:
//
//	BarChart(BarChartProps{
//	    Title:  "Revenue by Quarter",
//	    Labels: []string{"Q1", "Q2", "Q3", "Q4"},
//	    Series: []ChartSeries{
//	        {Name: "2024", Values: []float64{100, 200, 150, 300}, Color: "blue"},
//	    },
//	})
func BarChart(props BarChartProps) core.Node {
	if len(props.Series) == 0 {
		return core.Frag()
	}

	width, height := resolveChartDimensions(props.Width, props.Height)

	// Build values as [][]float64
	values := make([][]float64, len(props.Series))
	seriesNames := make([]string, len(props.Series))
	for i, s := range props.Series {
		values[i] = s.Values
		seriesNames[i] = s.Name
	}

	// Build options
	opts := []charts.OptionFunc{
		charts.SVGOutputOptionFunc(),
		charts.DimensionsOptionFunc(width, height),
	}

	if props.Title != "" {
		opts = append(opts, charts.TitleTextOptionFunc(props.Title))
	}
	if len(props.Labels) > 0 {
		opts = append(opts, charts.XAxisLabelsOptionFunc(props.Labels))
	}

	// Legend: show by default when multiple series
	showLegend := props.ShowLegend || len(props.Series) > 1
	if showLegend && len(seriesNames) > 0 {
		opts = append(opts, charts.LegendLabelsOptionFunc(seriesNames))
	}

	if props.ShowValues {
		opts = append(opts, charts.SeriesShowLabel(true))
	}

	// Apply custom series colors via theme
	seriesColors := resolveSeriesColors(props.Series)
	opts = append(opts, charts.ThemeOptionFunc(
		charts.GetTheme("light").WithSeriesColors(seriesColors),
	))

	// Render
	var p *charts.Painter
	var err error
	if props.Horizontal {
		p, err = charts.HorizontalBarRender(values, opts...)
	} else {
		p, err = charts.BarRender(values, opts...)
	}
	if err != nil {
		return chartError(props.Title, err)
	}

	svgBytes, err := p.Bytes()
	if err != nil {
		return chartError(props.Title, err)
	}

	return wrapChartSVG(string(svgBytes), props.Title, props.Class)
}

// LineChart renders a line chart as inline SVG using the go-analyze/charts library.
//
// Example:
//
//	LineChart(LineChartProps{
//	    Title:  "Monthly Revenue",
//	    Labels: []string{"Jan", "Feb", "Mar"},
//	    Series: []ChartSeries{
//	        {Name: "Revenue", Values: []float64{1200, 1500, 1800}, Color: "blue"},
//	    },
//	    ShowGrid: true,
//	    ShowDots: true,
//	})
func LineChart(props LineChartProps) core.Node {
	if len(props.Series) == 0 {
		return core.Frag()
	}

	width, height := resolveChartDimensions(props.Width, props.Height)

	// Build values as [][]float64
	values := make([][]float64, len(props.Series))
	seriesNames := make([]string, len(props.Series))
	for i, s := range props.Series {
		values[i] = s.Values
		seriesNames[i] = s.Name
	}

	// Build options
	opts := []charts.OptionFunc{
		charts.SVGOutputOptionFunc(),
		charts.DimensionsOptionFunc(width, height),
	}

	if props.Title != "" {
		opts = append(opts, charts.TitleTextOptionFunc(props.Title))
	}
	if len(props.Labels) > 0 {
		opts = append(opts, charts.XAxisLabelsOptionFunc(props.Labels))
	}

	// Legend: show when multiple series
	if len(props.Series) > 1 && len(seriesNames) > 0 {
		opts = append(opts, charts.LegendLabelsOptionFunc(seriesNames))
	}

	// Show data point dots
	if props.ShowDots {
		opts = append(opts, func(opt *charts.ChartOption) {
			opt.Symbol = "circle"
		})
	}

	// Apply custom series colors via theme
	seriesColors := resolveSeriesColors(props.Series)
	opts = append(opts, charts.ThemeOptionFunc(
		charts.GetTheme("light").WithSeriesColors(seriesColors),
	))

	p, err := charts.LineRender(values, opts...)
	if err != nil {
		return chartError(props.Title, err)
	}

	svgBytes, err := p.Bytes()
	if err != nil {
		return chartError(props.Title, err)
	}

	return wrapChartSVG(string(svgBytes), props.Title, props.Class)
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

// AreaChart renders an area chart (filled line chart) as inline SVG.
//
// Example:
//
//	AreaChart(AreaChartProps{
//	    Title:  "Traffic Over Time",
//	    Labels: []string{"Jan", "Feb", "Mar", "Apr"},
//	    Series: []ChartSeries{
//	        {Name: "Organic", Values: []float64{120, 180, 240, 300}, Color: "green"},
//	        {Name: "Paid", Values: []float64{60, 90, 120, 80}, Color: "blue"},
//	    },
//	    Stacked: true,
//	})
func AreaChart(props AreaChartProps) core.Node {
	if len(props.Series) == 0 {
		return core.Frag()
	}

	width, height := resolveChartDimensions(props.Width, props.Height)

	// Build values as [][]float64
	values := make([][]float64, len(props.Series))
	seriesNames := make([]string, len(props.Series))
	for i, s := range props.Series {
		values[i] = s.Values
		seriesNames[i] = s.Name
	}

	// Default opacity
	opacity := props.Opacity
	if opacity == 0 {
		opacity = 128
	}

	// Build options
	opts := []charts.OptionFunc{
		charts.SVGOutputOptionFunc(),
		charts.DimensionsOptionFunc(width, height),
		// Enable area fill
		func(opt *charts.ChartOption) {
			opt.FillArea = charts.Ptr(true)
			opt.FillOpacity = opacity
			if props.Stacked {
				opt.StackSeries = charts.Ptr(true)
			}
		},
	}

	if props.Title != "" {
		opts = append(opts, charts.TitleTextOptionFunc(props.Title))
	}
	if len(props.Labels) > 0 {
		opts = append(opts, charts.XAxisLabelsOptionFunc(props.Labels))
	}

	// Legend: show by default when multiple series
	showLegend := props.ShowLegend || len(props.Series) > 1
	if showLegend && len(seriesNames) > 0 {
		opts = append(opts, charts.LegendLabelsOptionFunc(seriesNames))
	}

	// Show data point dots
	if props.ShowDots {
		opts = append(opts, func(opt *charts.ChartOption) {
			opt.Symbol = "circle"
		})
	}

	// Apply custom series colors via theme
	seriesColors := resolveSeriesColors(props.Series)
	opts = append(opts, charts.ThemeOptionFunc(
		charts.GetTheme("light").WithSeriesColors(seriesColors),
	))

	p, err := charts.LineRender(values, opts...)
	if err != nil {
		return chartError(props.Title, err)
	}

	svgBytes, err := p.Bytes()
	if err != nil {
		return chartError(props.Title, err)
	}

	return wrapChartSVG(string(svgBytes), props.Title, props.Class)
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
