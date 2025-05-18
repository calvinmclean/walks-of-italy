package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"text/template"
	"time"

	"walks-of-italy/storage"
	"walks-of-italy/storage/db"
	"walks-of-italy/tours"

	"github.com/calvinmclean/babyapi"
	"github.com/go-chi/render"
)

type App struct {
	sc          *storage.Client
	nc          *NotifyClient
	api         *babyapi.API[*tours.TourDetail]
	accessToken string
	logger      slog.Logger
}

func New(addr string, accessToken string, sc *storage.Client, nc *NotifyClient) *App {
	api := babyapi.
		NewAPI("Tours", "/tours", func() *tours.TourDetail { return &tours.TourDetail{} }).
		SetAddress(addr).
		SetStorage(sc)

	return &App{sc: sc, nc: nc, api: api, accessToken: accessToken, logger: *slog.Default()}
}

func (a *App) Run(ctx context.Context, watchInterval time.Duration) error {
	ctx, cancel := context.WithCancel(ctx)

	var watchErr error
	go func() {
		watchErr = a.Watch(ctx, watchInterval)
		if watchErr != nil {
			cancel()
		}
	}()
	err := a.api.
		WithContext(ctx).
		SetOnCreateOrUpdate(func(w http.ResponseWriter, r *http.Request, td *tours.TourDetail) *babyapi.ErrResponse {
			updated, err := a.UpdateLatestAvailability(r.Context(), *td)
			if err != nil {
				a.logger.Error("error updating availability", "tour_id", td.ProductID, "err", err)
				return nil
			}
			a.logger.Debug("updated tour details", "tour_id", td.ProductID, "changed", updated != nil)
			return nil
		}).
		AddCustomRoute(http.MethodGet, "/summary", babyapi.Handler(a.SummarizeLatestAvailabilities)).
		AddCustomIDRoute(http.MethodGet, "/summary", a.api.GetRequestedResourceAndDo(a.SummarizeTourDates)).
		Serve()

	if err != nil {
		cancel()
	}
	return errors.Join(watchErr, err)
}

func (a *App) SummarizeLatestAvailabilities(w http.ResponseWriter, r *http.Request) render.Renderer {
	allTours, err := a.sc.GetAll(r.Context(), url.Values{})
	if err != nil {
		return babyapi.ErrInvalidRequest(fmt.Errorf("error getting tours: %w", err))
	}

	err = a.PrettySummary(r.Context(), w, allTours)
	if err != nil {
		return babyapi.ErrInvalidRequest(fmt.Errorf("error creating summary: %w", err))
	}

	return nil
}

func (a *App) SummarizeTourDates(w http.ResponseWriter, r *http.Request, td *tours.TourDetail) (render.Renderer, *babyapi.ErrResponse) {
	start := tours.DateFromTime(time.Now())
	end := start.Add(1, 0, 0)
	availability, err := td.FindAvailability(r.Context(), a.accessToken, start, end, func(a tours.AvailabilityDetail) bool {
		return true
	})
	if err != nil {
		return nil, babyapi.ErrInvalidRequest(fmt.Errorf("error getting availability: %w", err))
	}

	err = availability.PrettySummary(w)
	if err != nil {
		return nil, babyapi.ErrInvalidRequest(fmt.Errorf("error creating summary: %w", err))
	}

	return nil, nil
}

func (a *App) LogSummary(ctx context.Context, tours []tours.TourDetail) error {
	for _, tour := range tours {
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

func (a *App) PrettySummary(ctx context.Context, w io.Writer, tours []*tours.TourDetail) error {
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

	for _, tour := range tours {
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
	return tmpl.Execute(w, tourData)
}

func (a *App) UpdateLatestAvailabilities(ctx context.Context, tours []*tours.TourDetail, onUpdate func(tours.TourDetail, tours.AvailabilityDetail)) error {
	var wg sync.WaitGroup
	wg.Add(len(tours))

	errChan := make(chan error, len(tours))
	for _, tour := range tours {
		go func() {
			defer wg.Done()

			updated, err := a.UpdateLatestAvailability(ctx, *tour)
			if err != nil {
				errChan <- fmt.Errorf("error updating availability for %q: %w", tour.ProductID, err)
			}
			a.logger.Debug("updated tour details", "tour_id", tour.ProductID, "changed", updated != nil)

			if onUpdate != nil && updated != nil {
				onUpdate(*tour, *updated)
			}
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

func (a *App) UpdateLatestAvailability(ctx context.Context, tour tours.TourDetail) (*tours.AvailabilityDetail, error) {
	availability, err := tour.GetLatestAvailability(ctx, a.accessToken)
	if err != nil {
		return nil, fmt.Errorf("error getting availability: %w", err)
	}

	storedAvailability, err := a.sc.GetLatestAvailability(ctx, tour.ProductID)
	if errors.Is(err, sql.ErrNoRows) {
		err = a.storeLatestAvailability(ctx, tour, availability)
		if err != nil {
			return nil, err
		}
		return &availability, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting stored availability: %w", err)
	}

	// If stored date is already the latest, don't save
	if storedAvailability.AvailabilityDate.Compare(availability.LocalDateTimeStart) >= 0 {
		return nil, nil
	}

	err = a.storeLatestAvailability(ctx, tour, availability)
	if err != nil {
		return nil, err
	}

	return &availability, nil
}

func (a *App) storeLatestAvailability(ctx context.Context, tour tours.TourDetail, availability tours.AvailabilityDetail) error {
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

func (a *App) updateAvailabilitiesForWatch(ctx context.Context, t time.Time) error {
	allTours, err := a.sc.GetAll(ctx, url.Values{})
	if err != nil {
		return fmt.Errorf("error getting tours: %w", err)
	}

	a.logger.Debug("updating availabilities")
	err = a.UpdateLatestAvailabilities(ctx, allTours, func(tour tours.TourDetail, availability tours.AvailabilityDetail) {
		if a.nc == nil {
			return
		}

		err := a.nc.Send(
			"New tour availabilities posted",
			fmt.Sprintf("Tour: %s\nDate: %s", tour.Name, availability.LocalDateTimeStart.Format(time.DateOnly)),
		)
		if err != nil {
			a.logger.Error("error sending notification", "err", err)
		}
	})
	if err != nil {
		return fmt.Errorf("error updating availabilities: %w", err)
	}
	a.logger.Debug("finished updating availabilities", "duration", time.Since(t).String())
	return nil
}

func (a *App) Watch(ctx context.Context, interval time.Duration) error {
	err := a.updateAvailabilitiesForWatch(ctx, time.Now())
	if err != nil {
		a.logger.Error("error updating availabilities", "err", err)
	}

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

	for {
		select {
		case t := <-ticker.C:
			err := a.updateAvailabilitiesForWatch(ctx, t)
			if err != nil {
				a.logger.Error("error updating availabilities", "err", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
