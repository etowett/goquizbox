package util

type placeholder struct {
	count int
}

func NewPlaceholder() *placeholder {
	return &placeholder{
		count: 1,
	}
}

func (u *placeholder) Touch() int {
	count := u.count
	u.count++
	return count
}
