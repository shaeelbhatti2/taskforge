package scheduler

import (
	"time"
)

func NextInterval(from time.Time, intervalSec int, tz string) time.Time {
	loc := time.UTC
	if tz != "" {
		if l, err := time.LoadLocation(tz); err == nil {
			loc = l
		}
	}
	base := from.In(loc)
	return base.Add(time.Duration(intervalSec) * time.Second)
}

func ApplyMissedPolicy(lastRun, due time.Time, policy string) bool {
	switch policy {
	case "skip":
		return false
	case "coalesce":
		return due.Sub(lastRun) > time.Hour
	default:
		return true
	}
}
