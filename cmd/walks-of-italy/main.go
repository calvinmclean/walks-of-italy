package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"time"

	"walks-of-italy/ai"
	"walks-of-italy/app"
	"walks-of-italy/storage"
	"walks-of-italy/tours"

	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

func main() {
	var debug bool
	var dbFilename, pushoverAppToken, pushoverRecipientToken, addr, accessToken, model, dataFile, tourID string
	var watchInterval time.Duration
	var searchStart, searchEnd cli.Timestamp
	app := &cli.App{
		Name: "walks-of-italy",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "enable debug logs",
				Destination: &debug,
				EnvVars:     []string{"DEBUG"},
			},
			&cli.StringFlag{
				Name:        "db",
				Usage:       "filename for SQLite database",
				Destination: &dbFilename,
				EnvVars:     []string{"DB"},
				Value:       "file::memory:?cache=shared",
				TakesFile:   true,
			},
			&cli.StringFlag{
				Name:        "pushover-app-token",
				Usage:       "App token for Pushover notifications",
				Destination: &pushoverAppToken,
				EnvVars:     []string{"PUSHOVER_APP_TOKEN"},
			},
			&cli.StringFlag{
				Name:        "pushover-recipient-token",
				Usage:       "Recipient token for Pushover notifications",
				Destination: &pushoverRecipientToken,
				EnvVars:     []string{"PUSHOVER_RECIPIENT_TOKEN"},
			},
			&cli.StringFlag{
				Name:        "access-token",
				Usage:       "Access token for Walks of Italy API",
				Destination: &accessToken,
				EnvVars:     []string{"ACCESS_TOKEN"},
			},
		},
		DefaultCommand: "watch",
		Commands: []*cli.Command{
			{
				Name:  "watch",
				Usage: "Watch for new tour availabilities",
				Flags: []cli.Flag{
					&cli.DurationFlag{
						Name:        "interval",
						Usage:       "Interval for polling new dates",
						Destination: &watchInterval,
						Value:       15 * time.Second,
						EnvVars:     []string{"INTERVAL"},
					},
				},
				Action: func(ctx *cli.Context) error {
					app, sc, err := setupApp(addr, dbFilename, pushoverAppToken, pushoverRecipientToken, accessToken, debug)
					if err != nil {
						return fmt.Errorf("error creating app: %w", err)
					}
					defer sc.Close()
					return app.Watch(ctx.Context, watchInterval)
				},
			},
			{
				Name:  "update",
				Usage: "Update latest availabilities",
				Action: func(ctx *cli.Context) error {
					app, sc, err := setupApp(addr, dbFilename, pushoverAppToken, pushoverRecipientToken, accessToken, debug)
					if err != nil {
						return fmt.Errorf("error creating app: %w", err)
					}
					defer sc.Close()

					allTours, err := sc.GetAll(ctx.Context, url.Values{})
					if err != nil {
						return fmt.Errorf("error getting tours: %w", err)
					}

					err = app.UpdateLatestAvailabilities(ctx.Context, allTours, nil)
					if err != nil {
						return fmt.Errorf("error updating availability: %w", err)
					}

					err = app.PrettySummary(ctx.Context, os.Stdout, allTours)
					if err != nil {
						return fmt.Errorf("error getting summary: %w", err)
					}

					return nil
				},
			},
			{
				Name:  "details",
				Usage: "Print details from details API",
				Action: func(ctx *cli.Context) error {
					sc, err := storage.New(dbFilename)
					if err != nil {
						return fmt.Errorf("error creating db client: %w", err)
					}
					defer sc.Close()

					allTours, err := sc.GetAll(ctx.Context, url.Values{})
					if err != nil {
						return fmt.Errorf("error getting tours: %w", err)
					}

					for _, td := range allTours {
						desc, err := td.GetDescription(context.Background(), td.ApiUrl)
						if err != nil {
							return fmt.Errorf("error getting description for %q: %w", td.Name, err)
						}

						fmt.Println(td.Name)
						fmt.Println(desc)
						fmt.Println()
					}

					return nil
				},
			},
			{
				Name:  "chat",
				Usage: "Chat with an AI model about the tour dates",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "model",
						Usage:       "model name to interact with in Ollama",
						Destination: &model,
						EnvVars:     []string{"MODEL"},
						Value:       "qwen2.5:7b",
					},
				},
				Action: func(ctx *cli.Context) error {
					sc, err := storage.New(dbFilename)
					if err != nil {
						return fmt.Errorf("error creating db client: %w", err)
					}
					defer sc.Close()

					if debug {
						slog.SetLogLoggerLevel(slog.LevelDebug)
					}

					return ai.Chat(sc, model, accessToken)
				},
			},
			{
				Name:  "search",
				Usage: "Search for availability of a specified tour in a date range",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "tour-id",
						Usage:       "UUID of a tour to get availability for",
						Destination: &tourID,
						Required:    true,
					},
					&cli.TimestampFlag{
						Name:        "start",
						Usage:       "Date to start the search from",
						Destination: &searchStart,
						Layout:      time.DateOnly,
						Required:    true,
					},
					&cli.TimestampFlag{
						Name:        "end",
						Usage:       "Date to end the search at",
						Destination: &searchEnd,
						Layout:      time.DateOnly,
						Required:    true,
					},
				},
				Action: func(ctx *cli.Context) error {
					tourUUID, err := uuid.Parse(tourID)
					if err != nil {
						return err
					}

					tour := tours.TourDetail{
						Name:      "User-provided tour ID",
						ProductID: tourUUID,
					}

					availability, err := tour.GetAvailability(ctx.Context, accessToken, tours.DateFromTime(*searchStart.Value()), tours.DateFromTime(*searchEnd.Value()))
					if err != nil {
						return fmt.Errorf("error getting availability: %w", err)
					}

					err = availability.PrettySummary(os.Stdout)
					if err != nil {
						return fmt.Errorf("error printing summary: %w", err)
					}

					return nil
				},
			},
			{
				Name:  "serve",
				Usage: "Run server with API and UI. Also watches for new availability",
				Flags: []cli.Flag{
					&cli.DurationFlag{
						Name:        "interval",
						Usage:       "Interval for polling new dates",
						Destination: &watchInterval,
						Value:       15 * time.Second,
						EnvVars:     []string{"INTERVAL"},
					},
					&cli.StringFlag{
						Name:        "addr",
						Usage:       "address to serve on",
						Destination: &addr,
						Value:       ":7077",
						EnvVars:     []string{"ADDR"},
					},
				},
				Action: func(ctx *cli.Context) error {
					app, sc, err := setupApp(addr, dbFilename, pushoverAppToken, pushoverRecipientToken, accessToken, debug)
					if err != nil {
						return fmt.Errorf("error creating app: %w", err)
					}
					defer sc.Close()

					return app.Run(ctx.Context, watchInterval)
				},
			},
			{
				Name:  "load",
				Usage: "Load data from a JSON file into the DB",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "data",
						Usage:       "filename for JSON data to load",
						Destination: &dataFile,
						TakesFile:   true,
					},
				},
				Action: func(ctx *cli.Context) error {
					sc, err := storage.New(dbFilename)
					if err != nil {
						return fmt.Errorf("error creating db client: %w", err)
					}
					defer sc.Close()

					jsonData, err := os.ReadFile(dataFile)
					if err != nil {
						return fmt.Errorf("error reading JSON file: %w", err)
					}

					var tours []*tours.TourDetail
					err = json.Unmarshal(jsonData, &tours)
					if err != nil {
						return fmt.Errorf("error parsing JSON data: %w", err)
					}

					for _, td := range tours {
						if td.ProductID == (uuid.UUID{}) {
							td.ProductID = uuid.New()
						}
						err = sc.Set(ctx.Context, td)
						if err != nil {
							return fmt.Errorf("error inserting tour %q: %w", td.Name, err)
						}
					}

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func setupApp(addr, dbFilename, pushoverAppToken, pushoverRecipientToken, accessToken string, debug bool) (*app.App, *storage.Client, error) {
	sc, err := storage.New(dbFilename)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating db client: %w", err)
	}

	var nc *app.NotifyClient
	if pushoverAppToken != "" && pushoverRecipientToken != "" {
		nc, err = app.NewNotifyClient(pushoverAppToken, pushoverRecipientToken)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating notify client: %w", err)
		}
	}

	app := app.New(addr, accessToken, sc, nc)

	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	return app, sc, nil
}
