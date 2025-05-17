package tours

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

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

func (a *App) LogSummary(ctx context.Context) error {
	for _, tour := range Tours {
		availability, err := a.sc.GetLatestAvailability(ctx, tour.ProductID)
		if err != nil {
			return fmt.Errorf("error getting availability for tour %q: %w", tour, err)
		}

		a.logger.Info(
			tour.Name,
			"tour_id", tour.ProductID,
			"latest_availability", availability.AvailabilityDate,
			"recorded_at", availability.RecordedAt,
		)
	}
	return nil
}

func (a *App) PrettySummary(ctx context.Context) error {
	tmpl := template.Must(template.New("availability").
		Funcs(template.FuncMap{"truncate": func(s string, max int) string {
			if len(s) <= max {
				padding := max - len(s)
				return s + strings.Repeat(" ", padding)
			}
			return s[:max-3] + "..."
		}}).
		Parse(`
Tour Name                                                   | Available Date | Opened At
------------------------------------------------------------|----------------|----------------
{{ range . -}}
{{ truncate .TourName 59 }} | {{ .AvailabilityDate.Format "2006-01-02" }}     | {{ .RecordedAt.Format "2006-01-02 15:04:05" }}
{{ end }}`))

	tourData := []map[string]any{}

	for _, tour := range Tours {
		availability, err := a.sc.GetLatestAvailability(ctx, tour.ProductID)
		if err != nil {
			return fmt.Errorf("error getting availability for tour %q: %w", tour, err)
		}

		tourData = append(tourData, map[string]any{
			"TourName":         tour.Name,
			"AvailabilityDate": availability.AvailabilityDate,
			"RecordedAt":       availability.RecordedAt,
		})
	}
	return tmpl.Execute(os.Stdout, tourData)
}

func (a *App) UpdateLatestAvailabilities(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(len(Tours))

	errChan := make(chan error, len(Tours))
	for _, tour := range Tours {
		go func() {
			defer wg.Done()

			updated, err := a.UpdateLatestAvailability(ctx, tour)
			if err != nil {
				errChan <- fmt.Errorf("error updating availability for %q: %w", tour.ProductID, err)
			}

			a.logger.Debug("updated tour details", "tour_id", tour.ProductID, "changed", updated)
		}()
	}

	wg.Wait()
	close(errChan)

	errs := []error{}
	for err := range errChan {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (a *App) UpdateLatestAvailability(ctx context.Context, tour TourDetail) (bool, error) {
	availability, err := tour.GetLatestAvailability(ctx)
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

func (a *App) Watch(ctx context.Context, interval time.Duration) error {
	now := time.Now()

	next := now.Truncate(interval).Add(interval)
	untilNext := time.Until(next)

	a.logger.Debug("waiting to start", "duration", untilNext.String())
	select {
	case <-time.After(untilNext):
	case <-ctx.Done():
		return ctx.Err()
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	updateAvailabilities := func(t time.Time) error {
		a.logger.Debug("updating availabilities")
		err := a.UpdateLatestAvailabilities(ctx)
		if err != nil {
			return fmt.Errorf("error updating availabilities: %w", err)
		}
		a.logger.Debug("finished updating availabilities", "duration", time.Since(t).String())
		return nil
	}

	err := updateAvailabilities(time.Now())
	if err != nil {
		return err
	}

	for {
		select {
		case t := <-ticker.C:
			err := updateAvailabilities(t)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
