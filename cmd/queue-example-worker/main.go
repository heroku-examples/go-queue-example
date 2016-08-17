package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/bgentry/que-go"
	qe "github.com/heroku-examples/go-queue-example"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
)

var (
	log     = logrus.WithField("cmd", "queue-example-worker")
	qc      *que.Client
	pgxpool *pgx.ConnPool
)

// indexURLJob would do whatever indexing is necessary in the background
func indexURLJob(j *que.Job) error {
	var ir qe.IndexRequest
	if err := json.Unmarshal(j.Args, &ir); err != nil {
		return errors.Wrap(err, "Unable to unmarshal job arguments into IndexRequest: "+string(j.Args))
	}

	log.WithField("IndexRequest", ir).Info("Processing IndexRequest! (not really)")
	// You would do real work here...

	return nil
}

func main() {
	var err error
	dbURL := os.Getenv("DATABASE_URL")
	pgxpool, qc, err = qe.Setup(dbURL)
	if err != nil {
		log.WithField("DATABASE_URL", dbURL).Fatal("Errors setting up the queue / database: ", err)
	}
	defer pgxpool.Close()

	wm := que.WorkMap{
		qe.IndexRequestJob: indexURLJob,
	}

	// 2 worker go routines
	workers := que.NewWorkerPool(qc, wm, 2)

	// Catch signal so we can shutdown gracefully
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go workers.Start()

	// Wait for a signal
	sig := <-sigCh
	log.WithField("signal", sig).Info("Signal received. Shutting down.")

	workers.Shutdown()
}
