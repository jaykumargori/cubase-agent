package state

import "time"

type Connection struct {
	OutputPort        string
	Connected         bool
	ReadbackAvailable bool
	CheckedAt         time.Time
}
