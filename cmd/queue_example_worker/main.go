package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	qe "github.com/heroku-examples/go_queue_example"

	log "github.com/heroku-examples/go_queue_example/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/heroku-examples/go_queue_example/Godeps/_workspace/src/github.com/bgentry/que-go"
	"github.com/heroku-examples/go_queue_example/Godeps/_workspace/src/github.com/jackc/pgx"
)

var (
	qc      *que.Client
	pgxpool *pgx.ConnPool
)

// indexURLJob would do whatever indexing is necessary in the background
func indexURLJob(j *que.Job) error {
	var ir qe.IndexRequest
	err := json.Unmarshal(j.Args, &ir)
	if err != nil {
		log.WithField("args", string(j.Args)).Error("Unable to unmarshal job arguments into IndexRequest")
		return err
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
