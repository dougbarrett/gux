package ui

import (
	"strings"

	"github.com/go-analyze/charts"
)

// chartColorHex maps short Tailwind color names to hex values (500 shade).
var chartColorHex = map[string]string{
	"blue":   "#3b82f6",
	"green":  "#22c55e",
	"red":    "#ef4444",
	"purple": "#a855f7",
	"yellow": "#eab308",
	"orange": "#f97316",
	"pink":   "#ec4899",
	"indigo": "#6366f1",
	"cyan":   "#06b6d4",
	"teal":   "#14b8a6",
	"gray":   "#6b7280",
}

// defaultChartPalette is the default color cycle for chart series.
var defaultChartPalette = []string{
	"#3b82f6", // blue
	"#22c55e", // green
	"#ef4444", // red
	"#a855f7", // purple
	"#eab308", // yellow
	"#f97316", // orange
}

// resolveChartColor converts a Tailwind color name or hex value to a charts.Color.
func resolveChartColor(color string) charts.Color {
	if color == "" {
		return charts.Color{}
	}
	if strings.HasPrefix(color, "#") {
		return charts.ParseColor(color)
	}
	if hex, ok := chartColorHex[color]; ok {
		return charts.ParseColor(hex)
	}
	// Try as a known CSS color name
	return charts.ParseColor(color)
}

// resolveSeriesColors builds a []charts.Color from the series, falling back to the default palette.
func resolveSeriesColors(series []ChartSeries) []charts.Color {
	colors := make([]charts.Color, len(series))
	for i, s := range series {
		if s.Color != "" {
			colors[i] = resolveChartColor(s.Color)
		} else {
			hex := defaultChartPalette[i%len(defaultChartPalette)]
			colors[i] = charts.ParseColor(hex)
		}
	}
	return colors
}
