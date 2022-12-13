package model

type Answer struct {
	SequentialIdentifier
	UserID     int64  `json:"user_id"`
	QuestionID int64  `json:"question_id"`
	Body       string `json:"body"`
	Timestamps
}

func NewAnswer() *Answer {
	return &Answer{}
}

func (c *Answer) Validate() []string {
	errors := make([]string, 0)
	if c.Body == "" {
		errors = append(errors, "Body cannot be empty")
	}
	return errors
}
