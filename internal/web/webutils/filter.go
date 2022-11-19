package webutils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func FilterFromContext(
	c *gin.Context,
) (*Filter, error) {

	filter := &Filter{}

	page, per, err := paginationFromContext(c)
	if err != nil {
		return filter, err
	}

	filter.Page = page
	filter.Per = per
	filter.Term = strings.TrimSpace(c.Query("term"))

	deleted := strings.TrimSpace(c.Query("deleted"))
	if deleted != "" {
		isDeleted, err := strconv.ParseBool(deleted)
		if err != nil {
			return filter, fmt.Errorf("failed to parse 'deleted' query param")
		}

		filter.Deleted.SetValid(isDeleted)
	}

	return filter, nil
}

func ApiKeyFilterFromContext(
	c *gin.Context,
) (*ApiKeyFilter, error) {

	filter := &ApiKeyFilter{}

	filter.Name = strings.TrimSpace(c.Query("name"))

	return filter, nil
}

func MessageFilterFromContext(
	c *gin.Context,
) (*MessageFilter, error) {

	filter, err := FilterFromContext(c)
	if err != nil {
		return &MessageFilter{}, err
	}

	messageFilter := &MessageFilter{
		SendType:  strings.TrimSpace(c.Query("send_type")),
		SMSSource: strings.TrimSpace(c.Query("source")),
		Filter:    *filter,
	}

	return messageFilter, nil
}

func paginationFromContext(
	c *gin.Context,
) (int, int, error) {

	page := 1
	per := 20

	var err error

	pageQuery := strings.TrimSpace(c.Query("page"))
	if pageQuery != "" {
		page, err = strconv.Atoi(pageQuery)
		if err != nil {
			return page, per, fmt.Errorf("invalid page query given [%v]", pageQuery)
		}
	}

	perQuery := strings.TrimSpace(c.Query("per"))
	if perQuery != "" {
		per, err = strconv.Atoi(perQuery)
		if err != nil {
			return page, per, fmt.Errorf("invalid per query given [%v]", perQuery)
		}
	}

	return page, per, nil
}

func OrderFilterFromContext(
	c *gin.Context,
) (*OrderFilter, error) {

	filter := &OrderFilter{}
	filter.Field = strings.TrimSpace(c.Query("order_by"))
	filter.Order = strings.TrimSpace(c.Query("order"))

	return filter, nil
}
