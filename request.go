package tours

import (
	"time"

	"github.com/google/uuid"
)

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
