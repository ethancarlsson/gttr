package examples

import "time"

//go:generate gttr --type Test
type Test struct {
	t         string
	x         string
	createdAt time.Time
}
