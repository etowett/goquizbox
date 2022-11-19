package webutils

import (
	"errors"
	"fmt"
	"goquizbox/internal/util"

	null "gopkg.in/guregu/null.v4"
)

type Filter struct {
	Page     int
	Per      int
	From     string
	To       string
	Term     string
	FromTime null.Time
	ToTime   null.Time
	UserID   null.Int
	Deleted  null.Bool
}

func (f *Filter) ConvertTime() error {
	if f.From == "" || f.To == "" {
		return errors.New("time_filter: from or to filter time is empty")
	}

	fromTime, err := util.ParseTime(f.From)
	if err != nil {
		return fmt.Errorf("time_filter: parse from time [%v], err [%v]", f.From, err)
	}

	f.FromTime = null.TimeFrom(fromTime)

	toTime, err := util.ParseTime(f.To)
	if err != nil {
		return fmt.Errorf("time_filter: parse to time [%v], err [%v]", f.To, err)
	}

	f.ToTime = null.TimeFrom(toTime)

	return nil
}

func (f *Filter) NoPagination() *Filter {
	return &Filter{
		From:     f.From,
		To:       f.To,
		Term:     f.Term,
		FromTime: f.FromTime,
		ToTime:   f.ToTime,
		UserID:   f.UserID,
		Deleted:  f.Deleted,
	}
}

func (f *Filter) TimeFilterSet() bool {

	return f.From != "" && f.To != ""
}

func (f *Filter) ExportLimit() {

	f.Page = 1
	f.Per = 1000
}

type ApiKeyFilter struct {
	Name string
}

type MessageFilter struct {
	SendType  string
	SMSSource string
	Filter
}

func (f *MessageFilter) NoPagination() *MessageFilter {
	return &MessageFilter{
		SendType:  f.SendType,
		SMSSource: f.SMSSource,
		Filter: Filter{
			From:     f.From,
			To:       f.To,
			Term:     f.Term,
			FromTime: f.FromTime,
			ToTime:   f.ToTime,
			UserID:   f.UserID,
			Deleted:  f.Deleted,
		},
	}
}

type KeywordFilter struct {
	Keyword     string
	Page        int
	Per         int
	ShortCodeID null.Int
}

type OrderFilter struct {
	Field string
	Order string
}
