package storage

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"walks-of-italy/storage/db"
	"walks-of-italy/tours"

	_ "embed"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

//go:generate sqlc generate

//go:embed schema.sql
var ddl string

type Client struct {
	*db.Queries
	db *sql.DB
}

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

func fromDB(tour db.Tour) *tours.TourDetail {
	return &tours.TourDetail{
		Name:      tour.Name,
		URL:       tour.Url,
		ProductID: tour.Uuid,
	}
}

func (c Client) Get(ctx context.Context, id string) (*tours.TourDetail, error) {
	asUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	tour, err := c.Queries.GetTour(ctx, asUUID)
	if err != nil {
		return nil, err
	}

	return fromDB(tour), nil
}

func (c Client) GetAll(ctx context.Context, query url.Values) ([]*tours.TourDetail, error) {
	results, err := c.Queries.ListTours(ctx)
	if err != nil {
		return nil, err
	}

	var result []*tours.TourDetail
	for _, item := range results {
		result = append(result, fromDB(item))
	}

	return result, nil
}

func (c Client) Set(ctx context.Context, tour *tours.TourDetail) error {
	return c.Queries.UpsertTour(ctx, db.UpsertTourParams{
		Uuid: tour.ProductID,
		Name: tour.Name,
		Url:  tour.URL,
	})
}

func (c Client) Delete(ctx context.Context, id string) error {
	asUUID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return c.Queries.DeleteTour(ctx, asUUID)
}
