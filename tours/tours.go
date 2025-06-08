package tours

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	availabilityURL = "https://api.ventrata.com/octo/availability"
)

type TourDetail struct {
	Name      string
	Link      string
	ApiUrl    string
	ProductID uuid.UUID
}

func (td TourDetail) GetID() string {
	return td.ProductID.String()
}

func (td *TourDetail) Bind(r *http.Request) error {
	return nil
}

func (*TourDetail) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// availabilityFilter allows filtering available dates by criteria like price and number of tickets.
// The filter should return "true" if the date is considered to be available.
func (td TourDetail) FindAvailability(ctx context.Context, accessToken string, start, end Date, availabilityFilter func(AvailabilityDetail) bool) (Availabilities, error) {
	if availabilityFilter == nil {
		return nil, errors.New("missing availabilityFilter")
	}

	availability, err := td.GetAvailability(ctx, accessToken, start, end)
	if err != nil {
		return nil, fmt.Errorf("error getting availability: %w", err)
	}

	var result []AvailabilityDetail
	for _, a := range availability {
		if availabilityFilter(a) {
			result = append(result, a)
		}
	}

	return result, nil
}

func (td TourDetail) GetLatestAvailability(ctx context.Context, accessToken string) (AvailabilityDetail, error) {
	start := DateFromTime(time.Now())
	end := start.Add(1, 0, 0)

	availability, err := td.GetAvailability(ctx, accessToken, start, end)
	if err != nil {
		return AvailabilityDetail{}, fmt.Errorf("error getting availability: %w", err)
	}

	latest := AvailabilityDetail{LocalDateTimeStart: start.ToTime()}
	for _, a := range availability {
		if !a.Available {
			continue
		}

		if a.LocalDateTimeStart.After(latest.LocalDateTimeStart) {
			latest = a
		}
	}

	return latest, nil
}
