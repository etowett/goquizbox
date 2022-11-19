package app

type TemplateMap map[string]interface{}

func (t TemplateMap) AddTitle(title string) {
	t["title"] = title
}

func (t TemplateMap) AddErrors(errors ...string) {
	t["error"] = errors
}

func (t TemplateMap) AddSuccess(success ...string) {
	t["success"] = success
}
