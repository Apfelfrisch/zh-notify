package db

import (
	"database/sql"
)

func NewSqliteConn() (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", "database.sqlite")

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func NewQueries() (*Queries, error) {
	conn, err := NewSqliteConn()

	if err != nil {
		return nil, err
	}

	return New(conn), nil
}
