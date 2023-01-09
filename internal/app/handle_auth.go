package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"goquizbox/internal/entities"
	"goquizbox/internal/logger"
	"goquizbox/internal/repos"
	"goquizbox/internal/util"
	"goquizbox/internal/web/auth"
	"goquizbox/internal/web/ctxhelper"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	null "gopkg.in/guregu/null.v4"
)

const (
	userSessionkey = "goquizbox:user-session"
)

type (
	registerFormData struct {
		FirstName       string `json:"first_name" form:"first_name" binding:"required"`
		LastName        string `json:"last_name" form:"last_name" binding:"required"`
		Email           string `json:"email" form:"email" binding:"required"`
		Password        string `json:"password" form:"password" binding:"required"`
		PasswordConfirm string `json:"password_confirm" form:"password_confirm" binding:"required"`
	}

	loginFormData struct {
		Email    string `form:"email" json:"email" binding:"required"`
		Password string `form:"password" json:"password" binding:"required"`
		Remember bool   `form:"remember"`
	}
)

func (s *Server) HandleLogin(sessionAuthenticator auth.SessionAuthenticator) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var form loginFormData
		err := c.ShouldBindJSON(&form)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "invalid form provided",
			})
			return
		}

		db := repos.NewUserDB(s.env.Database())
		theUser, err := db.ByEmail(ctx, form.Email)
		if err != nil {
			logger.Errorf("get user by email failed: %v", err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "encountered error searching user by email",
			})
			return
		}

		if theUser == nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "invalid email provided",
			})
			return
		}

		err = util.MatchPassword(theUser.PasswordHash, form.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "invalid password provided",
			})
			return
		}

		if theUser.Status.IsUnverified() {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "user is unverified",
			})
			return
		}

		if !theUser.Status.IsActive() {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "user is inactive",
			})
			return
		}

		theSession := &entities.Session{
			IPAddress:       ctxhelper.IPAddress(ctx),
			LastRefreshedAt: time.Now(),
			ExpiresAt:       null.TimeFrom(time.Now().Add(time.Hour * time.Duration(3))),
			UserAgent:       ctxhelper.UserAgent(ctx),
			UserID:          theUser.ID,
		}

		sessionDB := repos.NewSessionDB(s.env.Database())
		if err := sessionDB.Save(ctx, theSession); err != nil {
			logger.Errorf("failed to insert user session: %v", err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "session could not be saved",
			})
			return
		}

		tokenValue, err := sessionAuthenticator.SetUserSessionInResponse(c.Writer, theUser, theSession)
		if err != nil {
			logger.Errorf("failed to set session in response: %v", err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "session could not set in response",
			})
			return
		}

		c.SetCookie("auth-token", tokenValue, 60*60, "/", "localhost", false, true)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"user":    theUser,
			"token":   tokenValue,
		})
	}
}

func (s *Server) HandleRegister() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var form registerFormData
		err := c.ShouldBindJSON(&form)
		if err != nil {
			logger.Errorf("failed to bind form: %v", err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "invalid form provided",
			})
			return
		}

		newUser := &entities.User{
			Email:           strings.ToLower(form.Email),
			FirstName:       cases.Title(language.English, cases.Compact).String(form.FirstName),
			LastName:        cases.Title(language.English, cases.Compact).String(form.LastName),
			Status:          "inactive",
			EmailVerified:   false,
			Password:        form.Password,
			PasswordConfirm: form.PasswordConfirm,
			PasswordHash:    util.GeneratePasswordHash(form.Password),
		}

		errors := newUser.Validate()
		if len(errors) > 0 {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("could not register user: %v", strings.Join(errors, ",")),
			})
			return
		}

		db := repos.NewUserDB(s.env.Database())
		theUser, err := db.ByEmail(ctx, newUser.Email)
		if err != nil {
			logger.Errorf("get user by email failed: %v", err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "encountered error searching user by email",
			})
			return
		}

		if theUser != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "that email is already registered",
			})
			return
		}

		if err := db.Save(ctx, newUser); err != nil {
			logger.Errorf("failed to insert user when registering: %v", err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "encountered an error registering user",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "check your email for activation code",
			"data":    newUser,
		})
	}
}

func (s *Server) HandleApiLogoutUser() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		tokenInfo := ctxhelper.TokenInfo(ctx)

		db := repos.NewSessionDB(s.env.Database())
		session, err := db.GetSessionByID(ctx, tokenInfo.SessionID)
		if err != nil {
			logger.Errorf("failed to get session by id %v, %v", tokenInfo.SessionID, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "an internal error occured fetching session",
			})
			return
		}

		if session.UserID != tokenInfo.UserID {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("cannot logout session=[%v] by user=[%v]", tokenInfo.SessionID, tokenInfo.UserID),
			})
			return
		}

		if session.DeactivatedAt.Valid {
			c.JSON(http.StatusOK, gin.H{"success": true})
		}

		session.DeactivatedAt = null.TimeFrom(time.Now())

		err = db.Save(ctx, session)

		if err != nil {
			logger.Errorf("failed to deactivate session id %v, %v", tokenInfo.SessionID, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("cannot deactivate session=[%v]", tokenInfo.SessionID),
			})
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
