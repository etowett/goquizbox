package app

import (
	"net/http"

	"goquizbox/pkg/logging"

	"github.com/gin-gonic/gin"
)

// Controller is the interfactor for controllers that can be pluggied into Gin
// for the admin console portion of this project.
type Controller interface {
	Execute(g *gin.Context)
}

func ErrorPage(c *gin.Context, messages ...string) {
	logger := logging.FromContext(c)
	logger.Errorf("error: %v", messages)
	c.HTML(http.StatusInternalServerError, "error", gin.H{"error": messages})
	c.Abort()
}
