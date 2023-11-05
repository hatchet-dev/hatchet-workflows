package dispatcher

import (
	"context"

	"go.temporal.io/sdk/client"

	"github.com/hashicorp/go-multierror"

	"github.com/hatchet-dev/hatchet-workflows/internal/config/loader"
	hatchetclient "github.com/hatchet-dev/hatchet-workflows/pkg/client"
	"github.com/hatchet-dev/hatchet-workflows/pkg/workflows/fileutils"
	"github.com/hatchet-dev/hatchet-workflows/pkg/workflows/types"
)

type Dispatcher struct {
	c     *hatchetclient.Client
	files []*types.WorkflowFile
}

type DispatchOpts struct {
	clientLoader func() *hatchetclient.Client
	filesLoader  func() []*types.WorkflowFile
}

type DispatchOptsFunc func(d *DispatchOpts)

func defaultDispatchOpts() *DispatchOpts {
	clientLoader := func() *hatchetclient.Client {
		configLoader := &loader.ConfigLoader{}

		hatchetClient, err := configLoader.LoadTemporalClient()

		if err != nil {
			panic(err)
		}

		return hatchetClient
	}

	return &DispatchOpts{clientLoader, fileutils.DefaultLoader}
}

func WithHatchetClient(hc *hatchetclient.Client) DispatchOptsFunc {
	return func(opts *DispatchOpts) {
		opts.clientLoader = func() *hatchetclient.Client {
			return hc
		}
	}
}

func WithWorkflowFiles(files []*types.WorkflowFile) DispatchOptsFunc {
	return func(opts *DispatchOpts) {
		opts.filesLoader = func() []*types.WorkflowFile {
			return files
		}
	}
}

type DispatcherInterface interface {
	Trigger(eventId string, data any) error
}

func NewDispatcher(
	opts ...DispatchOptsFunc,
) DispatcherInterface {
	dispatchOpts := defaultDispatchOpts()

	for _, opt := range opts {
		opt(dispatchOpts)
	}

	d := &Dispatcher{
		c:     dispatchOpts.clientLoader(),
		files: dispatchOpts.filesLoader(),
	}

	return d
}

func (d *Dispatcher) InitSchedules() error {
	var allErrs error

	for _, file := range d.files {
		if file.On.Cron.Schedule != "" {
			err := d.dispatchAllScheduledJobs(file.On.Cron.Schedule, file.Jobs, nil)

			if err != nil {
				allErrs = multierror.Append(allErrs, err)
			}

			allErrs = multierror.Append(allErrs, err)
		}
	}

	return allErrs
}

func (d *Dispatcher) Trigger(eventId string, data any) error {
	// find all the workflows triggered from this event id
	var allErrs error

	for _, file := range d.files {
		fileCp := file

		for _, event := range fileCp.On.Events {
			if event == eventId {
				err := d.dispatchAllJobs(fileCp.Jobs, data)

				if err != nil {
					allErrs = multierror.Append(allErrs, err)
				}
			}
		}
	}

	return allErrs
}

func (d *Dispatcher) dispatchAllJobs(jobs map[string]types.WorkflowJob, data any) error {
	var allErrs error

	for jobName, job := range jobs {
		jobCp := job
		err := d.dispatchJob(data, jobName, jobCp)

		if err != nil {
			allErrs = multierror.Append(allErrs, err)
		}
	}

	return allErrs
}

func (d *Dispatcher) dispatchJob(data any, jobName string, job types.WorkflowJob) error {
	tc, err := d.c.GetClient(job.Queue)

	if err != nil {
		return err
	}

	taskQueue := job.Queue

	if taskQueue == "" {
		taskQueue = d.c.GetDefaultQueueName()
	}

	startOpts := client.StartWorkflowOptions{
		ID:        jobName,
		TaskQueue: taskQueue,
	}

	_, err = tc.ExecuteWorkflow(
		context.Background(),
		startOpts,
		jobName,
		data,
	)

	if err != nil {
		return err
	}

	return nil
}

func (d *Dispatcher) dispatchAllScheduledJobs(inputSchedule string, jobs map[string]types.WorkflowJob, data any) error {
	schedule, skipUpdateSchedule := parseScheduleInput(inputSchedule)

	var allErrs error

	for jobName, job := range jobs {
		jobCp := job

		err := d.dispatchScheduledJob(schedule, skipUpdateSchedule, data, jobName, jobCp)

		if err != nil {
			allErrs = multierror.Append(allErrs, err)
		}
	}

	return allErrs
}

func (d *Dispatcher) dispatchScheduledJob(schedule string, skipUpdateSchedule bool, data any, jobName string, job types.WorkflowJob) error {
	tc, err := d.c.GetClient(job.Queue)
	if err != nil {
		return err
	}

	action := &client.ScheduleWorkflowAction{
		TaskQueue: job.Queue,
		Workflow:  jobName,
		Args:      []interface{}{data},
	}

	// determine if schedule exists
	scheduleHandle := tc.ScheduleClient().GetHandle(context.Background(), jobName)

	if scheduleHandle.GetID() == "" {
		_, err = tc.ScheduleClient().Create(
			context.Background(),
			client.ScheduleOptions{
				ID: jobName,
				Spec: client.ScheduleSpec{
					CronExpressions: []string{schedule},
				},
				Action: action,
			},
		)
	} else {
		err := scheduleHandle.Update(
			context.Background(),
			client.ScheduleUpdateOptions{
				DoUpdate: func(input client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
					if !skipUpdateSchedule {
						input.Description.Schedule.Spec = &client.ScheduleSpec{
							CronExpressions: []string{schedule},
						}
					}

					input.Description.Schedule.Action = action

					return &client.ScheduleUpdate{
						Schedule: &input.Description.Schedule,
					}, nil
				},
			},
		)
		if err != nil {
			return err
		}
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}
