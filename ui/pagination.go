package ui

import (
	"fmt"

	"github.com/dougbarrett/gux/core"
)

// PaginationProps configures the Pagination component.
type PaginationProps struct {
	CurrentPage  int       // 1-indexed current page
	TotalItems   int       // Total number of items
	PageSize     int       // Items per page (default: 10)
	OnPageChange func(int) // Called with new page number
	Class        string
}

// totalPages calculates the total number of pages.
func totalPages(totalItems, pageSize int) int {
	if pageSize <= 0 {
		pageSize = 10
	}
	return (totalItems + pageSize - 1) / pageSize
}

// Pagination renders page navigation controls with previous/next buttons and page numbers.
// Uses 1-indexed pages (user-facing convention).
//
// Example:
//
//	Pagination(PaginationProps{
//	    CurrentPage: 1,
//	    TotalItems: 100,
//	    PageSize: 10,
//	    OnPageChange: func(page int) { /* update state */ },
//	})
func Pagination(props PaginationProps) core.Node {
	pageSize := props.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	total := totalPages(props.TotalItems, pageSize)
	current := props.CurrentPage
	if current < 1 {
		current = 1
	}
	if current > total && total > 0 {
		current = total
	}

	// Don't render if only one page or no pages
	if total <= 1 {
		return core.Frag()
	}

	var children []core.Node

	// Previous button
	prevDisabled := current <= 1
	children = append(children, Button(ButtonProps{
		Variant:  ButtonOutline,
		Size:     ButtonSM,
		Disabled: prevDisabled,
		OnClick: func() {
			if props.OnPageChange != nil && current > 1 {
				props.OnPageChange(current - 1)
			}
		},
		Children: []core.Node{core.Text("Previous")},
	}))

	// Page numbers (show up to 5 pages around current)
	start := current - 2
	if start < 1 {
		start = 1
	}
	end := start + 4
	if end > total {
		end = total
		start = end - 4
		if start < 1 {
			start = 1
		}
	}

	for i := start; i <= end; i++ {
		page := i // capture for closure
		isActive := page == current

		variant := ButtonOutline
		if isActive {
			variant = ButtonPrimary
		}

		children = append(children, Button(ButtonProps{
			Variant: variant,
			Size:    ButtonSM,
			OnClick: func() {
				if props.OnPageChange != nil && page != current {
					props.OnPageChange(page)
				}
			},
			Children: []core.Node{core.Text(fmt.Sprintf("%d", page))},
		}))
	}

	// Next button
	nextDisabled := current >= total
	children = append(children, Button(ButtonProps{
		Variant:  ButtonOutline,
		Size:     ButtonSM,
		Disabled: nextDisabled,
		OnClick: func() {
			if props.OnPageChange != nil && current < total {
				props.OnPageChange(current + 1)
			}
		},
		Children: []core.Node{core.Text("Next")},
	}))

	class := MergeClasses(
		"flex items-center gap-2",
		props.Class,
	)

	return core.Div(core.Class(class), children...)
}
