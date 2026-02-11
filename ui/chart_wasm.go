//go:build js && wasm

package ui

import "github.com/dougbarrett/gux/core"

// BarChart returns an empty fragment in WASM. Charts are SSR-rendered as SVG;
// the hydrated DOM already contains the chart.
func BarChart(props BarChartProps) core.Node { return core.Frag() }

// LineChart returns an empty fragment in WASM. Charts are SSR-rendered as SVG;
// the hydrated DOM already contains the chart.
func LineChart(props LineChartProps) core.Node { return core.Frag() }

// AreaChart returns an empty fragment in WASM. Charts are SSR-rendered as SVG;
// the hydrated DOM already contains the chart.
func AreaChart(props AreaChartProps) core.Node { return core.Frag() }
