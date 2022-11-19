package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"net/http"
)

const userSessionkey = "goquizbox:user-session"

func MustBeLoggedIn(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userSessionkey)

	// ctx := c.Request.Context()
	// logger := logging.FromContext(ctx).Named("mustBeLoggedIn")

	// logger.Infof("Path: %+v, User session key: %v", c.Request.URL.Path, user)

	if user == nil {
		c.Redirect(http.StatusMovedPermanently, "/login")
		c.Abort()
		return
	}
	c.Next()
}

func MustNotBeLoggedIn(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userSessionkey)

	// ctx := c.Request.Context()
	// logger := logging.FromContext(ctx).Named("mustBeLoggedIn")

	// logger.Infof("Path: %+v, User session key: %v", c.Request.URL.Path, user)

	if user != nil {
		c.Redirect(http.StatusMovedPermanently, "/dashboard")
		c.Abort()
		return
	}
	c.Next()
}
