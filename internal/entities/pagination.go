package entities

import (
	null "gopkg.in/guregu/null.v4"
)

type Pagination struct {
	Count    int      `json:"count"`
	NextPage null.Int `json:"next_page"`
	NumPages int      `json:"num_pages"`
	Page     int      `json:"page"`
	Per      int      `json:"per"`
	PrevPage null.Int `json:"prev_page"`
}

func NewPagination(count, page, per int) *Pagination {
	var prevPage, nextPage null.Int

	if page > 1 {
		prevPage = null.IntFrom(int64(page - 1))
	}

	if per < 1 {
		per = 10
	}

	numPages := count / per
	if count == 0 {
		numPages = 1
	} else if count%per != 0 {
		numPages++
	}

	if page < numPages {
		nextPage = null.IntFrom(int64(page + 1))
	}

	return &Pagination{
		Count:    count,
		NextPage: nextPage,
		NumPages: numPages,
		Page:     page,
		Per:      per,
		PrevPage: prevPage,
	}
}
