package tours

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (td TourDetail) GetDescription(ctx context.Context, url, accessToken string) (Description, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return Description{}, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Description{}, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Description{}, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Description{}, fmt.Errorf("unexpected response code: %d, body: %q", resp.StatusCode, string(body))
	}

	var result Description
	err = json.Unmarshal(body, &result)
	if err != nil {
		return Description{}, fmt.Errorf("error parsing response: %w", err)
	}

	return result, nil
}

type Description struct {
	EventID                 string       `json:"eventId"`
	CitySlug                string       `json:"citySlug"`
	CityName                string       `json:"cityName"`
	CitySlugs               []string     `json:"citySlugs"`
	ShortTitle              string       `json:"shortTitle"`
	TourPageURL             string       `json:"tourPageUrl"`
	BookingTypeID           int          `json:"bookingTypeId"`
	PropertyID              string       `json:"propertyId"`
	TourPageMetaTitle       string       `json:"tourPageMetaTitle"`
	TourPageMetaDescription string       `json:"tourPageMetaDescription"`
	TourStartTime           string       `json:"tourStartTime"`
	MaxGroupSize            string       `json:"maxGroupSize"`
	Duration                string       `json:"duration"`
	Flag                    string       `json:"flag"`
	ListingText             string       `json:"listingText"`
	Highlights              string       `json:"highlights"`
	Description             string       `json:"description"`
	Name                    string       `json:"name"`
	Title                   string       `json:"title"`
	TourIncludes            []string     `json:"tourIncludes"`
	SitesVisited            []string     `json:"sitesVisited"`
	ListingImageTitle       string       `json:"listingImageTitle"`
	TourImportantInfo       string       `json:"tourImportantInfo"`
	Promo                   any          `json:"promo"`
	Awards                  []any        `json:"awards"`
	ReviewStatus            ReviewStatus `json:"reviewStatus"`
	Product                 Product      `json:"product"`
	PriceMap                PriceMap     `json:"priceMap"`
}

type StarsGroup struct {
	Num0 int `json:"0"`
	Num1 int `json:"1"`
	Num2 int `json:"2"`
	Num3 int `json:"3"`
	Num4 int `json:"4"`
	Num5 int `json:"5"`
}

type ReviewStatus struct {
	FeedbackAverage       float64    `json:"feedbackAverage"`
	FeedbackCount         int        `json:"feedbackCount"`
	FeedbackCount0        int        `json:"feedback_count"`
	FeedbackAverage0      float64    `json:"feedback_average"`
	StarsGroup            StarsGroup `json:"starsGroup"`
	ThirdPartyTotalReview int        `json:"thirdPartyTotalReview"`
}

type Restrictions struct {
	MinUnits    int `json:"minUnits"`
	MaxUnits    any `json:"maxUnits"`
	MinPaxCount int `json:"minPaxCount"`
	MaxPaxCount any `json:"maxPaxCount"`
}

type Options struct {
	ID                          string       `json:"id"`
	Default                     bool         `json:"default"`
	InternalName                string       `json:"internalName"`
	Reference                   any          `json:"reference"`
	Tags                        []any        `json:"tags"`
	AvailabilityLocalStartTimes []string     `json:"availabilityLocalStartTimes"`
	AvailabilityLocalDateStart  string       `json:"availabilityLocalDateStart"`
	AvailabilityLocalDateEnd    any          `json:"availabilityLocalDateEnd"`
	CancellationCutoff          string       `json:"cancellationCutoff"`
	CancellationCutoffAmount    int          `json:"cancellationCutoffAmount"`
	CancellationCutoffUnit      string       `json:"cancellationCutoffUnit"`
	AvailabilityCutoff          string       `json:"availabilityCutoff"`
	AvailabilityCutoffAmount    int          `json:"availabilityCutoffAmount"`
	AvailabilityCutoffUnit      string       `json:"availabilityCutoffUnit"`
	VisibleContactFields        []string     `json:"visibleContactFields"`
	RequiredContactFields       []string     `json:"requiredContactFields"`
	Restrictions                Restrictions `json:"restrictions"`
	Title                       string       `json:"title"`
	Subtitle                    any          `json:"subtitle"`
	Language                    string       `json:"language"`
	ShortDescription            string       `json:"shortDescription"`
	Duration                    string       `json:"duration"`
	DurationAmount              int          `json:"durationAmount"`
	DurationUnit                string       `json:"durationUnit"`
	CoverImageURL               string       `json:"coverImageUrl"`
	Itinerary                   any          `json:"itinerary"`
	FromPoint                   any          `json:"fromPoint"`
	ToPoint                     any          `json:"toPoint"`
}

type Contact struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Telephone any    `json:"telephone"`
	Address   any    `json:"address"`
	Website   string `json:"website"`
}

type Brand struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	BackgroundColor string  `json:"backgroundColor"`
	CheckoutLogoURL string  `json:"checkoutLogoUrl"`
	Color           string  `json:"color"`
	SecondaryColor  string  `json:"secondaryColor"`
	FaviconURL      any     `json:"faviconUrl"`
	LogoURL         string  `json:"logoUrl"`
	LogoWhiteURL    any     `json:"logoWhiteUrl"`
	AccentFont      any     `json:"accentFont"`
	BodyFont        any     `json:"bodyFont"`
	HeaderFont      any     `json:"headerFont"`
	Contact         Contact `json:"contact"`
}

type Destination struct {
	ID                  string   `json:"id"`
	Default             bool     `json:"default"`
	Name                string   `json:"name"`
	Title               string   `json:"title"`
	ShortDescription    any      `json:"shortDescription"`
	Featured            bool     `json:"featured"`
	Tags                []any    `json:"tags"`
	Country             string   `json:"country"`
	Contact             Contact  `json:"contact"`
	Brand               Brand    `json:"brand"`
	Address             any      `json:"address"`
	GooglePlaceID       any      `json:"googlePlaceId"`
	Latitude            any      `json:"latitude"`
	Longitude           any      `json:"longitude"`
	CoverImageURL       string   `json:"coverImageUrl"`
	BannerImageURL      any      `json:"bannerImageUrl"`
	VideoURL            any      `json:"videoUrl"`
	FacebookURL         any      `json:"facebookUrl"`
	GoogleURL           any      `json:"googleUrl"`
	TripadvisorURL      any      `json:"tripadvisorUrl"`
	TwitterURL          any      `json:"twitterUrl"`
	YoutubeURL          any      `json:"youtubeUrl"`
	InstagramURL        any      `json:"instagramUrl"`
	Notices             []any    `json:"notices"`
	DefaultCurrency     string   `json:"defaultCurrency"`
	AvailableCurrencies []string `json:"availableCurrencies"`
}

type Product struct {
	ID                                 string      `json:"id"`
	Tags                               []string    `json:"tags"`
	Locale                             string      `json:"locale"`
	TimeZone                           string      `json:"timeZone"`
	AllowFreesale                      bool        `json:"allowFreesale"`
	FreesaleDurationAmount             int         `json:"freesaleDurationAmount"`
	FreesaleDurationUnit               string      `json:"freesaleDurationUnit"`
	InstantConfirmation                bool        `json:"instantConfirmation"`
	InstantDelivery                    bool        `json:"instantDelivery"`
	AvailabilityRequired               bool        `json:"availabilityRequired"`
	Options                            []Options   `json:"options"`
	Title                              string      `json:"title"`
	Country                            string      `json:"country"`
	Location                           string      `json:"location"`
	Address                            any         `json:"address"`
	GooglePlaceID                      any         `json:"googlePlaceId"`
	Latitude                           any         `json:"latitude"`
	Longitude                          any         `json:"longitude"`
	Subtitle                           any         `json:"subtitle"`
	Tagline                            any         `json:"tagline"`
	Keywords                           []any       `json:"keywords"`
	PointToPoint                       bool        `json:"pointToPoint"`
	ShortDescription                   string      `json:"shortDescription"`
	Description                        any         `json:"description"`
	Highlights                         []string    `json:"highlights"`
	Alert                              any         `json:"alert"`
	Inclusions                         []string    `json:"inclusions"`
	Exclusions                         []string    `json:"exclusions"`
	BookingTerms                       any         `json:"bookingTerms"`
	PrivacyTerms                       any         `json:"privacyTerms"`
	RedemptionInstructions             any         `json:"redemptionInstructions"`
	CancellationPolicy                 string      `json:"cancellationPolicy"`
	Faqs                               []any       `json:"faqs"`
	Destination                        Destination `json:"destination"`
	Categories                         []any       `json:"categories"`
	DefaultCurrency                    string      `json:"defaultCurrency"`
	AvailableCurrencies                []string    `json:"availableCurrencies"`
	IncludeTax                         bool        `json:"includeTax"`
	PricingPer                         string      `json:"pricingPer"`
	OutstandingBalanceTitle            any         `json:"outstandingBalanceTitle"`
	OutstandingBalanceShortDescription any         `json:"outstandingBalanceShortDescription"`
}

type Currencies struct {
	ID            string  `json:"id"`
	Rate          float64 `json:"rate"`
	Label         string  `json:"label"`
	CurrencyID    int     `json:"currencyId"`
	IsGeoCurrency bool    `json:"isGeoCurrency"`
	IsDefault     bool    `json:"isDefault"`
}

type VentrataCurrencies struct {
	Original          int    `json:"original"`
	Retail            int    `json:"retail"`
	Net               any    `json:"net"`
	Currency          string `json:"currency"`
	CurrencyPrecision int    `json:"currencyPrecision"`
	IncludedTaxes     []any  `json:"includedTaxes"`
}

type PriceMap struct {
	AnchorPrice        int                  `json:"anchorPrice"`
	VentrataPrice      int                  `json:"ventrataPrice"`
	DiscountedPrice    int                  `json:"discountedPrice"`
	Currencies         []Currencies         `json:"currencies"`
	VentrataCurrencies []VentrataCurrencies `json:"ventrataCurrencies"`
}
