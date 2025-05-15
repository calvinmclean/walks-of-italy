package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	accessToken     = "b082247d-90d9-4e44-8623-3a8a4b5c14de"
	availabilityURL = "https://api.ventrata.com/octo/availability"
)

var tours = struct {
	KeyMasterVatican        TourDetail
	PrivateVatican          TourDetail
	CompleteVatican         TourDetail
	AloneInTheSistineChapel TourDetail
	PristineSistineEarly    TourDetail
}{
	KeyMasterVatican: TourDetail{
		Name:      "VIP Vatican Key Master’s Tour: Unlock the Sistine Chapel",
		URL:       "https://www.walksofitaly.com/vatican-tours/key-masters-tour-sistine-chapel-vatican-museums/",
		ProductID: "e9d2d819-5f04-4b1f-a07f-612387494b8f",
	},
	PrivateVatican: TourDetail{
		Name:      "Private Vatican Tour: Vatican Museums, Sistine Chapel & St. Peter’s",
		URL:       "https://www.walksofitaly.com/vatican-tours/private-vatican-tour/",
		ProductID: "c40d8e0e-6756-463b-a052-982c77a707aa",
	},
	CompleteVatican: TourDetail{
		Name:      "The Complete Vatican Tour with Vatican Museums, Sistine Chapel & St. Peter’s Basilica",
		URL:       "https://www.walksofitaly.com/vatican-tours/complete-vatican-tour/",
		ProductID: "3b263ef8-c280-49cc-a74f-ac95aa2f1b58",
	},
	AloneInTheSistineChapel: TourDetail{
		Name:      "Alone in the Sistine Chapel: VIP Entry at the Vatican’s Most Exclusive Hours",
		URL:       "https://www.walksofitaly.com/vatican-tours/vatican-after-hours-tour/",
		ProductID: "8c14824f-905d-4273-8b83-10b567db6e55",
	},
	PristineSistineEarly: TourDetail{
		Name:      "Pristine Sistine™ Early Entrance Small Group Vatican Tour",
		URL:       "https://www.walksofitaly.com/vatican-tours/pristine-sistine-chapel-tour/",
		ProductID: "a1249220-e5d8-4983-93b2-c31ddfb3ccb8",
	},
}

type TourDetail struct {
	Name      string
	URL       string
	ProductID string
}

type AvailabilityResponse []AvailabilityDetail

type AvailabilityDetail struct {
	ID                      time.Time      `json:"id"`
	LocalDateTimeStart      time.Time      `json:"localDateTimeStart"`
	LocalDateTimeEnd        time.Time      `json:"localDateTimeEnd"`
	AllDay                  bool           `json:"allDay"`
	Available               bool           `json:"available"`
	Status                  string         `json:"status"`
	Vacancies               int            `json:"vacancies"`
	Capacity                int            `json:"capacity"`
	LimitCapacity           any            `json:"limitCapacity"`
	TotalCapacity           int            `json:"totalCapacity"`
	PaxCount                int            `json:"paxCount"`
	LimitPaxCount           int            `json:"limitPaxCount"`
	TotalPaxCount           int            `json:"totalPaxCount"`
	NoShows                 int            `json:"noShows"`
	TotalNoShows            int            `json:"totalNoShows"`
	MaxUnits                int            `json:"maxUnits"`
	MaxPaxCount             int            `json:"maxPaxCount"`
	UtcCutoffAt             time.Time      `json:"utcCutoffAt"`
	OpeningHours            []OpeningHours `json:"openingHours"`
	MeetingPoint            string         `json:"meetingPoint"`
	MeetingPointCoordinates string         `json:"meetingPointCoordinates"`
	MeetingPointLatitude    float64        `json:"meetingPointLatitude"`
	MeetingPointLongitude   float64        `json:"meetingPointLongitude"`
	MeetingLocalDateTime    time.Time      `json:"meetingLocalDateTime"`
	TourGroup               any            `json:"tourGroup"`
	Fare                    any            `json:"fare"`
	Notices                 []any          `json:"notices"`
	UnitPricing             []UnitPricing  `json:"unitPricing"`
	Offers                  []any          `json:"offers"`
	OfferCode               any            `json:"offerCode"`
	OfferTitle              any            `json:"offerTitle"`
	Offer                   any            `json:"offer"`
	Pricing                 Pricing        `json:"pricing"`
	PickupAvailable         bool           `json:"pickupAvailable"`
	PickupRequired          bool           `json:"pickupRequired"`
	PickupPoints            []any          `json:"pickupPoints"`
	HasResources            bool           `json:"hasResources"`
}

type OpeningHours struct {
	From            string `json:"from"`
	To              string `json:"to"`
	Frequency       any    `json:"frequency"`
	FrequencyAmount any    `json:"frequencyAmount"`
	FrequencyUnit   string `json:"frequencyUnit"`
}

type OfferDiscount struct {
	Retail        int   `json:"retail"`
	Net           any   `json:"net"`
	IncludedTaxes []any `json:"includedTaxes"`
}

type UnitPricing struct {
	UnitID            string        `json:"unitId"`
	UnitType          string        `json:"unitType"`
	Original          int           `json:"original"`
	Retail            int           `json:"retail"`
	Net               any           `json:"net"`
	Currency          string        `json:"currency"`
	CurrencyPrecision int           `json:"currencyPrecision"`
	IncludedTaxes     []any         `json:"includedTaxes"`
	OfferDiscount     OfferDiscount `json:"offerDiscount"`
}

type Pricing struct {
	Original          int           `json:"original"`
	Retail            int           `json:"retail"`
	Net               any           `json:"net"`
	IncludedTaxes     []any         `json:"includedTaxes"`
	OfferDiscount     OfferDiscount `json:"offerDiscount"`
	Currency          string        `json:"currency"`
	CurrencyPrecision int           `json:"currencyPrecision"`
}

type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func NewDate(year int, month time.Month, day int) Date {
	return Date{year, month, day}
}

func (d Date) Add(year int, month time.Month, day int) Date {
	d.Year += year
	d.Month += month
	d.Day += day
	return d
}

func (d Date) ToTime() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}

func FromTime(t time.Time) Date {
	return Date{
		Year:  t.Year(),
		Month: t.Month(),
		Day:   t.Day(),
	}
}

func (d Date) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`"%04d-%02d-%02d"`, d.Year, d.Month, d.Day)
	return []byte(s), nil
}

// String implements the Stringer interface for pretty printing
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

func main() {
	start := NewDate(2025, time.September, 1)
	tours.PristineSistineEarly.GetAvailability(start, start.Add(0, 0, 3))
}

type AvailabilityRequestBody struct {
	ProductID      string `json:"productId"`
	OptionID       string `json:"optionId"`
	LocalDateStart Date   `json:"localDateStart"`
	LocalDateEnd   Date   `json:"localDateEnd"`
	Currency       string `json:"currency"`
}

func NewAvailabilityRequestBody(productID string, start, end Date) AvailabilityRequestBody {
	return AvailabilityRequestBody{
		ProductID:      productID,
		OptionID:       "DEFAULT",
		LocalDateStart: start,
		LocalDateEnd:   end,
		Currency:       "USD",
	}
}

func (ar AvailabilityRequestBody) JSON() io.Reader {
	var r bytes.Buffer
	_ = json.NewEncoder(&r).Encode(ar)
	return &r
}

func (td TourDetail) GetAvailability(start, end Date) {
	requestBody := NewAvailabilityRequestBody(td.ProductID, start, end)
	req, err := http.NewRequest(http.MethodPost, availabilityURL, requestBody.JSON())
	if err != nil {
		panic(err)
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
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}
