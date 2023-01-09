package app

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"goquizbox/internal/entities"
	"goquizbox/internal/logger"
	"goquizbox/internal/repos"
	"goquizbox/internal/web/ctxhelper"
	"goquizbox/internal/web/webutils"

	"github.com/gin-gonic/gin"
	null "gopkg.in/guregu/null.v4"
)

type (
	userUpdateFormData struct {
		Email     string `binding:"required" json:"email"`
		FirstName string `binding:"required" json:"first_name"`
		LastName  string `binding:"required" json:"last_name"`
	}
)

func (s *Server) HandleListUsers() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		filter, err := webutils.FilterFromContext(c)
		if err != nil {
			logger.Errorf("Failed to parse pagination filter for selecting users: %v", err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Failed to parse pagination",
			})
			return
		}

		db := repos.NewUserDB(s.env.Database())
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

func (s *Server) HandleGetUser() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		userIDStr := c.Param("id")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("failed to parse 'id' param=[%v] for get user", userIDStr),
			})
			return
		}

		db := repos.NewUserDB(s.env.Database())
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

		db := repos.NewUserDB(s.env.Database())
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

		user.FirstName = form.FirstName
		user.LastName = form.LastName
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

		userIDStr := c.Param("id")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("failed to parse 'id' param=[%v] for delete user", userIDStr),
			})
			return
		}

		if userID != ctxhelper.UserID(ctx) {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "could not delete another user",
			})
		}

		db := repos.NewUserDB(s.env.Database())
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
