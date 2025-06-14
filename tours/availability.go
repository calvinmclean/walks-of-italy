package tours

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
)

func (td TourDetail) GetAvailability(ctx context.Context, accessToken string, start, end Date) (Availabilities, error) {
	requestBody := NewAvailabilityRequest(td.ProductID, start, end)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, availabilityURL, requestBody.JSON())
	if err != nil {
		return Availabilities{}, fmt.Errorf("error creating request: %w", err)
	}

	capabilities := []string{
		"octo/pricing",
		// "octo/content",
		// "octo/pickups",
		// "octo/extras",
		// "octo/offers",
		// "octo/resources",
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

type AvailabilityRequest struct {
	ProductID      uuid.UUID `json:"productId"`
	OptionID       string    `json:"optionId"`
	LocalDateStart Date      `json:"localDateStart"`
	LocalDateEnd   Date      `json:"localDateEnd"`
	Currency       string    `json:"currency"`
}

func NewAvailabilityRequest(productID uuid.UUID, start, end Date) AvailabilityRequest {
	return AvailabilityRequest{
		ProductID:      productID,
		OptionID:       "DEFAULT",
		LocalDateStart: start,
		LocalDateEnd:   end,
		Currency:       "USD",
	}
}

func (ar AvailabilityRequest) JSON() io.Reader {
	var r bytes.Buffer
	_ = json.NewEncoder(&r).Encode(ar)
	return &r
}

type Availabilities []AvailabilityDetail

func (a Availabilities) PrettySummary(w io.Writer) error {
	tmpl := template.Must(template.New("availability").
		Funcs(template.FuncMap{"truncate": func(s string, max int) string {
			if len(s) <= max {
				padding := max - len(s)
				return s + strings.Repeat(" ", padding)
			}
			return s[:max-3] + "..."
		}}).
		Parse(`
 Date                 | Price   | Vacancies
----------------------|---------|-------------
{{ range . -}}
{{ .LocalDateTimeStart.Format "2006-01-02 15:04:05" }}   | {{ .AdultPrice }} | {{ .Vacancies }}
{{ end }}`))

	return tmpl.Execute(w, a)
}

// AdultPrice returns the price for one adult in USD
func (a AvailabilityDetail) AdultPrice() string {
	var adultPricing UnitPricing
	for _, p := range a.UnitPricing {
		if p.UnitType == "ADULT" {
			adultPricing = p
			break
		}
	}

	return fmt.Sprintf("$%.2f", float64(adultPricing.Retail)/100.0)
}

// https://docs.ventrata.com/octo-core/availability
// Fields are removed to simplify the response
type AvailabilityDetail struct {
	ID                   time.Time      `json:"id"`
	LocalDateTimeStart   time.Time      `json:"localDateTimeStart"`
	LocalDateTimeEnd     time.Time      `json:"localDateTimeEnd"`
	AllDay               bool           `json:"allDay"`
	Available            bool           `json:"available"`
	Status               string         `json:"status"`
	Vacancies            int            `json:"vacancies"` // current vacancies
	Capacity             int            `json:"capacity"`  // actual total capacity of the tour
	PaxCount             int            `json:"paxCount"`  // currently-booked count
	MaxUnits             int            `json:"maxUnits"`  // available to sell
	UtcCutoffAt          time.Time      `json:"utcCutoffAt"`
	OpeningHours         []OpeningHours `json:"openingHours"`
	MeetingPoint         string         `json:"meetingPoint"`
	MeetingLocalDateTime time.Time      `json:"meetingLocalDateTime"`
	TourGroup            any            `json:"tourGroup"`
	Fare                 any            `json:"fare"`
	Notices              []any          `json:"notices"`
	UnitPricing          []UnitPricing  `json:"unitPricing"`
	Offers               []any          `json:"offers"`
	OfferCode            any            `json:"offerCode"`
	OfferTitle           any            `json:"offerTitle"`
	Offer                any            `json:"offer"`
	Pricing              Pricing        `json:"pricing"`
	PickupAvailable      bool           `json:"pickupAvailable"`
	PickupRequired       bool           `json:"pickupRequired"`
	PickupPoints         []any          `json:"pickupPoints"`
	HasResources         bool           `json:"hasResources"`
}

type OpeningHours struct {
	From            string `json:"from"`
	To              string `json:"to"`
	Frequency       any    `json:"frequency"`
	FrequencyAmount any    `json:"frequencyAmount"`
	FrequencyUnit   string `json:"frequencyUnit"`
}

type UnitPricing struct {
	UnitType          string `json:"unitType"`
	Original          int    `json:"original"`
	Retail            int    `json:"retail"`
	Currency          string `json:"currency"`
	CurrencyPrecision int    `json:"currencyPrecision"`
}

type Pricing struct {
	Original          int    `json:"original"`
	Retail            int    `json:"retail"`
	Currency          string `json:"currency"`
	CurrencyPrecision int    `json:"currencyPrecision"`
}
