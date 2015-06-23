package go_queue_example

import (
	"github.com/bgentry/que-go"
	"github.com/jackc/pgx"
)

const (
	IndexRequestJob = "IndexRequests"
)

func GetPgxPool(dbURL string) (*pgx.ConnPool, error) {
	pgxcfg, err := pgx.ParseURI(dbURL)
	if err != nil {
		return nil, err
	}

	pgxpool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:   pgxcfg,
		AfterConnect: que.PrepareStatements,
	})

	if err != nil {
		return nil, err
	}

	return pgxpool, nil
}
