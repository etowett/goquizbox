package entities

type Vote struct {
	SequentialIdentifier
	UserID int64  `json:"user_id"`
	KindID int64  `json:"kind_id"`
	Kind   string `json:"kind"`
	Mode   string `json:"mode"`
	Timestamps
}

func NewVote() *Vote {
	return &Vote{}
}

func (c *Vote) Validate() []string {
	errors := make([]string, 0)
	if c.UserID < 1 {
		errors = append(errors, "UserID cannot be empty")
	}
	return errors
}
