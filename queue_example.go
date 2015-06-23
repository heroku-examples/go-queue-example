package go_queue_example

import (
	"os"

	log "github.com/heroku-examples/go_queue_example/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	que "github.com/heroku-examples/go_queue_example/Godeps/_workspace/src/github.com/bgentry/que-go"
	"github.com/heroku-examples/go_queue_example/Godeps/_workspace/src/github.com/jackc/pgx"
)

type IndexRequest struct {
	URL string `json:url`
}

const (
	IndexRequestJob = "IndexRequests"

	TableSQL = `
		CREATE TABLE IF NOT EXISTS que_jobs
		(
			priority    smallint    NOT NULL DEFAULT 100,
			run_at      timestamptz NOT NULL DEFAULT now(),
			job_id      bigserial   NOT NULL,
			job_class   text        NOT NULL,
			args        json        NOT NULL DEFAULT '[]'::json,
			error_count integer     NOT NULL DEFAULT 0,
			last_error  text,
			queue       text        NOT NULL DEFAULT '',

			CONSTRAINT que_jobs_pkey PRIMARY KEY (queue, priority, run_at, job_id)
		);
	`
)

var (
	PgxPool *pgx.ConnPool
	Qc      *que.Client
)

// prepQue ensures that the que table exists and que's prepared statements are
// run. It is meant to be used in a pgx.ConnPool's AfterConnect hook.
func prepQue(conn *pgx.Conn) error {
	_, err := conn.Exec(TableSQL)
	if err != nil {
		return err
	}

	return que.PrepareStatements(conn)
}

// GetPgxPool based on the provided database URL
func GetPgxPool(dbURL string) (*pgx.ConnPool, error) {
	pgxcfg, err := pgx.ParseURI(dbURL)
	if err != nil {
		return nil, err
	}

	pgxpool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:   pgxcfg,
		AfterConnect: prepQue,
	})

	if err != nil {
		return nil, err
	}

	return pgxpool, nil
}

func init() {
	var err error

	dbURL := os.Getenv("DATABASE_URL")
	PgxPool, err = GetPgxPool(dbURL)
	if err != nil {
		log.WithField("DATABASE_URL", dbURL).Fatal(err)
	}

	Qc = que.NewClient(PgxPool)
}
