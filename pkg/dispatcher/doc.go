/*
The dispatcher package provides a simple interface for triggering workflows using event keys.

# Usage

The dispatcher package can be used to trigger workflows using event keys. For example, if you have a workflow which listens to the `user:create` event, you can trigger it using:

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

# Adding Workflow Files

By default, the dispatcher will load workflow files from the .hatchet directory. You can override this using the [WithWorkflowFiles] option:

	  dispatcher.NewDispatcher(
		dispatcher.WithWorkflowFiles(
			myWorkflowFile,
		),
	  )

# Connecting to Temporal

By default, the dispatcher will connect to a Temporal instance using the following environment variables, which can be overriden:

	TEMPORAL_CLIENT_HOST_PORT=127.0.0.1:7233
	TEMPORAL_CLIENT_NAMESPACE=default

The default client supports the following options:

	TEMPORAL_CLIENT_HOST_PORT
	TEMPORAL_CLIENT_NAMESPACE
	TEMPORAL_CLIENT_TLS_ROOT_CA
	TEMPORAL_CLIENT_TLS_ROOT_CA_FILE
	TEMPORAL_CLIENT_TLS_CERT
	TEMPORAL_CLIENT_TLS_CERT_FILE
	TEMPORAL_CLIENT_TLS_KEY
	TEMPORAL_CLIENT_TLS_KEY_FILE
	TEMPORAL_CLIENT_TLS_SERVER_NAME

You can also override the Hatchet client (which generate a Temporal client) using the [WithHatchetClient] option:

	  dispatcher.NewWorker(
		worker.WithHatchetClient(
		  myHatchetClient,
		),
	  )

Unlike the worker, it is not possible to set the Temporal client directly. This is to avoid a situation where the dispatcher and worker are using different queues, but may be changed in the future.

See the [client] package for more information about setting up the Hatchet client.
*/
package dispatcher // import "github.com/hatchet-dev/hatchet-workflows/pkg/dispatcher"
