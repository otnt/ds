package clockService

import (
	"fmt"
)

// Logical Clock Timestamp is essentially just an 64-bit integer.
// For more information on Logical Clock, check logicalClock.go
type LogicalTimestamp struct {
	Time int64
}

// Compare current logical timestamp with another logical timestamp,
// this function will panic if input parameter is not logical timestamp
func (lt *LogicalTimestamp) CompareTo(timestamp Timestamp) int {
	lt2 := timestamp.(*LogicalTimestamp)

	if lt.Time < lt2.Time {
		return -1
	} else if lt.Time > lt2.Time {
		return 1
	} else {
		return 0
	}
}

func (lt *LogicalTimestamp) String() string {
	return fmt.Sprintf("%d", lt.Time)
}
