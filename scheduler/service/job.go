package service

import (
	"context"
	"errors"
	"fmt"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
)

type JobService struct {
	scheduleOperator  schedule_operator.Operator
	statisticsService *StatisticsService
}

func NewJobService(jobOperator schedule_operator.Operator, statisticsService *StatisticsService) *JobService {
	return &JobService{
		scheduleOperator:  jobOperator,
		statisticsService: statisticsService,
	}
}

func (s *JobService) FetchJobFromID(ctx context.Context, jobID uint) (*model.Job, error) {
	job, err := s.scheduleOperator.FetchJobFromID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *JobService) AddJob(ctx context.Context, job *model.Job) error {
	if err := s.ValidateJob(job); err != nil {
		return err
	}

	if err := s.scheduleOperator.InsertJob(ctx, job); err != nil {
		return err
	}

	return nil
}

func (s *JobService) AddJobs(ctx context.Context, jobs []*model.Job) error {
	for _, job := range jobs {
		if err := s.ValidateJob(job); err != nil {
			return fmt.Errorf("job:%+v error:%v", job, err)
		}
	}

	if err := s.scheduleOperator.InsertJobs(ctx, jobs); err != nil {
		return err
	}

	return nil
}

func (s *JobService) FindJobByName(ctx context.Context, name string) (*model.Job, error) {
	return s.scheduleOperator.FindJobByName(ctx, name)
}

func (s *JobService) DeleteJob(ctx context.Context, jobID uint) error {
	if err := s.scheduleOperator.DeleteJobFromID(ctx, jobID); err != nil {
		return err
	}

	return nil
}

func (s *JobService) ValidateJob(job *model.Job) error {
	// if job.Name == "" {
	// 	return errors.New("name cannot be empty")
	// }

	if !job.ExecutorRouteStrategy.Valid() {
		return errors.New("invalid ExecutorRouteStrategy")
	}

	if job.GlueType == "" {
		return errors.New("invalid GlueType")
	}

	if job.GlueSource == nil {
		return errors.New("glue_source cannot be empty")
	}

	return nil
}
