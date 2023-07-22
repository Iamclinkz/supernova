package config

import "time"

const (
	MaxHandleJobTimeout = 3 * time.Second
	JobRetryInterval    = 5 * time.Second
)
