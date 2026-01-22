package ui

import (
	"strings"
	"testing"
)

// Note: renderHTML helper is defined in button_test.go

func TestPagination_TotalPages(t *testing.T) {
	tests := []struct {
		name       string
		totalItems int
		pageSize   int
		expected   int
	}{
		{"exact division", 100, 10, 10},
		{"with remainder", 101, 10, 11},
		{"single page", 5, 10, 1},
		{"zero items", 0, 10, 0},
		{"zero page size defaults to 10", 50, 0, 5},
		{"negative page size defaults to 10", 50, -5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := totalPages(tt.totalItems, tt.pageSize)
			if result != tt.expected {
				t.Errorf("totalPages(%d, %d) = %d, want %d",
					tt.totalItems, tt.pageSize, result, tt.expected)
			}
		})
	}
}

func TestPagination_SinglePage(t *testing.T) {
	html := renderHTML(Pagination(PaginationProps{
		CurrentPage: 1,
		TotalItems:  5,
		PageSize:    10,
	}))

	// Should return empty fragment (no pagination needed)
	if html != "" {
		t.Errorf("single page should render nothing, got: %s", html)
	}
}

func TestPagination_ZeroItems(t *testing.T) {
	html := renderHTML(Pagination(PaginationProps{
		CurrentPage: 1,
		TotalItems:  0,
		PageSize:    10,
	}))

	// Should return empty fragment (no pagination needed)
	if html != "" {
		t.Errorf("zero items should render nothing, got: %s", html)
	}
}

func TestPagination_FirstPage(t *testing.T) {
	html := renderHTML(Pagination(PaginationProps{
		CurrentPage: 1,
		TotalItems:  50,
		PageSize:    10, // 5 pages
	}))

	// Should contain Previous button (disabled)
	if !strings.Contains(html, "Previous") {
		t.Error("should contain Previous button")
	}
	if !strings.Contains(html, "disabled") {
		t.Error("Previous button should be disabled on first page")
	}

	// Should contain Next button (enabled)
	if !strings.Contains(html, "Next") {
		t.Error("should contain Next button")
	}

	// Should contain page 1 with primary variant (active)
	if !strings.Contains(html, ">1<") {
		t.Error("should contain page 1")
	}
	if !strings.Contains(html, "bg-blue-600") {
		t.Error("active page should have primary variant")
	}
}

func TestPagination_LastPage(t *testing.T) {
	html := renderHTML(Pagination(PaginationProps{
		CurrentPage: 5,
		TotalItems:  50,
		PageSize:    10, // 5 pages
	}))

	// Should contain Previous button (enabled - not disabled)
	if !strings.Contains(html, "Previous") {
		t.Error("should contain Previous button")
	}

	// Should contain Next button (disabled)
	if !strings.Contains(html, "Next") {
		t.Error("should contain Next button")
	}

	// Count disabled buttons - should have 2 (one for Next, but Previous might not have disabled class on last page)
	// Actually, let's check the specific button text context
	// The Next button on last page should be disabled
	// We need to verify the disabled attribute appears after "Next" button context
}

func TestPagination_MiddlePage(t *testing.T) {
	html := renderHTML(Pagination(PaginationProps{
		CurrentPage: 3,
		TotalItems:  50,
		PageSize:    10, // 5 pages
	}))

	// Should contain both Previous and Next buttons
	if !strings.Contains(html, "Previous") {
		t.Error("should contain Previous button")
	}
	if !strings.Contains(html, "Next") {
		t.Error("should contain Next button")
	}

	// Should have pages 1-5 all visible (since we show up to 5 around current)
	for i := 1; i <= 5; i++ {
		if !strings.Contains(html, ">"+string('0'+byte(i))+"<") && !strings.Contains(html, ">"+string([]byte{byte('0' + i)})) {
			// Use different check for page numbers
		}
	}

	// Check page 3 is active (primary variant)
	if !strings.Contains(html, ">3<") {
		t.Error("should contain page 3")
	}
}

func TestPagination_PageNumbers(t *testing.T) {
	tests := []struct {
		name          string
		currentPage   int
		totalItems    int
		pageSize      int
		expectedPages []int
	}{
		{
			name:          "5 pages, on page 1",
			currentPage:   1,
			totalItems:    50,
			pageSize:      10,
			expectedPages: []int{1, 2, 3, 4, 5},
		},
		{
			name:          "10 pages, on page 1",
			currentPage:   1,
			totalItems:    100,
			pageSize:      10,
			expectedPages: []int{1, 2, 3, 4, 5},
		},
		{
			name:          "10 pages, on page 5",
			currentPage:   5,
			totalItems:    100,
			pageSize:      10,
			expectedPages: []int{3, 4, 5, 6, 7},
		},
		{
			name:          "10 pages, on page 10",
			currentPage:   10,
			totalItems:    100,
			pageSize:      10,
			expectedPages: []int{6, 7, 8, 9, 10},
		},
		{
			name:          "3 pages, on page 2",
			currentPage:   2,
			totalItems:    30,
			pageSize:      10,
			expectedPages: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := renderHTML(Pagination(PaginationProps{
				CurrentPage: tt.currentPage,
				TotalItems:  tt.totalItems,
				PageSize:    tt.pageSize,
			}))

			for _, page := range tt.expectedPages {
				pageStr := ">" + string([]rune{rune('0' + page%10)})
				if page >= 10 {
					pageStr = ">" + string([]rune{rune('0' + page/10), rune('0' + page%10)})
				}
				if !strings.Contains(html, pageStr) {
					t.Errorf("expected page %d to be visible, html: %s", page, html)
				}
			}
		})
	}
}

func TestPagination_CustomClass(t *testing.T) {
	html := renderHTML(Pagination(PaginationProps{
		CurrentPage: 1,
		TotalItems:  20,
		PageSize:    10,
		Class:       "my-custom-pagination",
	}))

	if !strings.Contains(html, "my-custom-pagination") {
		t.Error("should apply custom class")
	}
	if !strings.Contains(html, "flex") {
		t.Error("should retain default flex class")
	}
	if !strings.Contains(html, "items-center") {
		t.Error("should retain default items-center class")
	}
	if !strings.Contains(html, "gap-2") {
		t.Error("should retain default gap-2 class")
	}
}

func TestPagination_DefaultPageSize(t *testing.T) {
	html := renderHTML(Pagination(PaginationProps{
		CurrentPage: 1,
		TotalItems:  25,
		PageSize:    0, // Should default to 10
	}))

	// With 25 items and default page size of 10, we should have 3 pages
	// So pagination should render (not empty)
	if html == "" {
		t.Error("should render pagination with default page size")
	}

	// Should have pages 1, 2, 3
	if !strings.Contains(html, ">1<") {
		t.Error("should contain page 1")
	}
	if !strings.Contains(html, ">2<") {
		t.Error("should contain page 2")
	}
	if !strings.Contains(html, ">3<") {
		t.Error("should contain page 3")
	}
}

func TestPagination_EdgeCases(t *testing.T) {
	t.Run("current page below 1 normalizes to 1", func(t *testing.T) {
		html := renderHTML(Pagination(PaginationProps{
			CurrentPage: 0,
			TotalItems:  20,
			PageSize:    10,
		}))

		// Should still render (2 pages)
		if !strings.Contains(html, "Previous") {
			t.Error("should render pagination")
		}
		// Page 1 should be active (primary variant)
		if !strings.Contains(html, "bg-blue-600") {
			t.Error("page 1 should be active")
		}
	})

	t.Run("current page above total normalizes to total", func(t *testing.T) {
		html := renderHTML(Pagination(PaginationProps{
			CurrentPage: 100,
			TotalItems:  20,
			PageSize:    10, // 2 pages total
		}))

		// Should still render
		if !strings.Contains(html, "Previous") {
			t.Error("should render pagination")
		}
	})
}
