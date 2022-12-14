package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"goquizbox/internal/repo/database"
	"goquizbox/internal/repo/model"
	"goquizbox/internal/util"
	"goquizbox/internal/web/auth"
	"goquizbox/internal/web/ctxhelper"
	"goquizbox/pkg/logging"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	null "gopkg.in/guregu/null.v4"
)

const (
	userSessionkey = "goquizbox:user-session"
)

type (
	registerFormData struct {
		FirstName            string `json:"first_name" form:"first_name" binding:"required"`
		LastName             string `json:"last_name" form:"last_name" binding:"required"`
		Email                string `json:"email" form:"email" binding:"required"`
		Password             string `json:"password" form:"password" binding:"required"`
		PasswordConfirmation string `json:"password_confirmation" form:"password_confirmation" binding:"required"`
	}

	loginFormData struct {
		Email    string `form:"email" json:"email" binding:"required"`
		Password string `form:"password" json:"password" binding:"required"`
		Remember string `form:"remember"`
	}
)

func (f *registerFormData) PopulateUser(a *model.User) {
	a.FirstName = f.FirstName
	a.LastName = f.LastName
	a.Email = f.Email
	a.Password = f.Password
	a.Status = model.UserStatusActive
	a.PasswordConfirmation = f.PasswordConfirmation
	a.PasswordHash = util.GeneratePasswordHash(f.Password)
	a.CreatedAt = time.Now()
}

func (f *loginFormData) PopulateLogin(a *model.Login) {
	a.Email = strings.TrimSpace(f.Email)
	a.Password = strings.TrimSpace(f.Password)
	a.Remember = f.Remember == "on"
}

func (s *Server) validateUserLogin(
	ctx context.Context,
	newlogin *model.Login,
) ([]string, *model.User, *model.Session) {
	logger := logging.FromContext(ctx).Named("validateUserLogin")
	var theUser *model.User
	var dbSession *model.Session

	errors := newlogin.Validate()
	if len(errors) > 0 {
		return errors, theUser, dbSession
	}

	db := database.NewUserDB(s.env.Database())
	theUser, err := db.ByEmail(ctx, newlogin.Email)
	if err != nil {
		logger.Errorf("get user by email failed: %v", err)
		return []string{"encountered error searching user by email"}, theUser, dbSession
	}

	if theUser == nil {
		return []string{"that email could not be found"}, theUser, dbSession
	}

	err = util.MatchPassword(theUser.PasswordHash, newlogin.Password)
	if err != nil {
		return []string{"invalid password provided"}, theUser, dbSession
	}

	if theUser.Status.IsUnverified() {
		return []string{"user is unverified"}, theUser, dbSession
	}

	if !theUser.Status.IsActive() {
		return []string{"user is inactive"}, theUser, dbSession
	}

	dbSession = &model.Session{
		IPAddress:       ctxhelper.IPAddress(ctx),
		LastRefreshedAt: time.Now(),
		ExpiresAt:       null.TimeFrom(time.Now().Add(time.Hour * time.Duration(3))),
		UserAgent:       ctxhelper.UserAgent(ctx),
		UserID:          theUser.ID,
	}

	sessionDB := database.NewSessionDB(s.env.Database())
	if err := sessionDB.Save(ctx, dbSession); err != nil {
		logger.Errorf("failed to insert user session: %v", err)
		return []string{"could not save session in db"}, theUser, dbSession
	}
	return []string{}, theUser, dbSession
}

func (s *Server) HandleLoginShow() func(c *gin.Context) {
	return func(c *gin.Context) {
		m := getTemplateMap(c)
		m.AddTitle("GoQuizbox - Login")
		c.HTML(http.StatusOK, "login", m)
	}
}

func (s *Server) HandleLoginProcess() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		m := getTemplateMap(c)
		logger := logging.FromContext(ctx).Named("handleLoginProcess")

		var form loginFormData
		err := c.ShouldBind(&form)
		if err != nil {
			logger.Errorf("invalid login form: %v", err)
			ErrorPage(c, "invalid login form provided")
			return
		}

		newlogin := model.NewLogin()
		form.PopulateLogin(newlogin)

		errors, theUser, _ := s.validateUserLogin(ctx, newlogin)
		if len(errors) > 0 {
			m.AddErrors(errors...)
			c.HTML(http.StatusOK, "login", m)
			return
		}

		session := sessions.Default(c)
		session.Set(userSessionkey, theUser)
		if err := session.Save(); err != nil {
			logger.Errorf("session save error: %v", err)
			m.AddErrors("failed to save session")
			c.HTML(http.StatusInternalServerError, "login", m)
			return
		}

		c.Redirect(http.StatusMovedPermanently, "/")
	}
}

func (s *Server) HandleAPILogin(sessionAuthenticator auth.SessionAuthenticator) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var form loginFormData
		err := c.BindJSON(&form)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "invalid form provided",
			})
			return
		}

		newlogin := model.NewLogin()
		form.PopulateLogin(newlogin)

		errors, theUser, theSession := s.validateUserLogin(ctx, newlogin)
		if len(errors) > 0 {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("could not login: %v", strings.Join(errors, ",")),
			})
			return
		}

		tokenValue, _ := sessionAuthenticator.SetUserSessionInResponse(c.Writer, theUser, theSession)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"user":    theUser,
			"token":   tokenValue,
		})
	}
}

func (s *Server) HandleLogout() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handleLogout")

		// TODO: Deactivate session in DB
		// tokenInfo := ctxhelper.TokenInfo(ctx)

		// db := database.NewSessionDB(s.env.Database())
		// session, err := db.GetSessionByID(ctx, tokenInfo.SessionID)
		// if err != nil {
		// 	logger.Errorf("failed to get session by id %v, %v", tokenInfo.SessionID, err)
		// 	c.JSON(http.StatusInternalServerError, map[string]interface{}{
		// 		"success": false,
		// 		"message": "an internal error occured fetching session",
		// 	})
		// 	return
		// }

		// if session.UserID != tokenInfo.UserID {
		// 	c.JSON(http.StatusBadRequest, map[string]interface{}{
		// 		"success": false,
		// 		"message": fmt.Sprintf("cannot logout session=[%v] by user=[%v]", tokenInfo.SessionID, tokenInfo.UserID),
		// 	})
		// 	return
		// }

		// if session.DeactivatedAt.Valid {
		// 	c.JSON(http.StatusOK, gin.H{"success": true})
		// }

		// session.DeactivatedAt = null.TimeFrom(time.Now())

		// err = db.Save(ctx, session)

		session := sessions.Default(c)
		session.Delete(userSessionkey)
		if err := session.Save(); err != nil {
			logger.Infof("Failed to delete session:", err)
			c.Redirect(http.StatusMovedPermanently, "/")
		}

		c.Redirect(http.StatusMovedPermanently, "/")
	}
}

func (s *Server) HandleRegisterShow() func(c *gin.Context) {
	return func(c *gin.Context) {
		m := getTemplateMap(c)
		m.AddTitle("GoQuizbox - Register")
		c.HTML(http.StatusOK, "register", m)
	}
}

func (s *Server) validateUserRegistration(ctx context.Context, newUser *model.User) []string {
	logger := logging.FromContext(ctx).Named("validateUserRegistration")
	errors := newUser.Validate()
	if len(errors) > 0 {
		return errors
	}
	db := database.NewUserDB(s.env.Database())
	theUser, err := db.ByEmail(ctx, newUser.Email)
	if err != nil {
		logger.Errorf("get user by email failed: %v", err)
		return []string{"encountered error searching user by email"}
	}

	if theUser != nil {
		return []string{"that email is already registered"}
	}

	if err := db.Save(ctx, newUser); err != nil {
		logger.Errorf("failed to insert user when registering: %v", err)
		return []string{"encountered an error registering user"}
	}
	return []string{}
}

func (s *Server) HandleRegisterProcess() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		m := getTemplateMap(c)

		var form registerFormData
		err := c.ShouldBind(&form)
		if err != nil {
			ErrorPage(c, err.Error())
			return
		}

		newUser := model.NewUser()
		form.PopulateUser(newUser)

		errors := s.validateUserRegistration(ctx, newUser)
		if len(errors) > 0 {
			m.AddErrors(errors...)
			c.HTML(http.StatusOK, "register", m)
			return
		}

		c.Redirect(http.StatusMovedPermanently, "/login")
	}
}

func (s *Server) HandleAPIRegister() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var form registerFormData
		err := c.ShouldBindJSON(&form)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "invalid form provided",
			})
			return
		}

		newUser := model.NewUser()
		form.PopulateUser(newUser)

		errors := s.validateUserRegistration(ctx, newUser)
		if len(errors) > 0 {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("could not register user: %v", strings.Join(errors, ",")),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    newUser,
		})
	}
}

func (s *Server) HandleApiLogoutUser() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("handleApiLogoutUser")

		tokenInfo := ctxhelper.TokenInfo(ctx)

		db := database.NewSessionDB(s.env.Database())
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
