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
		logger := logging.FromContext(ctx).Named("handleApiAddQuestion")

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

func (s *Server) HandleApiAddQuestionAnswer() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handleApiAddQuestionAnswer")

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

		newAnswer := model.NewAnswer()
		form.populateAnswer(newAnswer)

		userID := ctxhelper.UserID(ctx)

		if newAnswer.UserID != userID {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "You do not have permission to ask",
			})
			return
		}

		db := database.NewAnswerDB(s.env.Database())
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
				"message": fmt.Sprintf("failed to parse 'id' param=[%v]", questionIDStr),
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

func (s *Server) HandleApiGetQuestionAnswers() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handleApiGetQuestionAnswers")

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

		answerDB := database.NewAnswerDB(s.env.Database())
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

func (s *Server) HandleGetQuestion() func(c *gin.Context) {
	return func(c *gin.Context) {
		m := getTemplateMap(c)
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handlegetQuestion")

		questionIDStr := c.Param("id")
		questionID, err := strconv.ParseInt(questionIDStr, 10, 64)
		if err != nil {
			ErrorPage(c, fmt.Sprintf("failed to parse 'id' param=[%v]", questionIDStr))
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

		votesDB := database.NewVoteDB(s.env.Database())
		upVotes, err := votesDB.CountVotes(ctx, questionID, "question", "up")
		if err != nil {
			logger.Errorf("failed to count question upvotes: %v", err)
			ErrorPage(c, "could not count question upvotes")
			return
		}

		downVotes, err := votesDB.CountVotes(ctx, questionID, "question", "down")
		if err != nil {
			logger.Errorf("failed to count question downvotes: %v", err)
			ErrorPage(c, "could not count question downvotes")
			return
		}

		logger.Infow("answers", answers)
		m["data"] = map[string]interface{}{
			"question":   question,
			"upvotes":    upVotes,
			"downvotes":  downVotes,
			"answers":    answers,
			"pagination": entities.NewPagination(*count, filter.Page, filter.Per),
		}
		m.AddTitle("GoQuizbox - Question Details")
		c.HTML(http.StatusOK, "question", m)
	}
}

func (s *Server) HandleQuestionVote() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handleQuestionVote")

		questionIDStr := c.Param("id")
		questionID, err := strconv.ParseInt(questionIDStr, 10, 64)
		if err != nil {
			ErrorPage(c, fmt.Sprintf("failed to parse 'id' param=[%v]", questionIDStr))
			return
		}

		action := c.Query("act")
		kind := c.Query("kind")

		session := sessions.Default(c)
		user := session.Get(userSessionkey).(*model.User)

		db := database.NewVoteDB(s.env.Database())
		vote, err := db.ByUserAndKind(ctx, user.ID, questionID, kind)
		if err != nil {
			logger.Errorf("failed to get vote: %v", err)
			ErrorPage(c, "could not get vote")
			return
		}

		if vote != nil {
			if vote.Mode == action {
				logger.Infof("already voted %v for it: %v", action, vote)
				c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/questions/%v", questionID))
			}
			vote.Mode = action
		} else {
			vote = &model.Vote{
				UserID: user.ID,
				KindID: questionID,
				Kind:   kind,
				Mode:   action,
			}
		}

		err = db.Save(ctx, vote)
		if err != nil {
			logger.Errorf("failed to save vote: %v", err)
			ErrorPage(c, "could not save vote")
			return
		}

		c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/questions/%v", questionID))
	}
}
