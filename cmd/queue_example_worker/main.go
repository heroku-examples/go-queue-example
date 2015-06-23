package main

import (
	"os"
	"os/signal"
	"syscall"

	qe "github.com/heroku-examples/go_queue_example"

	log "github.com/Sirupsen/logrus"
	"github.com/bgentry/que-go"
)

var (
	qc *que.Client
)

// indexURLJob would do whatever indexing is necessary in the background
func indexURLJob(j *que.Job) error {
	log.WithField("job", j).Info("Processing Job Now! (not really)")
	return nil
}

func main() {
	port := os.Getenv("PORT")
	if port != "" {
		log.WithField("PORT", port).Fatal("$PORT must be set")
	}

	dbURL := os.Getenv("DATABASE_URL")
	pgxpool, err := qe.GetPgxPool(dbURL)
	if err != nil {
		log.WithField("DATABASE_URL", dbURL).Fatal(err)
	}
	defer pgxpool.Close()

	qc = que.NewClient(pgxpool)

	wm := que.WorkMap{
		qe.IndexRequestJob: indexURLJob,
	}

	// 2 worker go routines
	workers := que.NewWorkerPool(qc, wm, 2)

	// Catch signal so we can gracefully shutdown
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go workers.Start()

	// Wait for a signal
	sig := <-sigCh
	log.WithField("signal", sig).Info("Signal received. Shutting down.")

	workers.Shutdown()
}
