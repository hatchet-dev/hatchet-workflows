Runs workflow: [sample-workflow.yaml](./.hatchet/sample-workflow.yaml)

## Explanation

This folder contains a demo example of a workflow that simply echoes the input message as an output. The workflow file showcases the following features:

- Running a simple job with a set of dependent steps
- Variable references within step arguments -- each subsequent step in a workflow can call `.steps.<step_id>.outputs` to access output arguments

While the `main.go` file showcases the following features:

- Creating an integration called `EchoIntegration` which simply echoes the input message and returns it as an output
- Initializing a worker via `worker.NewWorker` and registering the EchoIntegration on it
- Creating a dispatcher

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

You should see output resembling the following:

```sh
2023/11/04 15:57:08 INFO  No logger configured for temporal client. Created default one.
2023/11/04 15:57:08 INFO  No logger configured for temporal client. Created default one.
2023/11/04 15:57:08 INFO  Started Worker Namespace default TaskQueue default WorkerID PID@MBP@default
2023/11/04 15:57:08 DEBUG ExecuteActivity Namespace default TaskQueue default WorkerID PID@MBP@default WorkflowType print-user WorkflowID print-user RunID run_id Attempt 1 ActivityID 5 ActivityType echo:echo
Username is echo-test
2023/11/04 15:57:08 DEBUG ExecuteActivity Namespace default TaskQueue default WorkerID PID@MBP@default WorkflowType print-user WorkflowID print-user RunID run_id Attempt 1 ActivityID 11 ActivityType echo:echo
Above message is: Username is echo-test
2023/11/04 15:57:08 DEBUG ExecuteActivity Namespace default TaskQueue default WorkerID PID@MBP@default WorkflowType print-user WorkflowID print-user RunID run_id Attempt 1 ActivityID 17 ActivityType echo:echo
Above message is: Above message is: Username is echo-test
```

4. Within the Temporal UI, you should see a single workflow called `print-user` (which you'll notice is the name of the `job` in [sample-workflow.yaml](./sample-workflow.yaml)). You should also see a workflow execution consisting of three `echo:echo` activities (which you'll notice is the name of the `actionId` in [sample-workflow.yaml](./sample-workflow.yaml)):

![image](https://imagedelivery.net/hKvo7fgKu6IoDvMLV830jw/de513bc2-a8e6-431d-d660-e13561e46100/large)
