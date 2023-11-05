package main

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/hatchet-dev/hatchet-workflows/pkg/dispatcher"
	"github.com/hatchet-dev/hatchet-workflows/pkg/worker"
	"github.com/hatchet-dev/hatchet-workflows/pkg/workflows/types"
)

func main() {
	// Create a worker. This automatically reads in a TemporalClient from .env and workflow files from the .hatchet
	// directory, but this can be customized with the `worker.WithTemporalClient` and `worker.WithWorkflowFiles` options.
	worker, err := worker.NewWorker(
		worker.WithIntegrations(
			&EchoIntegration{},
		),
	)

	if err != nil {
		panic(err)
	}

	// Normally, you would start this worker in blocking fashion using worker.Run, but for this example we'll be triggering the
	// worker below, so we start it in non-blocking fashion.
	err = worker.Start()

	if err != nil {
		panic(err)
	}

	// It's good practice to call worker.Stop(), as this deregisters the worker from Temporal.
	defer worker.Stop()

	// Create a dispatcher. This automatically reads in a TemporalClient from .env and workflow files from the .hatchet
	// directory, but this can be customized with the `dispatcher.WithTemporalClient` and `dispatcher.WithWorkflowFiles` options.
	d := dispatcher.NewDispatcher()

	// Trigger a new event. This will trigger any workflows which listen to the `user:create` event.
	err = d.Trigger("user:create", map[string]any{
		"username": "echo-test",
	})

	if err != nil {
		panic(err)
	}

	// wait for workflows to complete
	time.Sleep(10 * time.Second)
}

// EchoIntegration simply prints the message it receives and stores it in the `messages` field.
// It also provides the message as an output.
type EchoIntegration struct {
	messages []string
}

func (e *EchoIntegration) GetId() string {
	return "echo"
}

func (e *EchoIntegration) Actions() []string {
	return []string{"echo"}
}

func (e *EchoIntegration) PerformAction(action types.Action, input map[string]any) (result map[string]any, err error) {
	switch action.Verb {
	case "echo":
		e.messages = append(e.messages, input["message"].(string))

		fmt.Println(input["message"])

		return map[string]any{
			"message": input["message"],
		}, nil
	default:
		return nil, fmt.Errorf("unsupported action: %s", action.Verb)
	}
}
