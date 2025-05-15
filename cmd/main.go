package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"walks-of-italy/storage"
	"walks-of-italy/storage/db"

	tours "walks-of-italy"
)

func main() {
	// start := tours.NewDate(2025, time.September, 1)
	// availability, err := tours.PristineSistineEarly.GetAvailability(start, start.Add(0, 0, 3))
	availability, err := tours.PristineSistineEarly.GetLatestAvailability()
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}

	client, err := storage.New("test.db")
	if err != nil {
		log.Fatalf("error creating db client: %v", err)
	}
	defer client.Close()

	availabilityJSON, err := json.Marshal(availability)
	if err != nil {
		log.Fatalf("error encoding availability JSON: %v", err)
	}

	err = client.AddLatestAvailability(context.Background(), db.AddLatestAvailabilityParams{
		TourUuid:         tours.PristineSistineEarly.ProductID,
		AvailabilityDate: availability.LocalDateTimeStart,
		RawData:          sql.NullString{String: string(availabilityJSON)},
	})
	if err != nil {
		log.Fatalf("error storing availability: %v", err)
	}

	storedAvailability, err := client.GetLatestAvailability(context.Background(), tours.PristineSistineEarly.ProductID)
	if err != nil {
		log.Fatalf("error getting stored availability: %v", err)
	}

	fmt.Println(storedAvailability)
}
