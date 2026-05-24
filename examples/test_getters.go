package examples

import "time"

func (t Test) T() string {
	return t.t
}
func (t Test) X() string {
	return t.x
}
func (t Test) CreatedAt() time.Time {
	return t.createdAt
}
