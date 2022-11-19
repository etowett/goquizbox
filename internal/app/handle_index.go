package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) HandleIndex() func(c *gin.Context) {
	return func(c *gin.Context) {
		m := TemplateMap{}
		m.AddTitle("GoQuizbox - Home")
		c.HTML(http.StatusOK, "index", m)
	}
}
