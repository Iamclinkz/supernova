package util

import "time"

const veryLongTime = 253402300799

func VeryLongTime() time.Time {
	return time.Unix(veryLongTime, 0)
}
