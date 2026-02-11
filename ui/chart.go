//go:build !js || !wasm

package ui

import (
	"github.com/dougbarrett/gux/core"
	"github.com/go-analyze/charts"
)

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
