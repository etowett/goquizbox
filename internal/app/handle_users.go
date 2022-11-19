package app

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"goquizbox/internal/entities"
	"goquizbox/internal/repo/database"
	"goquizbox/internal/web/ctxhelper"
	"goquizbox/internal/web/webutils"
	"goquizbox/pkg/logging"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	null "gopkg.in/guregu/null.v4"
)

type (
	userUpdateFormData struct {
		Username  string `binding:"required" json:"username"`
		Email     string `binding:"required" json:"email"`
		FirstName string `binding:"required" json:"first_name"`
		LastName  string `binding:"required" json:"last_name"`
	}
)

func (s *Server) HandleShowUserProfile() func(c *gin.Context) {
	return func(c *gin.Context) {
		m := TemplateMap{}
		m.AddTitle("goquizbox - UserProfile")
		m["user"] = sessions.Default(c).Get(userSessionkey)
		c.HTML(http.StatusOK, "user_profile", m)
	}
}

func (s *Server) HandleApiListUsers() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handleApiListUsers")

		filter, err := webutils.FilterFromContext(c)
		if err != nil {
			logger.Errorf("Failed to parse pagination filter for selecting users: %v", err)

			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Failed to parse pagination",
			})
			return
		}

		userID := ctxhelper.UserID(ctx)
		logger.Infof("to validate user: %v", userID)

		db := database.NewUserDB(s.env.Database())
		users, err := db.List(ctx, filter)
		if err != nil {
			logger.Errorf("failed to list users: %v", err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not list users",
			})
			return
		}

		count, err := db.Count(ctx, filter)
		if err != nil {
			logger.Errorf("failed to count users: %v", err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not count users",
			})
			return
		}

		pagination := entities.NewPagination(*count, filter.Page, filter.Per)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": map[string]interface{}{
				"users":      users,
				"pagination": pagination,
			},
		})
	}
}

func (s *Server) HandleApiGetUser() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handleApiGetUser")

		userID := ctxhelper.UserID(ctx)
		logger.Infof("to validate user: %v", userID)

		userIDStr := c.Param("id")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("failed to parse 'id' param=[%v] for get user", userIDStr),
			})
			return
		}

		db := database.NewUserDB(s.env.Database())
		user, err := db.GetByID(ctx, userID)
		if err != nil {
			logger.Errorf("failed to get user by id %v: %v", userID, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not get user",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    user,
		})
	}
}

func (s *Server) HandleApiUpdateUser() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handleApiUpdateUser")

		loggedInUserID := ctxhelper.UserID(ctx)

		userIDStr := c.Param("id")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("failed to parse 'id' param=[%v] for add api dlr", userIDStr),
			})
			return
		}

		if userID != loggedInUserID {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "could not update that user",
			})
			return
		}

		var form userUpdateFormData
		err = c.BindJSON(&form)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "invalid request form provided",
			})
			return
		}

		db := database.NewUserDB(s.env.Database())
		user, err := db.GetByID(ctx, userID)
		if err != nil {
			logger.Errorf("failed to get user by id %v: %v", userID, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not get user",
			})
			return
		}

		if user == nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not find the given user",
			})
			return
		}

		if user.Username != form.Username {
			userSearch, err := db.GetByUsername(ctx, form.Username)
			if err != nil {
				logger.Errorf("failed to get user by username %v: %v", form.Username, err)
				c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "could not get user for validation",
				})
				return
			}
			if userSearch != nil {
				c.JSON(http.StatusConflict, map[string]interface{}{
					"success": false,
					"message": "that username is already in use",
				})
				return
			}
		}

		user.FirstName = form.FirstName
		user.LastName = form.LastName
		user.Username = form.Username
		user.Email = form.Email
		user.UpdatedAt = null.TimeFrom(time.Now())

		err = db.Save(ctx, user)
		if err != nil {
			logger.Errorf("failed to get update the user: %v", err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not update user",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    user,
		})
	}
}

func (s *Server) HandleApiDeleteUser() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handleApiDeleteUser")

		userID := ctxhelper.UserID(ctx)
		logger.Infof("to validate user %v", userID)

		userIDStr := c.Param("id")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("failed to parse 'id' param=[%v] for delete user", userIDStr),
			})
			return
		}

		db := database.NewUserDB(s.env.Database())
		user, err := db.GetByID(ctx, userID)
		if err != nil {
			logger.Errorf("failed to get user by id %v: %v", userID, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not get user",
			})
			return
		}

		if user.ID != userID {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not delete another user",
			})
		}

		err = db.Delete(ctx, userID)
		if err != nil {
			logger.Errorf("failed to delete user by id %v: %v", userID, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not delete user",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
