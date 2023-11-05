package types

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v2"
)

type WorkflowFile struct {
	Name string `yaml:"name"`

	On WorkflowOn `yaml:"on"`

	Jobs map[string]WorkflowJob `yaml:"jobs"`
}

func (w *WorkflowFile) GetJobByName(name string) *WorkflowJob {
	for jobName, job := range w.Jobs {
		if jobName == name {
			return &job
		}
	}

	return nil
}

func (w *WorkflowFile) ListJobNames() []string {
	res := []string{}

	for jobName := range w.Jobs {
		res = append(res, jobName)
	}

	return res
}

type WorkflowOn struct {
	Events []string       `yaml:"events"`
	Cron   WorkflowOnCron `yaml:"cron"`
}

type RandomScheduleOpt string

const (
	Random15Min  RandomScheduleOpt = "random_15_min"
	RandomHourly RandomScheduleOpt = "random_hourly"
	RandomDaily  RandomScheduleOpt = "random_daily"
)

type WorkflowOnCron struct {
	Schedule string `yaml:"schedule"`
}

type WorkflowEvent struct {
	Name string `yaml:"name"`
}

type WorkflowJob struct {
	Queue string `yaml:"queue"`

	Timeout string `yaml:"timeout"`

	Steps []WorkflowStep `yaml:"steps"`
}

type WorkflowStep struct {
	Name     string                 `yaml:"name"`
	ID       string                 `yaml:"id"`
	ActionID string                 `yaml:"actionId"`
	Timeout  string                 `yaml:"timeout"`
	With     map[string]interface{} `yaml:"with,omitempty"`
}

func ParseYAML(ctx context.Context, yamlBytes []byte) (WorkflowFile, error) {
	var workflowFile WorkflowFile

	if yamlBytes == nil {
		return workflowFile, fmt.Errorf("workflow yaml input is nil")
	}

	err := yaml.Unmarshal(yamlBytes, &workflowFile)
	if err != nil {
		return workflowFile, fmt.Errorf("error unmarshaling workflow yaml: %w", err)
	}

	return workflowFile, nil
}
