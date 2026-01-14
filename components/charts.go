//go:build js && wasm

package components

import (
	"fmt"
	"math"
	"syscall/js"
)

// ChartData represents data for charts
type ChartData struct {
	Label string
	Value float64
	Color string
}

// BarChartProps configures a BarChart
type BarChartProps struct {
	Data       []ChartData
	Width      string // default "100%"
	Height     string // default "200px"
	ShowLabels bool
	ShowValues bool
	Horizontal bool
	BarColor   string // default color if not specified per-item
	ClassName  string
}

// BarChart creates a bar chart component
func BarChart(props BarChartProps) js.Value {
	document := js.Global().Get("document")

	if props.Width == "" {
		props.Width = "100%"
	}
	if props.Height == "" {
		props.Height = "200px"
	}
	if props.BarColor == "" {
		props.BarColor = "#3b82f6" // blue-500
	}

	container := document.Call("createElement", "div")
	className := "bar-chart w-full"
	if props.ClassName != "" {
		className += " " + props.ClassName
	}
	container.Set("className", className)
	container.Get("style").Set("width", props.Width)
	container.Get("style").Set("height", props.Height)
	container.Get("style").Set("maxWidth", "100%")

	if len(props.Data) == 0 {
		return container
	}

	// Find max value for scaling
	maxVal := 0.0
	for _, d := range props.Data {
		if d.Value > maxVal {
			maxVal = d.Value
		}
	}

	if props.Horizontal {
		container.Get("style").Set("display", "flex")
		container.Get("style").Set("flexDirection", "column")
		container.Get("style").Set("gap", "8px")

		for _, d := range props.Data {
			row := document.Call("createElement", "div")
			row.Set("className", "flex items-center gap-2")

			if props.ShowLabels {
				label := document.Call("createElement", "div")
				label.Set("className", "w-20 text-sm text-gray-600 truncate")
				label.Set("textContent", d.Label)
				row.Call("appendChild", label)
			}

			barContainer := document.Call("createElement", "div")
			barContainer.Set("className", "flex-1 bg-gray-100 rounded h-6 overflow-hidden")

			bar := document.Call("createElement", "div")
			bar.Set("className", "h-full rounded transition-all duration-300")
			color := d.Color
			if color == "" {
				color = props.BarColor
			}
			bar.Get("style").Set("backgroundColor", color)
			percentage := (d.Value / maxVal) * 100
			bar.Get("style").Set("width", fmt.Sprintf("%.1f%%", percentage))

			barContainer.Call("appendChild", bar)
			row.Call("appendChild", barContainer)

			if props.ShowValues {
				value := document.Call("createElement", "div")
				value.Set("className", "w-12 text-sm text-gray-700 text-right")
				value.Set("textContent", formatNumber(d.Value))
				row.Call("appendChild", value)
			}

			container.Call("appendChild", row)
		}
	} else {
		// Vertical bars
		container.Get("style").Set("display", "flex")
		container.Get("style").Set("alignItems", "flex-end")
		container.Get("style").Set("gap", "4px")
		container.Get("style").Set("paddingTop", "20px")

		for _, d := range props.Data {
			col := document.Call("createElement", "div")
			col.Set("className", "flex-1 flex flex-col items-center")

			if props.ShowValues {
				value := document.Call("createElement", "div")
				value.Set("className", "text-xs text-gray-600 mb-1")
				value.Set("textContent", formatNumber(d.Value))
				col.Call("appendChild", value)
			}

			bar := document.Call("createElement", "div")
			bar.Set("className", "w-full rounded-t transition-all duration-300")
			color := d.Color
			if color == "" {
				color = props.BarColor
			}
			bar.Get("style").Set("backgroundColor", color)
			percentage := (d.Value / maxVal) * 100
			bar.Get("style").Set("height", fmt.Sprintf("%.1f%%", percentage))
			bar.Get("style").Set("minHeight", "4px")
			col.Call("appendChild", bar)

			if props.ShowLabels {
				label := document.Call("createElement", "div")
				label.Set("className", "text-xs text-gray-600 mt-1 truncate w-full text-center")
				label.Set("textContent", d.Label)
				col.Call("appendChild", label)
			}

			container.Call("appendChild", col)
		}
	}

	return container
}

// LineChartProps configures a LineChart
type LineChartProps struct {
	Data       []ChartData
	Width      string
	Height     string
	LineColor  string
	FillColor  string // Gradient fill under line
	ShowPoints bool
	ShowLabels bool
	ShowGrid   bool
	ClassName  string
}

// LineChart creates a line chart using SVG
func LineChart(props LineChartProps) js.Value {
	document := js.Global().Get("document")

	if props.Width == "" {
		props.Width = "100%"
	}
	if props.Height == "" {
		props.Height = "200px"
	}
	if props.LineColor == "" {
		props.LineColor = "#3b82f6"
	}

	container := document.Call("createElement", "div")
	className := "w-full"
	if props.ClassName != "" {
		className += " " + props.ClassName
	}
	container.Set("className", className)
	container.Get("style").Set("width", props.Width)
	container.Get("style").Set("height", props.Height)
	container.Get("style").Set("maxWidth", "100%")
	container.Get("style").Set("position", "relative")

	if len(props.Data) == 0 {
		return container
	}

	// SVG dimensions
	svgWidth := 400
	svgHeight := 200
	padding := 30

	// Create SVG
	svg := document.Call("createElementNS", "http://www.w3.org/2000/svg", "svg")
	svg.Call("setAttribute", "width", "100%")
	svg.Call("setAttribute", "height", "100%")
	svg.Call("setAttribute", "viewBox", fmt.Sprintf("0 0 %d %d", svgWidth, svgHeight))
	svg.Get("style").Set("overflow", "visible")

	// Find min/max
	minVal, maxVal := props.Data[0].Value, props.Data[0].Value
	for _, d := range props.Data {
		if d.Value < minVal {
			minVal = d.Value
		}
		if d.Value > maxVal {
			maxVal = d.Value
		}
	}
	if minVal == maxVal {
		minVal -= 1
		maxVal += 1
	}

	// Calculate points
	chartWidth := float64(svgWidth - 2*padding)
	chartHeight := float64(svgHeight - 2*padding)
	points := make([]struct{ x, y float64 }, len(props.Data))

	for i, d := range props.Data {
		x := float64(padding) + (float64(i)/float64(len(props.Data)-1))*chartWidth
		y := float64(svgHeight-padding) - ((d.Value-minVal)/(maxVal-minVal))*chartHeight
		points[i] = struct{ x, y float64 }{x, y}
	}

	// Draw grid
	if props.ShowGrid {
		for i := 0; i <= 4; i++ {
			y := float64(padding) + (float64(i)/4)*chartHeight
			line := document.Call("createElementNS", "http://www.w3.org/2000/svg", "line")
			line.Call("setAttribute", "x1", fmt.Sprintf("%d", padding))
			line.Call("setAttribute", "y1", fmt.Sprintf("%.1f", y))
			line.Call("setAttribute", "x2", fmt.Sprintf("%d", svgWidth-padding))
			line.Call("setAttribute", "y2", fmt.Sprintf("%.1f", y))
			line.Call("setAttribute", "stroke", "#e5e7eb")
			line.Call("setAttribute", "stroke-width", "1")
			svg.Call("appendChild", line)
		}
	}

	// Draw fill gradient
	if props.FillColor != "" {
		pathData := fmt.Sprintf("M %.1f %.1f", points[0].x, float64(svgHeight-padding))
		for _, p := range points {
			pathData += fmt.Sprintf(" L %.1f %.1f", p.x, p.y)
		}
		pathData += fmt.Sprintf(" L %.1f %.1f Z", points[len(points)-1].x, float64(svgHeight-padding))

		fill := document.Call("createElementNS", "http://www.w3.org/2000/svg", "path")
		fill.Call("setAttribute", "d", pathData)
		fill.Call("setAttribute", "fill", props.FillColor)
		fill.Call("setAttribute", "opacity", "0.3")
		svg.Call("appendChild", fill)
	}

	// Draw line
	pathData := fmt.Sprintf("M %.1f %.1f", points[0].x, points[0].y)
	for i := 1; i < len(points); i++ {
		pathData += fmt.Sprintf(" L %.1f %.1f", points[i].x, points[i].y)
	}

	line := document.Call("createElementNS", "http://www.w3.org/2000/svg", "path")
	line.Call("setAttribute", "d", pathData)
	line.Call("setAttribute", "fill", "none")
	line.Call("setAttribute", "stroke", props.LineColor)
	line.Call("setAttribute", "stroke-width", "2")
	line.Call("setAttribute", "stroke-linecap", "round")
	line.Call("setAttribute", "stroke-linejoin", "round")
	svg.Call("appendChild", line)

	// Draw points
	if props.ShowPoints {
		for _, p := range points {
			circle := document.Call("createElementNS", "http://www.w3.org/2000/svg", "circle")
			circle.Call("setAttribute", "cx", fmt.Sprintf("%.1f", p.x))
			circle.Call("setAttribute", "cy", fmt.Sprintf("%.1f", p.y))
			circle.Call("setAttribute", "r", "4")
			circle.Call("setAttribute", "fill", props.LineColor)
			svg.Call("appendChild", circle)
		}
	}

	// Draw labels
	if props.ShowLabels {
		for i, d := range props.Data {
			text := document.Call("createElementNS", "http://www.w3.org/2000/svg", "text")
			text.Call("setAttribute", "x", fmt.Sprintf("%.1f", points[i].x))
			text.Call("setAttribute", "y", fmt.Sprintf("%d", svgHeight-5))
			text.Call("setAttribute", "text-anchor", "middle")
			text.Call("setAttribute", "font-size", "10")
			text.Call("setAttribute", "fill", "#6b7280")
			text.Set("textContent", d.Label)
			svg.Call("appendChild", text)
		}
	}

	container.Call("appendChild", svg)
	return container
}

// PieChartProps configures a PieChart
type PieChartProps struct {
	Data       []ChartData
	Size       string // Width and height (default "200px")
	ShowLabels bool
	ShowLegend bool
	DonutWidth int    // If > 0, creates a donut chart
	ClassName  string
}

// PieChart creates a pie/donut chart using SVG
func PieChart(props PieChartProps) js.Value {
	document := js.Global().Get("document")

	if props.Size == "" {
		props.Size = "200px"
	}

	container := document.Call("createElement", "div")
	// Responsive: stack vertically on mobile, horizontal on larger screens
	className := "flex flex-col sm:flex-row items-center gap-4"
	if props.ClassName != "" {
		className += " " + props.ClassName
	}
	container.Set("className", className)

	if len(props.Data) == 0 {
		return container
	}

	// Default colors
	defaultColors := []string{"#3b82f6", "#ef4444", "#22c55e", "#f59e0b", "#8b5cf6", "#ec4899", "#06b6d4", "#84cc16"}

	// Calculate total
	total := 0.0
	for _, d := range props.Data {
		total += d.Value
	}

	// SVG setup
	size := 200
	cx, cy := float64(size)/2, float64(size)/2
	radius := float64(size)/2 - 10
	innerRadius := 0.0
	if props.DonutWidth > 0 {
		innerRadius = radius - float64(props.DonutWidth)
	}

	svg := document.Call("createElementNS", "http://www.w3.org/2000/svg", "svg")
	svg.Call("setAttribute", "width", props.Size)
	svg.Call("setAttribute", "height", props.Size)
	svg.Call("setAttribute", "viewBox", fmt.Sprintf("0 0 %d %d", size, size))

	// Draw slices
	startAngle := -math.Pi / 2
	for i, d := range props.Data {
		color := d.Color
		if color == "" {
			color = defaultColors[i%len(defaultColors)]
		}

		sliceAngle := (d.Value / total) * 2 * math.Pi
		endAngle := startAngle + sliceAngle

		// Calculate arc points
		x1 := cx + radius*math.Cos(startAngle)
		y1 := cy + radius*math.Sin(startAngle)
		x2 := cx + radius*math.Cos(endAngle)
		y2 := cy + radius*math.Sin(endAngle)

		largeArc := 0
		if sliceAngle > math.Pi {
			largeArc = 1
		}

		var pathData string
		if innerRadius > 0 {
			// Donut
			ix1 := cx + innerRadius*math.Cos(startAngle)
			iy1 := cy + innerRadius*math.Sin(startAngle)
			ix2 := cx + innerRadius*math.Cos(endAngle)
			iy2 := cy + innerRadius*math.Sin(endAngle)

			pathData = fmt.Sprintf("M %.2f %.2f A %.2f %.2f 0 %d 1 %.2f %.2f L %.2f %.2f A %.2f %.2f 0 %d 0 %.2f %.2f Z",
				x1, y1, radius, radius, largeArc, x2, y2,
				ix2, iy2, innerRadius, innerRadius, largeArc, ix1, iy1)
		} else {
			// Pie
			pathData = fmt.Sprintf("M %.2f %.2f L %.2f %.2f A %.2f %.2f 0 %d 1 %.2f %.2f Z",
				cx, cy, x1, y1, radius, radius, largeArc, x2, y2)
		}

		slice := document.Call("createElementNS", "http://www.w3.org/2000/svg", "path")
		slice.Call("setAttribute", "d", pathData)
		slice.Call("setAttribute", "fill", color)
		slice.Call("setAttribute", "stroke", "white")
		slice.Call("setAttribute", "stroke-width", "2")
		svg.Call("appendChild", slice)

		// Label on slice
		if props.ShowLabels && sliceAngle > 0.3 { // Only show label if slice is big enough
			labelAngle := startAngle + sliceAngle/2
			labelRadius := radius * 0.7
			if innerRadius > 0 {
				labelRadius = (radius + innerRadius) / 2
			}
			lx := cx + labelRadius*math.Cos(labelAngle)
			ly := cy + labelRadius*math.Sin(labelAngle)

			text := document.Call("createElementNS", "http://www.w3.org/2000/svg", "text")
			text.Call("setAttribute", "x", fmt.Sprintf("%.1f", lx))
			text.Call("setAttribute", "y", fmt.Sprintf("%.1f", ly))
			text.Call("setAttribute", "text-anchor", "middle")
			text.Call("setAttribute", "dominant-baseline", "middle")
			text.Call("setAttribute", "font-size", "12")
			text.Call("setAttribute", "fill", "white")
			text.Call("setAttribute", "font-weight", "bold")
			percentage := (d.Value / total) * 100
			text.Set("textContent", fmt.Sprintf("%.0f%%", percentage))
			svg.Call("appendChild", text)
		}

		startAngle = endAngle
	}

	container.Call("appendChild", svg)

	// Legend
	if props.ShowLegend {
		legend := document.Call("createElement", "div")
		legend.Set("className", "space-y-1")

		for i, d := range props.Data {
			color := d.Color
			if color == "" {
				color = defaultColors[i%len(defaultColors)]
			}

			item := document.Call("createElement", "div")
			item.Set("className", "flex items-center gap-2 text-sm")

			dot := document.Call("createElement", "div")
			dot.Set("className", "w-3 h-3 rounded-full")
			dot.Get("style").Set("backgroundColor", color)
			item.Call("appendChild", dot)

			label := document.Call("createElement", "span")
			label.Set("className", "text-gray-700")
			label.Set("textContent", d.Label)
			item.Call("appendChild", label)

			value := document.Call("createElement", "span")
			value.Set("className", "text-gray-500")
			value.Set("textContent", fmt.Sprintf("(%.0f)", d.Value))
			item.Call("appendChild", value)

			legend.Call("appendChild", item)
		}

		container.Call("appendChild", legend)
	}

	return container
}

// formatNumber formats a number for display
func formatNumber(n float64) string {
	if n == math.Floor(n) {
		return fmt.Sprintf("%.0f", n)
	}
	return fmt.Sprintf("%.1f", n)
}

// SimpleBarChart creates a quick bar chart from labels and values
func SimpleBarChart(labels []string, values []float64) js.Value {
	data := make([]ChartData, len(labels))
	for i := range labels {
		data[i] = ChartData{Label: labels[i], Value: values[i]}
	}
	return BarChart(BarChartProps{Data: data, ShowLabels: true, ShowValues: true})
}

// SimpleLineChart creates a quick line chart from labels and values
func SimpleLineChart(labels []string, values []float64) js.Value {
	data := make([]ChartData, len(labels))
	for i := range labels {
		data[i] = ChartData{Label: labels[i], Value: values[i]}
	}
	return LineChart(LineChartProps{Data: data, ShowLabels: true, ShowPoints: true, ShowGrid: true, FillColor: "#3b82f6"})
}

// SimplePieChart creates a quick pie chart from labels and values
func SimplePieChart(labels []string, values []float64) js.Value {
	data := make([]ChartData, len(labels))
	for i := range labels {
		data[i] = ChartData{Label: labels[i], Value: values[i]}
	}
	return PieChart(PieChartProps{Data: data, ShowLegend: true})
}

// DonutChart creates a donut chart
func DonutChart(data []ChartData, donutWidth int) js.Value {
	return PieChart(PieChartProps{Data: data, DonutWidth: donutWidth, ShowLabels: true, ShowLegend: true})
}
