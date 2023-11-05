[![slack](https://img.shields.io/badge/Join%20Our%20Community-Slack-blue)](https://join.slack.com/t/hatchet-co/signup) [![License: MIT](https://img.shields.io/badge/License-MIT-purple.svg)](https://opensource.org/licenses/MIT) [![Go Reference](https://pkg.go.dev/badge/github.com/hatchet-dev/hatchet-workflows.svg)](https://pkg.go.dev/github.com/hatchet-dev/hatchet-workflows)

## Introduction

_**Note:** Hatchet workflows are in early development. Changes are not guaranteed to be backwards-compatible. If you'd like to run them in production, feel free to reach out on Slack for tips._

Hatchet is a declarative workflow builder for Golang apps. Using Hatchet, you can create workers which process a set of background tasks based on different triggers, using a declarative file syntax that draws heavy inspiration from Github actions. Unlike Github actions, code runs inside your Go application, with triggers and actions that you've defined.

As a simple example, let's say you want to perform 3 actions when a user has signed up for your app:

1. Initialize a set of resources for the user (perhaps a sandbox environment for testing).
2. Send the user an automated greeting over email
3. Add the user to a newsletter campaign

With Hatchet workflows, this would look something like the following:

```yaml
name: "Post User Sign Up"
on:
  - user:create
jobs:
  create-resources:
    steps:
      - name: Create sandbox environment
        id: createSandbox
        actionId: sandbox:create
        timeout: 60s
  greet-user:
    steps:
      - name: Greet user
        id: greetUser
        actionId: postmark:email-from-template
        timeout: 15s
        with:
          firstName: "{{ .user.firstName }}"
          email: "{{ .user.email }}"
  add-to-newsletter:
    steps:
      - name: Add to newsletter
        id: addUserToNewsletter
        actionId: newsletter:add-user
        timeout: 15s
        with:
          email: "{{ .user.email }}"
```

In your codebase, you would then create the following integrations (see [Writing an integration](#writing-an-integration)):

- A `sandbox` integration responsible for creating/tearing down a sandbox environment
- A `postmark` integration for sending an email from a template
- A `newsletter` integration for adding a user to a newsletter campaign

Ultimately, the goal of Hatchet workflows are that you don't need to write these integrations yourself -- creating a robust set of prebuilt integrations is one of the goals of the project.

### Why is this useful?

- No need to build all of your plumbing logic (action 1 -> event 1 -> action 2 -> event 2). Just define your jobs and steps and write your business logic. This is particularly useful the more complex your workflows become.
- Using prebuilt integrations with a standard interface makes building auxiliary services like notification systems, billing, backups, and auditing much easier. **Please file an issue if you'd like to see an integration supported.** The following are on the roadmap:
  - Email providers: Sendgrid, Postmark, AWS SES
  - Stripe
  - AWS S3
- Additionally, if you're already familiar with/using Temporal, making workflows declarative provides several benefits:
  - Makes spec'ing, debugging and visualizing workflows much simpler
  - Automatically updates triggers, schedules, and timeouts when they change, rather than doing this through a UI/CLI/SDK
  - Makes monitoring easier to build by logically separating units of work - jobs will automatically correspond to `BeginSpan`. OpenTelemetry support is on the roadmap.

## Getting Started

For a set of end-to-end examples, see the [examples](./examples) directory.

### Prerequisites

- Go 1.21 installed
- Taskfile installed (instructions [here](https://taskfile.dev/installation/))

### Setting up Temporal

First, you need to get a Temporal cluster running. There are many ways to do this: see [here](https://docs.temporal.io/kb/all-the-ways-to-run-a-cluster) for all options.

To make things easier, there's a bundled server and UI in `./cmd/temporal-server`. Run the following commands to get it working:

```sh
task write-default-env
task generate-certs
task start-temporal-server
```

You can then navigate to 127.0.0.1:8233 to view the Temporal UI.

### Writing an Integration

An integration needs to satisfy the following interface:

```go
type Integration interface {
	GetId() string
	Actions() []string
	PerformAction(action types.Action, data map[string]interface{}) (map[string]interface{}, error)
}
```

See the [Slack integration](./pkg/integrations/slack) for an example.

### Writing a Workflow

By default, Hatchet searches for workflows in the `.hatchet` folder relative to the directory you run your application in. However, you can configure this using `worker.WithWorkflowFiles` and the exported `fileutils` package (`fileutils.ReadAllValidFilesInDir`).

There are two main sections of a workflow file:

**Triggers (using `on`)**

This section specifies what triggers a workflow. This can be events or a crontab-like schedule. For example, the following are valid triggers:

```yaml
on:
  - eventkey1
  - eventkey2
```

```yaml
on:
  cron:
    schedule: "*/15 * * * *"
```

There are also a set of keywords `random_15_min`, `random_hourly`, `random_daily` for cron-like schedules. Upon creation of these schedules, a random minute is picked in the given interval - for example, `random_hourly` might result in a schedule `49 * * * *` (the 49th minute of every hour). After creation, these schedules will **not** be updated with a new random schedule.

```yaml
on:
  cron:
    schedule: "random_hourly"
```

The point of this is to avoid burstiness if all jobs have the exact same schedule (i.e. runs at the 0th minute of every hour), you may start to run out of memory on your workers.

**Jobs**

After defining your triggers, you define a list of jobs to run based on the triggers. **Jobs run in parallel.** Jobs contain the following fields:

```yaml
# ...
jobs:
  my-awesome-job:
    # (optional) A queue name
    queue: internal
    # (optional) A timeout value for the entire job
    timeout: 60s
    # (required) A set of steps for the job; see below
    steps: []
```

Within each job, there are a set of **steps** which run sequentially. A step can contain the following fields:

```yaml
# (required) the name of the step
name: Step 1
# (required) a unique id for the step (can be referenced by future steps)
id: step-1
# (required) the action id in the form of "integration_id:action".
actionId: "slack:create-channel"
# (required) the timeout of the individual step
timeout: 15s
# (optional or required, depending on integration) input data to the integration
with:
  key: val
```

### Creating a Worker

Workers can be created using:

```go
import "github.com/hatchet-dev/hatchet-workflows/pkg/worker"

func main() {
  // ... application code
  worker, err := worker.NewWorker(
    worker.WithIntegrations(
      myIntegration,
    ),
  )

  if err != nil {
    // TODO: error handling here
    panic(err)
  }

  // Start worker in non-blocking fashion
  worker.Start()

  // Start worker in blocking fashion
  worker.Run()
}
```

You can configure the worker with your own set of workflow files using the `worker.WithWorkflowFiles` option.

### Triggering Events

To trigger events from your main application, use the `dispatcher` package:

```go
import "github.com/hatchet-dev/hatchet-workflows/pkg/worker"

func main() {
  d := dispatcher.NewDispatcher()

  // this would typically be called within a handler
	err = d.Trigger("user:create", map[string]any{
		"username": "testing12345",
	})

	if err != nil {
		panic(err)
	}
}
```

You can configure the dispatcher with your own set of workflow files using the `dispatcher.WithWorkflowFiles` option.

## Why should I care?

**If you're unfamiliar with background task processing**

Many Go APIs start out without a task processing/worker service. You might not need it, but at a certain level of complexity, you probably will. There are a few use-cases where workers start to make sense:

1. You need to run scheduled tasks which that aren't triggered from your core API. For example, this may be a daily cleanup task, like traversing soft-deleted database entries or backing up data to S3.
2. You need to run tasks which are triggered by API events, but aren't required for the core business logic of the handler. For example, you want to add a user to your CRM after they sign up.

For both of these cases, it's typical to re-use a lot of core functionality from your API, so the most natural place to start is by adding some automation within your API itself; for example, after returning `201 Created`, you might send a greeting to the user, initialize a sandbox environment, send an internal notification that a user signed up, etc, all within your API handlers. Let's say you've handled this case as following:

```go
// Hypothetical handler called via a routing package, let's just pretend it returns an error
func MyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
    // Boilerplate code to parse the request
    var newUser User
    err := json.NewDecoder(r.Body).Decode(&newUser)
    if err != nil {
      http.Error(w, "Invalid user data", http.StatusBadRequest)
      return err
    }

    // Validate email and password fields...
    // (Add your validation logic here)

    // Create a user in the database
    user, err := createUser(ctx, newUser.Email, newUser.Password)
    if err != nil {
      // Handle database errors, such as unique constraint violation
      http.Error(w, "Error creating user", http.StatusInternalServerError)
      return err
    }

    // Return 201 created with user type
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)

    // Send user a greeting
    err := email.SendGreetingEmail(context.Background(), user)

    if err != nil {
      // can't return an error, since header is already set
      fmt.Println(err)
    }

    // ... other post-signup operations
}
```

At some point, you realize all of these background operations don't really belong in the handler -- when they're part of the handler, they're more difficult to monitor and observe, difficult to retry (especially if a third-party service goes down), and bloat your handlers (which could cause goroutine leakage or memory issues).

This is where a service (like [Temporal](https://github.com/temporalio/temporal)) suited for background/task processing comes in.

**If you're familiar with/already using Temporal**

If you're familiar with Temporal, Hatchet utilizes Temporal as a backend for processing workflows and activities, and adds a set of prebuilt workflows and utilities to make Temporal easier to use. For an understanding of how Hatchet works:

- Each Hatchet job corresponds to a different Temporal workflow
- Each step in a job corresponds to a Temporal activity

Hatchet is compatible with both Temporal Cloud and self-hosted versions of Temporal.

## I'd Like to Contribute

Hatchet is still in very early development -- as a result, there are no development docs. However, please feel free to reach out on the #contributing channel on [Slack](https://join.slack.com/t/hatchet-co/signup) to shape the direction of the project.
