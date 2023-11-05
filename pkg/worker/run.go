package worker

import "context"

// StartWorkers starts a set of workers. It will block until the context is cancelled.
func StartWorkers(ctx context.Context, workers ...Worker) error {
	errCh := make(chan bool)

	go func() {
		switch {

		}
	}()

	for _, worker := range workers {
		err := worker.Start()
		if err != nil {
			errCh <- true
			return err
		}
	}

	isDone := false

	for !isDone {
		select {
		case <-ctx.Done():
			isDone = true
			break
		case <-errCh:
			isDone = true
			break
		}
	}

	for _, worker := range workers {
		worker.Stop()
	}

	return nil
}
