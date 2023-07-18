package service

import (
	"context"
	"errors"
	"fmt"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
	"time"
)

type JobService struct {
	jobOperator       schedule_operator.Operator
	statisticsService *StatisticsService
}

func NewJobService(jobOperator schedule_operator.Operator, statisticsService *StatisticsService) *JobService {
	return &JobService{
		jobOperator:       jobOperator,
		statisticsService: statisticsService,
	}
}

func (s *JobService) FetchJobFromID(ctx context.Context, jobID uint) (*model.Job, error) {
	job, err := s.jobOperator.FetchJobFromID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *JobService) AddJob(ctx context.Context, job *model.Job) error {
	if err := s.ValidateJob(job); err != nil {
		return err
	}

	if err := s.jobOperator.InsertJob(ctx, job); err != nil {
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

	if err := s.jobOperator.InsertJobs(ctx, jobs); err != nil {
		return err
	}

	return nil
}

func (s *JobService) FindJobByName(ctx context.Context, name string) (*model.Job, error) {
	return s.jobOperator.FindJobByName(ctx, name)
}

func (s *JobService) DeleteJob(ctx context.Context, jobID uint) error {
	if err := s.jobOperator.DeleteJobFromID(ctx, jobID); err != nil {
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

	if !job.GlueType.Valid() {
		return errors.New("invalid GlueType")
	}

	if job.GlueSource == "" {
		return errors.New("glue_source cannot be empty")
	}

	if job.ExecutorExecuteTimeout < time.Millisecond*20 {
		return errors.New("executor_execute_timeout must be greater than 20ms")
	}

	return nil
}
