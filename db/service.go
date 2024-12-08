package db

import (
	"database/sql"
)

type Service struct {
	Db      *sql.DB
	Queries *Queries
}

func NewSqliteService() (*Service, error) {
	conn, err := sql.Open("sqlite3", "./db/database.sqlite")

	if err != nil {
		return nil, err
	}

	return &Service{Db: conn, Queries: New(conn)}, nil
}
