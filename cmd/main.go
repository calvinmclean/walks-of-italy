package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"walks-of-italy/app"
	"walks-of-italy/storage"
	"walks-of-italy/tours"
)

func main() {
	var debug bool
	var dbFilename, pushoverAppToken, pushoverRecipientToken, addr string
	var watchInterval time.Duration
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
		},
		DefaultCommand: "watch",
		Commands: []*cli.Command{
			{
				Name: "watch",
				Flags: []cli.Flag{
					&cli.DurationFlag{
						Name:        "interval",
						Usage:       "Interval for polling new dates",
						Destination: &watchInterval,
						Value:       15 * time.Second,
						EnvVars:     []string{"INTERVAL"},
					},
				},
				Description: "Watch for new tour availabilities",
				Action: func(ctx *cli.Context) error {
					app, close, err := setupApp(addr, dbFilename, pushoverAppToken, pushoverRecipientToken, debug)
					if err != nil {
						return fmt.Errorf("error creating app: %w", err)
					}
					defer close()
					return app.Watch(ctx.Context, watchInterval)
				},
			},
			{
				Name:        "update",
				Description: "Update local data",
				Action: func(ctx *cli.Context) error {
					app, close, err := setupApp(addr, dbFilename, pushoverAppToken, pushoverRecipientToken, debug)
					if err != nil {
						return fmt.Errorf("error creating app: %w", err)
					}
					defer close()

					err = app.UpdateLatestAvailabilities(ctx.Context, nil)
					if err != nil {
						return fmt.Errorf("error updating availability: %w", err)
					}

					err = app.PrettySummary(ctx.Context)
					if err != nil {
						return fmt.Errorf("error getting summary: %w", err)
					}

					return nil
				},
			},
			{
				Name:        "search",
				Description: "search for tours",
				Action: func(ctx *cli.Context) error {
					start := tours.DateFromTime(time.Now())
					end := start.Add(1, 0, 0)
					availability, err := tours.KeyMasterVatican.FindAvailability(ctx.Context, start, end, func(a tours.AvailabilityDetail) bool {
						return a.Vacancies >= 7
					})
					if err != nil {
						return fmt.Errorf("error getting summary: %w", err)
					}

					err = availability.PrettySummary()
					if err != nil {
						return fmt.Errorf("error printing summary: %w", err)
					}

					return nil
				},
			},
			{
				Name: "serve",
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
				Description: "Watch for new tour availabilities",
				Action: func(ctx *cli.Context) error {
					app, close, err := setupApp(addr, dbFilename, pushoverAppToken, pushoverRecipientToken, debug)
					if err != nil {
						return fmt.Errorf("error creating app: %w", err)
					}
					defer close()

					return app.Run(ctx.Context, watchInterval)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func setupApp(addr, dbFilename, pushoverAppToken, pushoverRecipientToken string, debug bool) (*app.App, func(), error) {
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

	app := app.New(addr, sc, nc)

	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	return app, sc.Close, nil
}
