package main

import (
	"context"
	"log"
	"log/slog"
	"walks-of-italy/storage"

	tours "walks-of-italy"
)

func main() {
	// start := tours.NewDate(2025, time.September, 1)
	// availability, err := tours.PristineSistineEarly.GetAvailability(start, start.Add(0, 0, 3))
	// availability, err := tours.PristineSistineEarly.GetLatestAvailability()
	// if err != nil {
	// 	log.Fatalf("request failed: %v", err)
	// }

	client, err := storage.New("test.db")
	if err != nil {
		log.Fatalf("error creating db client: %v", err)
	}
	defer client.Close()

	app := tours.NewApp(client)

	// updated, err := app.UpdateLatestAvailability(context.Background(), tours.PristineSistineEarly)
	// if err != nil {
	// 	log.Fatalf("error updating availability: %v", err)
	// }
	// fmt.Println("Updated:", updated)

	slog.SetLogLoggerLevel(slog.LevelDebug)

	err = app.UpdateLatestAvailabilities(context.Background())
	if err != nil {
		log.Fatalf("error updating availability: %v", err)
	}

	// storedAvailability, err := client.GetLatestAvailability(context.Background(), tours.PristineSistineEarly.ProductID)
	// if err != nil {
	// 	log.Fatalf("error getting stored availability: %v", err)
	// }
}
