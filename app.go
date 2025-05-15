package tours

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"walks-of-italy/storage"
	"walks-of-italy/storage/db"
)

type App struct {
	sc     *storage.Client
	logger slog.Logger
}

func NewApp(sc *storage.Client) *App {
	return &App{sc: sc, logger: *slog.Default()}
}

func (a *App) UpdateLatestAvailabilities(ctx context.Context) error {
	for _, tour := range Tours {
		updated, err := a.UpdateLatestAvailability(ctx, tour)
		if err != nil {
			return fmt.Errorf("error updating availability for %q: %w", tour.ProductID, err)
		}

		a.logger.Debug("updated tour details", "tour_id", tour.ProductID, "changed", updated)
	}
	return nil
}

func (a *App) UpdateLatestAvailability(ctx context.Context, tour TourDetail) (bool, error) {
	availability, err := tour.GetLatestAvailability()
	if err != nil {
		return false, fmt.Errorf("error getting availability: %w", err)
	}

	storedAvailability, err := a.sc.GetLatestAvailability(ctx, tour.ProductID)
	if errors.Is(err, sql.ErrNoRows) {
		err = a.storeLatestAvailability(ctx, tour, availability)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("error getting stored availability: %w", err)
	}

	// If stored date is already the latest, don't save
	if storedAvailability.AvailabilityDate.Compare(availability.LocalDateTimeStart) >= 0 {
		return false, nil
	}

	err = a.storeLatestAvailability(ctx, tour, availability)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (a *App) storeLatestAvailability(ctx context.Context, tour TourDetail, availability AvailabilityDetail) error {
	availabilityJSON, err := json.Marshal(availability)
	if err != nil {
		return fmt.Errorf("error marshalling availability JSON: %w", err)
	}

	err = a.sc.AddLatestAvailability(ctx, db.AddLatestAvailabilityParams{
		TourUuid:         tour.ProductID,
		AvailabilityDate: availability.LocalDateTimeStart,
		RawData:          string(availabilityJSON),
	})
	if err != nil {
		return fmt.Errorf("error storing availability: %w", err)
	}

	return nil
}
