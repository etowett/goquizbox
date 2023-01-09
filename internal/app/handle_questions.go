package app

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"goquizbox/internal/entities"
	"goquizbox/internal/logger"
	"goquizbox/internal/repos"
	"goquizbox/internal/web/ctxhelper"
	"goquizbox/internal/web/webutils"

	"github.com/gin-gonic/gin"
)

type (
	questionFormData struct {
		UserID int64  `json:"user_id" form:"user_id" binding:"required"`
		Title  string `json:"title" form:"title" binding:"required"`
		Body   string `json:"body" form:"body" binding:"required"`
		Tags   string `json:"tags" form:"tags" binding:"required"`
	}

	answerFormData struct {
		UserID     int64  `json:"user_id" form:"user_id" binding:"required"`
		QuestionID int64  `json:"question_id" form:"question_id" binding:"required"`
		Body       string `json:"body" form:"body" binding:"required"`
	}
)

func (f *questionFormData) populateQuestion(m *entities.Question) {
	m.Title = f.Title
	m.Body = f.Body
	m.Tags = f.Tags
}

func (f *answerFormData) populateAnswer(m *entities.Answer) {
	m.UserID = f.UserID
	m.QuestionID = f.QuestionID
	m.Body = f.Body
}

func (s *Server) validateCreateQuestion(
	ctx context.Context,
	newQuestion *entities.Question,
) []string {
	errors := newQuestion.Validate()
	if len(errors) > 0 {
		return errors
	}
	db := repos.NewQuestionDB(s.env.Database())
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
			logger.Errorf("failed to bind form: %v", err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "invalid form provided",
			})
			return
		}

		newQuestion := entities.NewQuestion()
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

func (s *Server) HandleApiAddQuestionAnswer() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var form answerFormData
		err := c.ShouldBind(&form)
		if err != nil {
			logger.Errorf("failed to bind form: %v", err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "invalid form provided",
			})
			return
		}

		newAnswer := entities.NewAnswer()
		form.populateAnswer(newAnswer)

		userID := ctxhelper.UserID(ctx)

		if newAnswer.UserID != userID {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "You do not have permission to ask",
			})
			return
		}

		db := repos.NewAnswerDB(s.env.Database())
		if err := db.Save(ctx, newAnswer); err != nil {
			logger.Errorf("failed to save answer: %v", err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "encountered an error creating the answer",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    newAnswer,
		})
	}
}

func (s *Server) HandleListQuestions() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		filter, err := webutils.FilterFromContext(c)
		if err != nil {
			logger.Errorf("Failed to parse pagination filter for selecting questions: %v", err)

			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Failed to parse pagination",
			})
			return
		}

		db := repos.NewQuestionDB(s.env.Database())
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

		questionIDStr := c.Param("id")
		questionID, err := strconv.ParseInt(questionIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("failed to parse 'id' param=[%v]", questionIDStr),
			})
			return
		}

		db := repos.NewQuestionDB(s.env.Database())
		question, err := db.ByID(ctx, questionID)
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

func (s *Server) HandleApiGetQuestionAnswers() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		questionIDStr := c.Param("id")
		questionID, err := strconv.ParseInt(questionIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("failed to parse 'id' param=[%v]", questionIDStr),
			})
			return
		}

		filter, err := webutils.FilterFromContext(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to parse pagination",
			})
			return
		}

		answerDB := repos.NewAnswerDB(s.env.Database())
		answers, err := answerDB.ByQuestion(ctx, questionID, filter)
		if err != nil {
			logger.Errorf("failed to get answers for question: %v", err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not get answers",
			})
			return
		}

		count, err := answerDB.CountByQuestion(ctx, questionID, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not count answers",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": map[string]interface{}{
				"answers":    answers,
				"pagination": entities.NewPagination(*count, filter.Page, filter.Per),
			},
		})
	}
}
