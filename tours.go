package tours

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	accessToken     = "b082247d-90d9-4e44-8623-3a8a4b5c14de"
	availabilityURL = "https://api.ventrata.com/octo/availability"
)

var (
	KeyMasterVatican = TourDetail{
		Name:      "VIP Vatican Key Master’s Tour: Unlock the Sistine Chapel",
		URL:       "https://www.walksofitaly.com/vatican-tours/key-masters-tour-sistine-chapel-vatican-museums/",
		ProductID: uuid.MustParse("e9d2d819-5f04-4b1f-a07f-612387494b8f"),
	}
	PrivateVatican = TourDetail{
		Name:      "Private Vatican Tour: Vatican Museums, Sistine Chapel & St. Peter’s",
		URL:       "https://www.walksofitaly.com/vatican-tours/private-vatican-tour/",
		ProductID: uuid.MustParse("c40d8e0e-6756-463b-a052-982c77a707aa"),
	}
	CompleteVatican = TourDetail{
		Name:      "The Complete Vatican Tour with Vatican Museums, Sistine Chapel & St. Peter’s Basilica",
		URL:       "https://www.walksofitaly.com/vatican-tours/complete-vatican-tour/",
		ProductID: uuid.MustParse("3b263ef8-c280-49cc-a74f-ac95aa2f1b58"),
	}
	AloneInTheSistineChapel = TourDetail{
		Name:      "Alone in the Sistine Chapel: VIP Entry at the Vatican’s Most Exclusive Hours",
		URL:       "https://www.walksofitaly.com/vatican-tours/vatican-after-hours-tour/",
		ProductID: uuid.MustParse("8c14824f-905d-4273-8b83-10b567db6e55"),
	}
	PristineSistineEarly = TourDetail{
		Name:      "Pristine Sistine™ Early Entrance Small Group Vatican Tour",
		URL:       "https://www.walksofitaly.com/vatican-tours/pristine-sistine-chapel-tour/",
		ProductID: uuid.MustParse("a1249220-e5d8-4983-93b2-c31ddfb3ccb8"),
	}

	Tours = []TourDetail{
		KeyMasterVatican,
		PrivateVatican,
		CompleteVatican,
		AloneInTheSistineChapel,
		PristineSistineEarly,
	}
)

type TourDetail struct {
	Name      string
	URL       string
	ProductID uuid.UUID
}

func (ar AvailabilityRequest) JSON() io.Reader {
	var r bytes.Buffer
	_ = json.NewEncoder(&r).Encode(ar)
	return &r
}

func (td TourDetail) GetAvailability(start, end Date) (AvailabilityResponse, error) {
	requestBody := NewAvailabilityRequest(td.ProductID, start, end)
	req, err := http.NewRequest(http.MethodPost, availabilityURL, requestBody.JSON())
	if err != nil {
		return AvailabilityResponse{}, fmt.Errorf("error creating request: %w", err)
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
		return AvailabilityResponse{}, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AvailabilityResponse{}, fmt.Errorf("error reading response: %w", err)
	}

	var result AvailabilityResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return AvailabilityResponse{}, fmt.Errorf("error parsing response: %w", err)
	}

	return result, nil
}

func (td TourDetail) GetLatestAvailability() (AvailabilityDetail, error) {
	start := DateFromTime(time.Now())
	end := start.Add(1, 0, 0)

	availability, err := td.GetAvailability(start, end)
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
