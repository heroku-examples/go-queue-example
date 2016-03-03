package go_queue_example

import (
	que "github.com/bgentry/que-go"
	"github.com/jackc/pgx"
)

// IndexRequest container.
// The URL is the url to index content from.
type IndexRequest struct {
	URL string `json:url`
}

const (
	IndexRequestJob = "IndexRequests" // IndexRequestJob queue name

	QueTableSQL = `
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
		);` // QueTableSQL to create table idempotently
)

// prepQue ensures that the que table exists and que's prepared statements are
// run. It is meant to be used in a pgx.ConnPool's AfterConnect hook.
func prepQue(conn *pgx.Conn) error {
	_, err := conn.Exec(QueTableSQL)
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

// Setup a *pgx.ConnPool and *que.Client
// This is here so that setup routines can easily be shared between web and
// workers
func Setup(dbURL string) (*pgx.ConnPool, *que.Client, error) {
	pgxpool, err := GetPgxPool(dbURL)
	if err != nil {
		return nil, nil, err
	}

	qc := que.NewClient(pgxpool)

	return pgxpool, qc, err
}
