package system

import "time"

type UTCClock struct{}

func NewUTCClock() UTCClock {
	return UTCClock{}
}

func (UTCClock) Now() time.Time {
	return time.Now().UTC()
}
