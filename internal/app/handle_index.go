package app

import (
	"fmt"
	"net/http"

	"goquizbox/internal/entities"
	"goquizbox/internal/repo/database"
	"goquizbox/internal/web/webutils"
	"goquizbox/pkg/logging"

	"github.com/gin-gonic/gin"
)

func (s *Server) HandleIndex() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handleIndex")
		m := getTemplateMap(c)

		filter, err := webutils.FilterFromContext(c)
		if err != nil {
			logger.Errorf("Failed to parse pagination filter for selecting questions: %v", err)
			ErrorPage(c, "could not parse pagination")
			return
		}

		db := database.NewQuestionDB(s.env.Database())
		questions, err := db.List(ctx, filter)
		if err != nil {
			ErrorPage(c, fmt.Sprintf("failed to list questions: %v", err))
			return
		}

		count, err := db.Count(ctx, filter)
		if err != nil {
			logger.Errorf("failed to count questions: %v", err)
			ErrorPage(c, "failed to count questions")
			return
		}

		m["data"] = map[string]interface{}{
			"questions":  questions,
			"pagination": entities.NewPagination(*count, filter.Page, filter.Per),
		}

		m.AddTitle("GoQuizbox - All Questions")
		c.HTML(http.StatusOK, "index", m)
	}
}
