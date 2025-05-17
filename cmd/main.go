package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	tours "walks-of-italy"
	"walks-of-italy/storage"
)

func main() {
	var debug bool
	var dbFilename, pushoverAppToken, pushoverRecipientToken string
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
				Value:       ":memory:",
			},
		},
		DefaultCommand: "watch",
		Commands: []*cli.Command{
			{
				Name: "watch",
				Flags: []cli.Flag{
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
					return watch(ctx.Context, dbFilename, pushoverAppToken, pushoverRecipientToken, watchInterval, debug)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func findTourForSevenPeople() {
	start := tours.DateFromTime(time.Now())
	end := start.Add(1, 0, 0)
	availability, err := tours.KeyMasterVatican.FindAvailability(context.Background(), start, end, func(a tours.AvailabilityDetail) bool {
		return a.Vacancies >= 7
	})
	if err != nil {
		log.Fatalf("error getting summary: %v", err)
	}

	err = availability.PrettySummary()
	if err != nil {
		log.Fatalf("error printing summary: %v", err)
	}
}

func updateData() {
	client, err := storage.New("test.db")
	if err != nil {
		log.Fatalf("error creating db client: %v", err)
	}
	defer client.Close()

	app := tours.NewApp(client, nil)

	slog.SetLogLoggerLevel(slog.LevelDebug)

	err = app.UpdateLatestAvailabilities(context.Background(), nil)
	if err != nil {
		log.Fatalf("error updating availability: %v", err)
	}

	err = app.PrettySummary(context.Background())
	if err != nil {
		log.Fatalf("error getting summary: %v", err)
	}
}

func watch(ctx context.Context, dbFilename, pushoverAppToken, pushoverRecipientToken string, interval time.Duration, debug bool) error {
	sc, err := storage.New(dbFilename)
	if err != nil {
		return fmt.Errorf("error creating db client: %w", err)
	}
	defer sc.Close()

	var nc *tours.NotifyClient
	if pushoverAppToken != "" && pushoverRecipientToken != "" {
		nc, err = tours.NewNotifyClient(pushoverAppToken, pushoverRecipientToken)
		if err != nil {
			return fmt.Errorf("error creating notify client: %w", err)
		}
	}

	app := tours.NewApp(sc, nc)

	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	err = app.Watch(ctx, interval)
	if err != nil {
		return fmt.Errorf("error watching: %w", err)
	}

	return nil
}
