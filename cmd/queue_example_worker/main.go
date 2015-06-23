package main

import (
	"os"
	"os/signal"
	"syscall"

	qe "github.com/heroku-examples/go_queue_example"

	log "github.com/heroku-examples/go_queue_example/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/heroku-examples/go_queue_example/Godeps/_workspace/src/github.com/bgentry/que-go"
)

// indexURLJob would do whatever indexing is necessary in the background
func indexURLJob(j *que.Job) error {
	log.WithField("job", j).Info("Processing Job Now! (not really)")
	return nil
}

func main() {
	defer qe.PgxPool.Close()

	wm := que.WorkMap{
		qe.IndexRequestJob: indexURLJob,
	}

	// 2 worker go routines
	workers := que.NewWorkerPool(qe.Qc, wm, 2)

	// Catch signal so we can gracefully shutdown
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go workers.Start()

	// Wait for a signal
	sig := <-sigCh
	log.WithField("signal", sig).Info("Signal received. Shutting down.")

	workers.Shutdown()
}
