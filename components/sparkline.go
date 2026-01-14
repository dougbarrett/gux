//go:build js && wasm

package components

import (
	"fmt"
	"syscall/js"
)

// SparklineType defines the type of sparkline
type SparklineType string

const (
	SparklineLine SparklineType = "line"
	SparklineBar  SparklineType = "bar"
	SparklineArea SparklineType = "area"
)

// SparklineProps configures a Sparkline component
type SparklineProps struct {
	Data      []float64
	Type      SparklineType
	Width     string // default "100px"
	Height    string // default "24px"
	Color     string // default "#3b82f6"
	FillColor string // For area type
	ShowMin   bool   // Highlight minimum point
	ShowMax   bool   // Highlight maximum point
	ClassName string
}

// Sparkline creates a mini inline chart component
func Sparkline(props SparklineProps) js.Value {
	document := js.Global().Get("document")

	if props.Width == "" {
		props.Width = "100px"
	}
	if props.Height == "" {
		props.Height = "24px"
	}
	if props.Color == "" {
		props.Color = "#3b82f6"
	}
	if props.Type == "" {
		props.Type = SparklineLine
	}

	container := document.Call("createElement", "span")
	container.Set("className", "inline-block align-middle "+props.ClassName)
	container.Get("style").Set("width", props.Width)
	container.Get("style").Set("height", props.Height)

	if len(props.Data) == 0 {
		return container
	}

	// Find min/max
	minVal, maxVal := props.Data[0], props.Data[0]
	minIdx, maxIdx := 0, 0
	for i, v := range props.Data {
		if v < minVal {
			minVal = v
			minIdx = i
		}
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}

	// Handle flat data
	if minVal == maxVal {
		minVal -= 1
		maxVal += 1
	}

	// SVG dimensions (use fixed internal dimensions, scale with viewBox)
	svgWidth := 100
	svgHeight := 24
	padding := 2

	svg := document.Call("createElementNS", "http://www.w3.org/2000/svg", "svg")
	svg.Call("setAttribute", "width", "100%")
	svg.Call("setAttribute", "height", "100%")
	svg.Call("setAttribute", "viewBox", fmt.Sprintf("0 0 %d %d", svgWidth, svgHeight))
	svg.Call("setAttribute", "preserveAspectRatio", "none")

	chartWidth := float64(svgWidth - 2*padding)
	chartHeight := float64(svgHeight - 2*padding)

	switch props.Type {
	case SparklineBar:
		barWidth := chartWidth / float64(len(props.Data)) * 0.8
		gap := chartWidth / float64(len(props.Data)) * 0.2

		for i, v := range props.Data {
			x := float64(padding) + float64(i)*(barWidth+gap)
			height := ((v - minVal) / (maxVal - minVal)) * chartHeight
			y := float64(svgHeight-padding) - height

			rect := document.Call("createElementNS", "http://www.w3.org/2000/svg", "rect")
			rect.Call("setAttribute", "x", fmt.Sprintf("%.1f", x))
			rect.Call("setAttribute", "y", fmt.Sprintf("%.1f", y))
			rect.Call("setAttribute", "width", fmt.Sprintf("%.1f", barWidth))
			rect.Call("setAttribute", "height", fmt.Sprintf("%.1f", height))

			// Highlight min/max
			if props.ShowMin && i == minIdx {
				rect.Call("setAttribute", "fill", "#ef4444") // red
			} else if props.ShowMax && i == maxIdx {
				rect.Call("setAttribute", "fill", "#22c55e") // green
			} else {
				rect.Call("setAttribute", "fill", props.Color)
			}

			svg.Call("appendChild", rect)
		}

	case SparklineLine, SparklineArea:
		// Calculate points
		points := make([]struct{ x, y float64 }, len(props.Data))
		for i, v := range props.Data {
			x := float64(padding) + (float64(i)/float64(len(props.Data)-1))*chartWidth
			y := float64(svgHeight-padding) - ((v-minVal)/(maxVal-minVal))*chartHeight
			points[i] = struct{ x, y float64 }{x, y}
		}

		// Draw area fill
		if props.Type == SparklineArea || props.FillColor != "" {
			fillColor := props.FillColor
			if fillColor == "" {
				fillColor = props.Color
			}

			pathData := fmt.Sprintf("M %.1f %.1f", points[0].x, float64(svgHeight-padding))
			for _, p := range points {
				pathData += fmt.Sprintf(" L %.1f %.1f", p.x, p.y)
			}
			pathData += fmt.Sprintf(" L %.1f %.1f Z", points[len(points)-1].x, float64(svgHeight-padding))

			fill := document.Call("createElementNS", "http://www.w3.org/2000/svg", "path")
			fill.Call("setAttribute", "d", pathData)
			fill.Call("setAttribute", "fill", fillColor)
			fill.Call("setAttribute", "opacity", "0.2")
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
		line.Call("setAttribute", "stroke", props.Color)
		line.Call("setAttribute", "stroke-width", "1.5")
		line.Call("setAttribute", "stroke-linecap", "round")
		line.Call("setAttribute", "stroke-linejoin", "round")
		svg.Call("appendChild", line)

		// Highlight min/max points
		if props.ShowMin {
			circle := document.Call("createElementNS", "http://www.w3.org/2000/svg", "circle")
			circle.Call("setAttribute", "cx", fmt.Sprintf("%.1f", points[minIdx].x))
			circle.Call("setAttribute", "cy", fmt.Sprintf("%.1f", points[minIdx].y))
			circle.Call("setAttribute", "r", "2")
			circle.Call("setAttribute", "fill", "#ef4444")
			svg.Call("appendChild", circle)
		}
		if props.ShowMax {
			circle := document.Call("createElementNS", "http://www.w3.org/2000/svg", "circle")
			circle.Call("setAttribute", "cx", fmt.Sprintf("%.1f", points[maxIdx].x))
			circle.Call("setAttribute", "cy", fmt.Sprintf("%.1f", points[maxIdx].y))
			circle.Call("setAttribute", "r", "2")
			circle.Call("setAttribute", "fill", "#22c55e")
			svg.Call("appendChild", circle)
		}
	}

	container.Call("appendChild", svg)
	return container
}

// LineSparkline creates a simple line sparkline
func LineSparkline(data []float64) js.Value {
	return Sparkline(SparklineProps{Data: data, Type: SparklineLine})
}

// BarSparkline creates a simple bar sparkline
func BarSparkline(data []float64) js.Value {
	return Sparkline(SparklineProps{Data: data, Type: SparklineBar})
}

// AreaSparkline creates a simple area sparkline
func AreaSparkline(data []float64) js.Value {
	return Sparkline(SparklineProps{Data: data, Type: SparklineArea})
}

// TrendSparkline creates a sparkline with trend indicators (min/max highlighted)
func TrendSparkline(data []float64) js.Value {
	return Sparkline(SparklineProps{
		Data:    data,
		Type:    SparklineLine,
		ShowMin: true,
		ShowMax: true,
	})
}

// ColoredSparkline creates a sparkline with custom color
func ColoredSparkline(data []float64, color string) js.Value {
	return Sparkline(SparklineProps{Data: data, Type: SparklineLine, Color: color})
}
