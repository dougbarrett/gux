//go:build js && wasm

package components

import (
	"fmt"
	"syscall/js"
)

// PaginationProps configures a Pagination component
type PaginationProps struct {
	CurrentPage  int
	TotalPages   int
	TotalItems   int
	ItemsPerPage int
	OnPageChange func(page int)
	ShowInfo     bool // Show "Showing X-Y of Z items"
	MaxVisible   int  // Max page buttons to show (default 5)
}

// Pagination creates a pagination component
type Pagination struct {
	container js.Value
	props     PaginationProps
}

// NewPagination creates a new Pagination component
func NewPagination(props PaginationProps) *Pagination {
	if props.MaxVisible == 0 {
		props.MaxVisible = 5
	}
	if props.TotalPages == 0 && props.TotalItems > 0 && props.ItemsPerPage > 0 {
		props.TotalPages = (props.TotalItems + props.ItemsPerPage - 1) / props.ItemsPerPage
	}

	p := &Pagination{props: props}
	p.render()
	return p
}

func (p *Pagination) render() {
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	container.Set("className", "flex items-center justify-between")

	// Info section
	if p.props.ShowInfo && p.props.TotalItems > 0 {
		info := document.Call("createElement", "div")
		info.Set("className", "text-sm text-gray-600 dark:text-gray-400")
		start := (p.props.CurrentPage-1)*p.props.ItemsPerPage + 1
		end := start + p.props.ItemsPerPage - 1
		if end > p.props.TotalItems {
			end = p.props.TotalItems
		}
		info.Set("textContent", fmt.Sprintf("Showing %d-%d of %d items", start, end, p.props.TotalItems))
		container.Call("appendChild", info)
	}

	// Navigation
	nav := document.Call("createElement", "nav")
	nav.Set("className", "flex items-center gap-1")
	nav.Set("role", "navigation")
	nav.Set("aria-label", "Pagination")

	// Previous button
	prevBtn := p.createNavButton("←", "Previous", p.props.CurrentPage > 1, func() {
		if p.props.OnPageChange != nil && p.props.CurrentPage > 1 {
			p.props.OnPageChange(p.props.CurrentPage - 1)
		}
	})
	nav.Call("appendChild", prevBtn)

	// Page numbers
	pages := p.getVisiblePages()
	for _, page := range pages {
		if page == -1 {
			// Ellipsis
			ellipsis := document.Call("createElement", "span")
			ellipsis.Set("className", "px-2 text-gray-400 dark:text-gray-500")
			ellipsis.Set("textContent", "...")
			nav.Call("appendChild", ellipsis)
		} else {
			btn := p.createPageButton(page)
			nav.Call("appendChild", btn)
		}
	}

	// Next button
	nextBtn := p.createNavButton("→", "Next", p.props.CurrentPage < p.props.TotalPages, func() {
		if p.props.OnPageChange != nil && p.props.CurrentPage < p.props.TotalPages {
			p.props.OnPageChange(p.props.CurrentPage + 1)
		}
	})
	nav.Call("appendChild", nextBtn)

	container.Call("appendChild", nav)
	p.container = container
}

func (p *Pagination) createNavButton(text, label string, enabled bool, onClick func()) js.Value {
	document := js.Global().Get("document")
	btn := document.Call("createElement", "button")

	className := "px-3 py-1 rounded border text-sm"
	if enabled {
		className += " border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300 cursor-pointer"
	} else {
		className += " border-gray-200 dark:border-gray-700 text-gray-400 dark:text-gray-500 cursor-not-allowed"
	}

	btn.Set("className", className)
	btn.Set("textContent", text)
	btn.Set("aria-label", label)
	btn.Set("disabled", !enabled)

	if enabled {
		btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			onClick()
			return nil
		}))
	}

	return btn
}

func (p *Pagination) createPageButton(page int) js.Value {
	document := js.Global().Get("document")
	btn := document.Call("createElement", "button")

	isCurrent := page == p.props.CurrentPage
	className := "w-8 h-8 rounded text-sm"
	if isCurrent {
		className += " bg-blue-500 text-white"
	} else {
		className += " hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300 cursor-pointer"
	}

	btn.Set("className", className)
	btn.Set("textContent", fmt.Sprintf("%d", page))
	if isCurrent {
		btn.Set("aria-current", "page")
	}

	if !isCurrent && p.props.OnPageChange != nil {
		pageNum := page
		btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			p.props.OnPageChange(pageNum)
			return nil
		}))
	}

	return btn
}

func (p *Pagination) getVisiblePages() []int {
	total := p.props.TotalPages
	current := p.props.CurrentPage
	max := p.props.MaxVisible

	if total <= max {
		pages := make([]int, total)
		for i := 0; i < total; i++ {
			pages[i] = i + 1
		}
		return pages
	}

	var pages []int
	half := max / 2

	// Always show first page
	pages = append(pages, 1)

	start := current - half
	end := current + half

	if start <= 2 {
		// Near the beginning
		for i := 2; i <= max-1 && i < total; i++ {
			pages = append(pages, i)
		}
		pages = append(pages, -1) // ellipsis
	} else if end >= total-1 {
		// Near the end
		pages = append(pages, -1) // ellipsis
		for i := total - max + 2; i < total; i++ {
			pages = append(pages, i)
		}
	} else {
		// Middle
		pages = append(pages, -1) // ellipsis
		for i := start; i <= end; i++ {
			pages = append(pages, i)
		}
		pages = append(pages, -1) // ellipsis
	}

	// Always show last page
	pages = append(pages, total)

	return pages
}

// Element returns the container DOM element
func (p *Pagination) Element() js.Value {
	return p.container
}

// SetPage updates the current page and re-renders
func (p *Pagination) SetPage(page int) {
	p.props.CurrentPage = page
	p.render()
}

// SimplePagination creates a basic pagination without info
func SimplePagination(currentPage, totalPages int, onPageChange func(int)) *Pagination {
	return NewPagination(PaginationProps{
		CurrentPage:  currentPage,
		TotalPages:   totalPages,
		OnPageChange: onPageChange,
	})
}

// ItemPagination creates pagination based on total items
func ItemPagination(currentPage, totalItems, itemsPerPage int, onPageChange func(int)) *Pagination {
	return NewPagination(PaginationProps{
		CurrentPage:  currentPage,
		TotalItems:   totalItems,
		ItemsPerPage: itemsPerPage,
		OnPageChange: onPageChange,
		ShowInfo:     true,
	})
}
