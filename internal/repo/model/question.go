package model

type Question struct {
	SequentialIdentifier
	UserID int64  `json:"user_id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	Tags   string `json:"tags"`
	Timestamps
}

func NewQuestion() *Question {
	return &Question{}
}

func (c *Question) Validate() []string {
	errors := make([]string, 0)
	if c.Title == "" {
		errors = append(errors, "Title cannot be empty")
	}

	if c.Body == "" {
		errors = append(errors, "Body cannot be empty")
	}

	return errors
}
