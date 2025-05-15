package storage

import (
	"database/sql"
	"fmt"

	"walks-of-italy/storage/db"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:generate sqlc generate

//go:embed schema.sql
var ddl string

type Client struct {
	*db.Queries
	db *sql.DB
}

// ":memory:" is in-memory for filename
// "file::memory:?cache=shared"
func New(filename string) (*Client, error) {
	database, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	err = database.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	_, err = database.Exec(ddl)
	if err != nil {
		return nil, fmt.Errorf("error creating tables: %w", err)
	}

	return &Client{
		db.New(database),
		database,
	}, nil
}

func (c Client) Close() {
	c.db.Close()
}
