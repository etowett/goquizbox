package entities

type SequentialIdentifier struct {
	ID int64 `json:"id"`
}

func (si SequentialIdentifier) IsNew() bool {
	return si.ID == 0
}
