// THIS FILE CREATED WITH GENERATOR DO NOT EDIT!
package temporal_worker

import (
	temporal "github.com/iwrk-platform/framework/temporal"
	client "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	zap "go.uber.org/zap"
)

type TemporalWorker struct {
	Config     *Config
	Logger     *zap.Logger
	Activities []interface{}
	Workflows  []interface{}
	Temporal   client.Client
	Worker     worker.Worker
}

func NewWorker(config *Config, logger *zap.Logger, temporal *temporal.Temporal) *TemporalWorker {
	wrk := &TemporalWorker{
		Config:   config,
		Logger:   logger,
		Temporal: temporal.Client,
		Worker:   worker.New(temporal.Client, config.TaskQueue, worker.Options{}),
	}
	return wrk
}

func (s *TemporalWorker) AddActivities(activities ...interface{}) {
	s.Activities = append(s.Activities, activities...)
}

func (s *TemporalWorker) AddWorkflows(workflows ...interface{}) {
	s.Workflows = append(s.Workflows, workflows...)
}

func (s *TemporalWorker) StartWorker() error {
	for _, activity := range s.Activities {
		s.Worker.RegisterActivity(activity)
	}
	for _, workflow := range s.Workflows {
		s.Worker.RegisterWorkflow(workflow)
	}
	return s.Worker.Run(worker.InterruptCh())
}
func (s *TemporalWorker) StopWorker() error {
	s.Worker.Stop()
	return nil
}
