package ui

import (
	"fmt"
	"math"
)

type Page struct {
	Value    string
	Href     string
	Active   bool
	Disabled bool
}

type Paginator struct {
	Pages []Page
	Start int64
	End   int64
	Total int64
}

func NewPaginator(rowCount, rowLimit, activePage int64) *Paginator {

	pages := []Page{}

	totalPages := int64(math.Ceil(float64(rowCount) / float64(rowLimit)))

	pages = append(pages, Page{
		Value:    "<",
		Href:     fmt.Sprintf("?page=%d&limit=%d", activePage-1, rowLimit),
		Disabled: activePage <= 1,
	})

	for i := int64(1); i <= totalPages; i++ {

		if totalPages <= 4 {
			pages = append(pages, Page{
				Value:  fmt.Sprintf("%d", i),
				Href:   fmt.Sprintf("?page=%d&limit=%d", i, rowLimit),
				Active: activePage == i,
			})
			continue
		}

		if i == 1 || i == totalPages || (i <= 3 && activePage <= 3) || (i >= totalPages-2 && activePage == totalPages) || (i >= activePage-1 && i <= activePage+1) {
			pages = append(pages, Page{
				Value:  fmt.Sprintf("%d", i),
				Href:   fmt.Sprintf("?page=%d&limit=%d", i, rowLimit),
				Active: activePage == i,
			})
			continue
		}

		if len(pages) >= 1 && pages[len(pages)-1].Value != "..." {
			pages = append(pages, Page{
				Value:    "...",
				Disabled: true,
				Href:     "",
			})
		}
	}

	pages = append(pages, Page{
		Value:    ">",
		Href:     fmt.Sprintf("?page=%d&limit=%d", activePage+1, rowLimit),
		Disabled: activePage >= totalPages,
	})

	end := (activePage-1)*rowLimit + rowLimit
	if end > rowCount {
		end = rowCount
	}
	return &Paginator{
		Pages: pages,
		Start: (activePage-1)*rowLimit + 1,
		End:   end,
		Total: rowCount,
	}
}
