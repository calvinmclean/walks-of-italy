package tours

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	availabilityURL = "https://api.ventrata.com/octo/availability"
)

type TourDetail struct {
	Name      string
	URL       string
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

func (ar AvailabilityRequest) JSON() io.Reader {
	var r bytes.Buffer
	_ = json.NewEncoder(&r).Encode(ar)
	return &r
}

func (td TourDetail) GetAvailability(ctx context.Context, accessToken string, start, end Date) (Availabilities, error) {
	requestBody := NewAvailabilityRequest(td.ProductID, start, end)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, availabilityURL, requestBody.JSON())
	if err != nil {
		return Availabilities{}, fmt.Errorf("error creating request: %w", err)
	}

	capabilities := []string{
		"octo/content",
		"octo/pricing",
		// "octo/pickups",
		// "octo/extras",
		"octo/offers",
		"octo/resources",
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Octo-Capabilities", strings.Join(capabilities, ","))
	req.Header.Set("Octo-Env", "live")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Availabilities{}, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Availabilities{}, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Availabilities{}, fmt.Errorf("unexpected response code: %d, body: %q", resp.StatusCode, string(body))
	}

	var result Availabilities
	err = json.Unmarshal(body, &result)
	if err != nil {
		return Availabilities{}, fmt.Errorf("error parsing response: %w", err)
	}

	return result, nil
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
