package service

import (
	"errors"
	"supernova/executor/processor"
)

type ProcessorService struct {
	processors map[string]processor.JobProcessor
}

func NewProcessorService() *ProcessorService {
	return &ProcessorService{processors: make(map[string]processor.JobProcessor)}
}

func (s *ProcessorService) Register(p processor.JobProcessor) error {
	if p.GetGlueType() == "" {
		return errors.New("empty processor glue type")
	}

	if s.processors[p.GetGlueType()] != nil {
		return errors.New("duplicate register")
	}

	s.processors[p.GetGlueType()] = p
	return nil
}

func (s *ProcessorService) GetRegister(glueType string) processor.JobProcessor {
	return s.processors[glueType]
}
