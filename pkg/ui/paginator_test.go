package ui

import (
	"reflect"
	"testing"
)

type PaginatorInput struct {
	count int64
	limit int64
	page  int64
}
type PaginatorExpected struct {
	length int
	pages  []Page
}

func TestPaginator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    PaginatorInput
		output   Page
		expected PaginatorExpected
	}{
		{
			name:  "Count10Limit10Page1",
			input: PaginatorInput{count: 10, limit: 10, page: 1},
			expected: PaginatorExpected{
				length: 3,
				pages: []Page{
					{Value: "<", Href: "?page=0&limit=10", Disabled: true},
					{Value: "1", Href: "?page=1&limit=10", Active: true},
					{Value: ">", Href: "?page=2&limit=10", Disabled: true},
				},
			},
		},
		{
			name:  "Count40Limit10Page1",
			input: PaginatorInput{count: 40, limit: 10, page: 1},
			expected: PaginatorExpected{
				length: 6,
				pages: []Page{
					{Value: "<", Href: "?page=0&limit=10", Disabled: true},
					{Value: "1", Href: "?page=1&limit=10", Active: true},
					{Value: "2", Href: "?page=2&limit=10"},
					{Value: "3", Href: "?page=3&limit=10"},
					{Value: "4", Href: "?page=4&limit=10"},
					{Value: ">", Href: "?page=2&limit=10"},
				},
			},
		},
		{
			name:  "Count100Limit10Page1",
			input: PaginatorInput{count: 100, limit: 10, page: 1},
			expected: PaginatorExpected{
				length: 7,
				pages: []Page{
					{Value: "<", Href: "?page=0&limit=10", Disabled: true},
					{Value: "1", Href: "?page=1&limit=10", Active: true},
					{Value: "2", Href: "?page=2&limit=10"},
					{Value: "3", Href: "?page=3&limit=10"},
					{Value: "...", Href: "", Disabled: true},
					{Value: "10", Href: "?page=10&limit=10"},
					{Value: ">", Href: "?page=2&limit=10"},
				},
			},
		},
		{
			name:  "Count100Limit10Page4",
			input: PaginatorInput{count: 100, limit: 10, page: 4},
			expected: PaginatorExpected{
				length: 9,
				pages: []Page{
					{Value: "<", Href: "?page=3&limit=10"},
					{Value: "1", Href: "?page=1&limit=10"},
					{Value: "...", Href: "", Disabled: true},
					{Value: "3", Href: "?page=3&limit=10"},
					{Value: "4", Href: "?page=4&limit=10", Active: true},
					{Value: "5", Href: "?page=5&limit=10"},
					{Value: "...", Href: "", Disabled: true},
					{Value: "10", Href: "?page=10&limit=10"},
					{Value: ">", Href: "?page=5&limit=10"},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			t.Log(test.name)
			paginator := NewPaginator(test.input.count, test.input.limit, test.input.page)
			if len(paginator.Pages) != test.expected.length {
				t.Errorf("expected length %d, got %d", test.expected.length, len(paginator.Pages))
			}

			for i, page := range paginator.Pages {
				if !reflect.DeepEqual(page, test.expected.pages[i]) {
					t.Errorf("expected page %d to be %v, got %v", i, test.expected.pages[i], page)
				}
			}
		})
	}
}
