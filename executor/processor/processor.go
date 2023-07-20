package processor

import (
	"supernova/pkg/api"
)

type JobProcessor interface {
	Process(job *api.Job) *api.JobResult
	GetGlueType() string
}
