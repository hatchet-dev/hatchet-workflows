/*
Workers can be created via

	w, err := worker.NewWorker(
	  // ... options
	)

You can then run the workers in blocking fashion using:

	w.Run()

Or in non-blocking fashion using:

	w.Start()

# Registering Integrations

You can register all integrations using the [WithIntegrations] option, which takes a list of integrations as an argument:

	  worker.NewWorker(
		worker.WithIntegrations(
		  myIntegration,
		),
	  )

# Adding Workflow Files

By default, the worker will load workflow files from the .hatchet directory. You can override this using the [WithWorkflowFiles] option:

	  worker.NewWorker(
		worker.WithWorkflowFiles(
			myWorkflowFile,
		),
	  )

# Connecting to Temporal

By default, the worker will connect to a Temporal instance using the following environment variables, which can be overriden:

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

You can also override the Temporal client using the [WithTemporalClient] option:

	  worker.NewWorker(
		worker.WithTemporalClient(
		  myTemporalClient,
		),
	  )
*/
package worker // import "github.com/hatchet-dev/hatchet-workflows/pkg/worker"
