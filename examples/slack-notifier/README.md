## Simple Workflow Example

This example runs the [slack-channel.yaml](./.hatchet/slack-channel.yaml).

## Explanation

This folder contains a demo example of a workflow that creates a Slack channel, adds a default user to that Slack channel, and send an initial message to the channel. The workflow file showcases the following features:

- Running a simple job with a set of dependent steps
- Variable references within step arguments -- each subsequent step in a workflow can call `.steps.<step_id>.outputs` to access output arguments

While the `main.go` file showcases the following features:

- Using an existing integration called `SlackIntegration` which provides several actions to perform
- Providing a custom workflow file (as the workflow file needs to be populated with an env var `$SLACK_USER_ID`)

## How to run

Navigate to this directory and run the following steps:

1. Make sure you have a Temporal server running (see the instructions [here](../../README.md)).
2. Set your environment variables -- if you're using the bundled Temporal server, this will look like:

```sh
cat > .env <<EOF
TEMPORAL_CLIENT_TLS_ROOT_CA_FILE=../../hack/dev/certs/ca.cert
TEMPORAL_CLIENT_TLS_CERT_FILE=../../hack/dev/certs/client-worker.pem
TEMPORAL_CLIENT_TLS_KEY_FILE=../../hack/dev/certs/client-worker.key
TEMPORAL_CLIENT_TLS_SERVER_NAME=cluster

SLACK_USER_ID=<TODO>
SLACK_TOKEN=<TODO>
SLACK_TEAM_ID=<TODO>
EOF
```

3. Run the following within this directory:

```sh
/bin/bash -c '
set -a
. .env
set +a

go run main.go';
```
