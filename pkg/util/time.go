package util

import "time"

func VeryLateTime() time.Time {
	return time.Date(2030, 12, 31, 23, 59, 59, 0, time.UTC)
}

func VeryEarlyTime() time.Time {
	return time.Date(2000, 12, 31, 23, 59, 59, 0, time.UTC)
}
