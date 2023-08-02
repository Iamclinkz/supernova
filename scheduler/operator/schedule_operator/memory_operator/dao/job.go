package dao

import "supernova/scheduler/model"

type Job struct {
	model.Job
}

func FromModelJob(mJob *model.Job) *Job {
	return &Job{
		Job: *mJob,
	}
}

func (j *Job) ToModelJob() *model.Job {
	ret := new(model.Job)
	*ret = j.Job
	return ret
}
