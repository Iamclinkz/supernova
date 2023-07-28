package service

import (
	"context"
	"errors"
	"fmt"
	"supernova/pkg/discovery"
	"supernova/pkg/util"
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

	s.insertGlueTag(job)

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
		s.insertGlueTag(job)
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

func (s *JobService) insertGlueTag(job *model.Job) {
	//检查用户指定的glueType是否已经加到了Tag中，作为executor筛选的条件之一
	glueTag := discovery.GlueTypeTagPrefix + job.GlueType
	userTagsSlice := util.DecodeTags(job.Tags)
	found := false
	for _, tag := range userTagsSlice {
		if tag == glueTag {
			found = true
		}
	}

	if !found {
		userTagsSlice = append(userTagsSlice, glueTag)
		job.Tags = util.EncodeTag(userTagsSlice)
	}
}
