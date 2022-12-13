package app

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"goquizbox/internal/entities"
	"goquizbox/internal/repo/database"
	"goquizbox/internal/repo/model"
	"goquizbox/internal/web/ctxhelper"
	"goquizbox/internal/web/webutils"
	"goquizbox/pkg/logging"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type (
	questionFormData struct {
		UserID string `json:"user_id" form:"user_id" binding:"required"`
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

func (f *questionFormData) populateQuestion(m *model.Question) {
	m.Title = f.Title
	m.Body = f.Body
	m.Tags = f.Tags
}

func (f *answerFormData) populateAnswer(m *model.Answer) {
	m.UserID = f.UserID
	m.QuestionID = f.QuestionID
	m.Body = f.Body
}

func (s *Server) HandleAskQuestionShow() func(c *gin.Context) {
	return func(c *gin.Context) {
		m := getTemplateMap(c)
		m.AddTitle("GoQuizbox - Ask Question")
		c.HTML(http.StatusOK, "question_ask", m)
	}
}

func (s *Server) HandleAskQuestionProcess() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		m := getTemplateMap(c)

		var form questionFormData
		err := c.ShouldBind(&form)
		if err != nil {
			ErrorPage(c, err.Error())
			return
		}

		newQuiz := model.NewQuestion()
		form.populateQuestion(newQuiz)

		session := sessions.Default(c)
		user := session.Get(userSessionkey).(*model.User)

		newQuiz.UserID = user.ID

		errors := s.validateCreateQuestion(ctx, newQuiz)
		if len(errors) > 0 {
			m.AddErrors(errors...)
			c.HTML(http.StatusOK, "question_ask", m)
			return
		}

		c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/questions/%v", newQuiz.ID))
	}
}

func (s *Server) validateCreateQuestion(
	ctx context.Context,
	newQuestion *model.Question,
) []string {
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

func (s *Server) HandleAnswerQuestion() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handleAnswerQuestion")
		m := getTemplateMap(c)

		var form answerFormData
		err := c.ShouldBind(&form)
		if err != nil {
			m.AddErrors(err.Error())
			c.HTML(http.StatusOK, "question", m)
			return
		}

		newAnswer := model.NewAnswer()
		form.populateAnswer(newAnswer)

		session := sessions.Default(c)
		user := session.Get(userSessionkey).(*model.User)

		if newAnswer.UserID != user.ID {
			m.AddErrors("You do not have permission to ask")
			c.HTML(http.StatusOK, "question", m)
			return
		}

		db := database.NewAnswerDB(s.env.Database())
		if err := db.Save(ctx, newAnswer); err != nil {
			logger.Errorf("failed to save answer: %v", err)
			m.AddErrors("encountered an error creating the answer")
			c.HTML(http.StatusOK, "question", m)
			return
		}

		c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/questions/%v", newAnswer.QuestionID))
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

func (s *Server) HandleGetQuestion() func(c *gin.Context) {
	return func(c *gin.Context) {
		m := getTemplateMap(c)
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handlegetQuestion")

		questionIDStr := c.Param("id")
		questionID, err := strconv.ParseInt(questionIDStr, 10, 64)
		if err != nil {
			ErrorPage(c, fmt.Sprintf("failed to parse 'id' param=[%v] for get question", questionIDStr))
			return
		}

		filter, err := webutils.FilterFromContext(c)
		if err != nil {
			logger.Errorf("Failed to parse pagination filter for selecting answer list: %v", err)
			ErrorPage(c, "Failed to parse pagination")
			return
		}

		db := database.NewQuestionDB(s.env.Database())
		question, err := db.ByID(ctx, questionID)
		if err != nil {
			logger.Errorf("failed to get question: %v", err)
			ErrorPage(c, "could not get question")
			return
		}

		answerDB := database.NewAnswerDB(s.env.Database())
		answers, err := answerDB.ByQuestion(ctx, questionID, filter)
		if err != nil {
			logger.Errorf("failed to get answers for question: %v", err)
			ErrorPage(c, "failed to get answer list")
			return
		}

		count, err := answerDB.CountByQuestion(ctx, questionID, filter)
		if err != nil {
			logger.Errorf("failed to count answers: %v", err)
			ErrorPage(c, "could not count answers")
			return
		}

		logger.Infow("answers", answers)
		m["data"] = map[string]interface{}{
			"question":   question,
			"answers":    answers,
			"pagination": entities.NewPagination(*count, filter.Page, filter.Per),
		}
		m.AddTitle("GoQuizbox - Question Details")
		c.HTML(http.StatusOK, "question", m)
	}
}
