package main

import (
	"fmt"
	"log"

	tours "walks-of-italy"
)

func main() {
	// start := tours.NewDate(2025, time.September, 1)
	// availability, err := tours.PristineSistineEarly.GetAvailability(start, start.Add(0, 0, 3))
	availability, err := tours.PristineSistineEarly.GetLatestAvailability()
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}

	fmt.Println(availability)
}
