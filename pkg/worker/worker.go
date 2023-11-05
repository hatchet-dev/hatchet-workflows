package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/hatchet-dev/hatchet-workflows/internal/config/loader"
	"github.com/hatchet-dev/hatchet-workflows/internal/datautils"
	hatchetclient "github.com/hatchet-dev/hatchet-workflows/pkg/client"
	"github.com/hatchet-dev/hatchet-workflows/pkg/integrations"
	"github.com/hatchet-dev/hatchet-workflows/pkg/workflows/fileutils"
	"github.com/hatchet-dev/hatchet-workflows/pkg/workflows/types"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// Worker is a wrapper around the temporal worker
type Worker worker.Worker

type activityFunc func(ctx context.Context, input any) (result any, err error)

type activities map[string]activityFunc

type workerOptions struct {
	*worker.Options

	queueName string

	activities activities

	filesLoader  func() []*types.WorkflowFile
	clientLoader func(queueName string) client.Client
}

func defaultWorkerOptions() *workerOptions {
	clientLoader := func(queueName string) client.Client {
		configLoader := &loader.ConfigLoader{}

		hatchetClient, err := configLoader.LoadTemporalClient()

		if err != nil {
			panic(err)
		}

		tc, err := hatchetClient.GetClient(queueName)

		if err != nil {
			panic(err)
		}

		return tc
	}

	return &workerOptions{
		Options:      &worker.Options{},
		queueName:    hatchetclient.HatchetDefaultQueueName,
		activities:   make(activities),
		clientLoader: clientLoader,
		filesLoader:  fileutils.DefaultLoader,
	}
}

type workerOptFunc func(*workerOptions)

// WithTemporalClient sets the temporal client to use for the worker. Note that the default queue registered with the Temporal client
// will be used if jobs and steps do not specify a queue.
func WithTemporalClient(tc client.Client) workerOptFunc {
	return func(opts *workerOptions) {
		opts.clientLoader = func(queueName string) client.Client {
			return tc
		}
	}
}

// WithWorkflowFiles sets the workflow files to use for the worker. If this is not passed in, the workflows files will be loaded
// from the .hatchet folder in the current directory.
func WithWorkflowFiles(files []*types.WorkflowFile) workerOptFunc {
	return func(opts *workerOptions) {
		opts.filesLoader = func() []*types.WorkflowFile {
			return files
		}
	}
}

// WithQueueName sets the queue name to use for the worker. Note that this will override the queue name set in the default Temporal client,
// but will not override the queue name set in the Temporal client passed in with [WithTemporalClient].
func WithQueueName(queueName string) workerOptFunc {
	return func(opts *workerOptions) {
		opts.queueName = queueName
	}
}

// WithIntegrations registers all integrations with the worker. See [integrations.Integration] to see the interface
// integrations must satisfy.
func WithIntegrations(ints ...integrations.Integration) workerOptFunc {
	return func(opts *workerOptions) {
		for _, i := range ints {
			intCp := i

			// get list of actions
			for _, action := range i.Actions() {
				actionCp := action
				fmt.Println("registering action", intCp.GetId()+":"+actionCp)

				opts.activities[intCp.GetId()+":"+actionCp] = func(ctx context.Context, input any) (result any, err error) {
					return intCp.PerformAction(types.Action{
						IntegrationID: intCp.GetId(),
						Verb:          actionCp,
					}, input.(map[string]any))
				}
			}
		}
	}
}

// NewWorker creates a new worker from opts.
func NewWorker(opts ...workerOptFunc) (Worker, error) {
	workerOptions := defaultWorkerOptions()

	for _, opt := range opts {
		opt(workerOptions)
	}

	tc := workerOptions.clientLoader(workerOptions.queueName)

	workerInstance := worker.New(tc, workerOptions.queueName, *workerOptions.Options)

	workflowFiles := workerOptions.filesLoader()

	// register all workflow with the worker
	for _, workflowFile := range workflowFiles {
		for jobName, job := range workflowFile.Jobs {
			jobCp := job

			temporalWorkflow := func(ctx workflow.Context, input any) (result []byte, err error) {
				retrypolicy := &temporal.RetryPolicy{
					MaximumAttempts: 1,
				}

				options := workflow.ActivityOptions{
					ScheduleToCloseTimeout: 10 * time.Minute,
					StartToCloseTimeout:    10 * time.Minute,
					RetryPolicy:            retrypolicy,
				}

				sharedInput := map[string]any{
					"steps": map[string]any{},
				}

				activityCtx := workflow.WithActivityOptions(ctx, options)

				for _, step := range jobCp.Steps {
					var activityRes any

					globalInput, err := datautils.ToJSONMap(input)

					if err != nil {
						return nil, err
					}

					inputMaps := []map[string]any{
						globalInput,
						sharedInput,
					}

					activityInput := map[string]any{}

					// if the "With" map is not nil, it was set by the user
					if step.With != nil {
						activityDataInput := datautils.MergeMaps(inputMaps...)

						withData := step.With

						datautils.RenderTemplateFields(activityDataInput, withData)

						activityInput = datautils.MergeMaps(activityDataInput, withData)
					}

					action, err := types.ParseActionID(step.ActionID)

					if err != nil {
						return nil, err
					}

					integrationVerb := action.IntegrationVerbString()

					err = workflow.ExecuteActivity(activityCtx, integrationVerb, activityInput).Get(activityCtx, &activityRes)

					if err != nil {
						// TODO: call any recovery activities
						return nil, err
					}

					// set the output in shared data
					sharedInput["steps"].(map[string]any)[step.ID] = map[string]any{
						"outputs": activityRes,
					}
				}

				return nil, nil
			}

			workerInstance.RegisterWorkflowWithOptions(temporalWorkflow, workflow.RegisterOptions{
				Name: jobName,
			})

			// register all activities for the job
			registeredActivities := make(map[string]bool)

			for _, step := range job.Steps {
				action, err := types.ParseActionID(step.ActionID)

				if err != nil {
					return nil, err
				}

				integrationVerb := action.IntegrationVerbString()

				// make sure activity is registered
				activityFunction, alreadyRegistered := workerOptions.activities[integrationVerb]

				if !alreadyRegistered {
					return nil, fmt.Errorf("activity %s (%s) is not registered", step.Name, integrationVerb)
				}

				if _, exists := registeredActivities[integrationVerb]; !exists {
					workerInstance.RegisterActivityWithOptions(activityFunction, activity.RegisterOptions{
						Name: integrationVerb,
					})
				}

				registeredActivities[integrationVerb] = true
			}
		}
	}

	return workerInstance, nil
}
