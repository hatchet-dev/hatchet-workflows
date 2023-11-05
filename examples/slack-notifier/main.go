package main

import (
	"context"
	_ "embed"
	"os"
	"strings"
	"time"

	"github.com/hatchet-dev/hatchet-workflows/pkg/dispatcher"
	"github.com/hatchet-dev/hatchet-workflows/pkg/integrations/slack"
	"github.com/hatchet-dev/hatchet-workflows/pkg/worker"
	"github.com/hatchet-dev/hatchet-workflows/pkg/workflows/types"
)

//go:embed .hatchet/slack-channel.yaml
var SlackChannelWorkflow []byte

func init() {
	// initialize the slack channel workflow with SLACK_USER_ID
	slackUserId := os.Getenv("SLACK_USER_ID")

	if slackUserId == "" {
		panic("SLACK_USER_ID environment variable must be set")
	}

	slackFileWithReplacedEnv := strings.Replace(string(SlackChannelWorkflow), "$SLACK_USER_ID", slackUserId, 1)

	SlackChannelWorkflow = []byte(slackFileWithReplacedEnv)
}

func main() {
	// read the slack workflow
	slackWorkflowFile, err := types.ParseYAML(context.Background(), SlackChannelWorkflow)

	if err != nil {
		panic(err)
	}

	// render the slack workflow using the environment variable SLACK_USER_ID
	slackToken := os.Getenv("SLACK_TOKEN")
	slackTeamId := os.Getenv("SLACK_TEAM_ID")

	if slackToken == "" {
		panic("SLACK_TOKEN environment variable must be set")
	}

	if slackTeamId == "" {
		panic("SLACK_TEAM_ID environment variable must be set")
	}

	slackInt := slack.NewSlackIntegration(slackToken, slackTeamId, true)

	// create a worker
	worker, err := worker.NewWorker(
		worker.WithWorkflowFiles([]*types.WorkflowFile{
			&slackWorkflowFile,
		}),
		worker.WithIntegrations(
			slackInt,
		),
	)

	if err != nil {
		panic(err)
	}

	err = worker.Start()

	if err != nil {
		panic(err)
	}

	d := dispatcher.NewDispatcher(
		dispatcher.WithWorkflowFiles(
			[]*types.WorkflowFile{
				&slackWorkflowFile,
			},
		),
	)

	err = d.Trigger("user:create", map[string]any{
		"username": "testing12345",
	})

	if err != nil {
		panic(err)
	}

	// wait for workflows to complete
	time.Sleep(10 * time.Second)
}
