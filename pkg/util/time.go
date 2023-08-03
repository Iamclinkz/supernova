package util

import (
	"math/rand"
	"time"
)

func VeryLateTime() time.Time {
	return time.Date(2030, 12, 31, 23, 59, 59, 0, time.UTC)
}

func VeryEarlyTime() time.Time {
	return time.Date(2000, 12, 31, 23, 59, 59, 0, time.UTC)
}

func TimeRandBetween(min, max time.Duration) time.Duration {
	if min > max {
		panic("")
	}
	return min + time.Duration(rand.Int63n(int64(max-min)))
}

func GetRandomOffsetByNum(count int) time.Duration {
	if count == 0 {
		return time.Millisecond
	}
	var maxOffsetMicroSec int64 = int64(count) * 200
	return time.Duration(rand.Int63n(maxOffsetMicroSec)) * time.Microsecond
}
