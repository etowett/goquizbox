package app

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type TemplateMap map[string]interface{}

func getTemplateMap(c *gin.Context) TemplateMap {
	session := sessions.Default(c)
	user := session.Get(userSessionkey)
	return TemplateMap{
		"user": user,
	}
}

func (t TemplateMap) AddTitle(title string) {
	t["title"] = title
}

func (t TemplateMap) AddErrors(errors ...string) {
	t["error"] = errors
}

func (t TemplateMap) AddSuccess(success ...string) {
	t["success"] = success
}
