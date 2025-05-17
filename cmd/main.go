package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	tours "walks-of-italy"
	"walks-of-italy/storage"
)

func main() {
	// updateData()
	// findTourForSevenPeople()
	watch()
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

func watch() {
	client, err := storage.New("test.db")
	if err != nil {
		log.Fatalf("error creating db client: %v", err)
	}
	defer client.Close()

	var nc *tours.NotifyClient
	appToken := os.Getenv("PUSHOVER_APP_TOKEN")
	recipientToken := os.Getenv("PUSHOVER_RECIPIENT_TOKEN")
	if appToken != "" && recipientToken != "" {
		nc, err = tours.NewNotifyClient(appToken, recipientToken)
		if err != nil {
			log.Fatalf("error creating notify client: %v", err)
		}
	}

	app := tours.NewApp(client, nc)

	slog.SetLogLoggerLevel(slog.LevelDebug)

	err = app.Watch(context.Background(), 15*time.Second)
	if err != nil {
		log.Fatalf("error watching: %v", err)
	}
}
