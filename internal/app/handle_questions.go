package app

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"goquizbox/internal/entities"
	"goquizbox/internal/repo/database"
	"goquizbox/internal/repo/model"
	"goquizbox/internal/web/ctxhelper"
	"goquizbox/internal/web/webutils"
	"goquizbox/pkg/logging"

	"github.com/gin-gonic/gin"
)

type (
	questionFormData struct {
		Title string `json:"title" form:"username" binding:"required"`
		Body  string `json:"body" form:"first_name" binding:"required"`
		Tags  string `json:"tags" form:"last_name" binding:"required"`
	}
)

func (f *questionFormData) populateQuestion(m *model.Question) {
	m.Title = f.Title
	m.Body = f.Body
	m.Tags = f.Tags
	m.CreatedAt = time.Now()
}

func (s *Server) validateCreateQuestion(ctx context.Context, newQuestion *model.Question) []string {
	logger := logging.FromContext(ctx).Named("validateCreateQuestion")
	errors := newQuestion.Validate()
	if len(errors) > 0 {
		return errors
	}
	db := database.NewQuestionDB(s.env.Database())

	if err := db.Save(ctx, newQuestion); err != nil {
		logger.Errorf("failed to insert question: %v", err)
		return []string{"encountered an error creating the question"}
	}
	return []string{}
}

func (s *Server) HandleApiAddQuestion() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var form questionFormData
		err := c.ShouldBindJSON(&form)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "invalid form provided",
			})
			return
		}

		newQuestion := model.NewQuestion()
		form.populateQuestion(newQuestion)

		newQuestion.UserID = ctxhelper.UserID(ctx)

		errors := s.validateCreateQuestion(ctx, newQuestion)
		if len(errors) > 0 {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("could not create question: %v", strings.Join(errors, ",")),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    newQuestion,
		})
	}
}

func (s *Server) HandleApiListQuestions() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handleApiListQuestions")

		filter, err := webutils.FilterFromContext(c)
		if err != nil {
			logger.Errorf("Failed to parse pagination filter for selecting questions: %v", err)

			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Failed to parse pagination",
			})
			return
		}

		userID := ctxhelper.UserID(ctx)
		logger.Infof("to validate user: %v", userID)

		db := database.NewQuestionDB(s.env.Database())
		questions, err := db.List(ctx, filter)
		if err != nil {
			logger.Errorf("failed to list questions: %v", err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not list questions",
			})
			return
		}

		count, err := db.Count(ctx, filter)
		if err != nil {
			logger.Errorf("failed to count questions: %v", err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not count questions",
			})
			return
		}

		pagination := entities.NewPagination(*count, filter.Page, filter.Per)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": map[string]interface{}{
				"questions":  questions,
				"pagination": pagination,
			},
		})
	}
}

func (s *Server) HandleApiGetQuestion() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handleApiGetQuestion")

		userID := ctxhelper.UserID(ctx)
		logger.Infof("to validate user: %v", userID)

		questionIDStr := c.Param("id")
		questionID, err := strconv.ParseInt(questionIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("failed to parse 'id' param=[%v] for get question", questionIDStr),
			})
			return
		}

		db := database.NewQuestionDB(s.env.Database())
		question, err := db.GetByID(ctx, questionID)
		if err != nil {
			logger.Errorf("failed to get question by id %v: %v", questionID, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not get question",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    question,
		})
	}
}
