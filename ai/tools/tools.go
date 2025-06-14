package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"

	"walks-of-italy/storage"
	"walks-of-italy/tours"

	"github.com/mitchellh/mapstructure"
	"github.com/ollama/ollama/api"
)

type Tools struct {
	sc            *storage.Client
	ventrataToken string
	walksToken    string
	logger        slog.Logger
	cache         map[string]string
}

func New(sc *storage.Client, ventrataToken, walksToken string, logger slog.Logger) Tools {
	return Tools{
		sc:            sc,
		ventrataToken: ventrataToken,
		walksToken:    walksToken,
		logger:        logger,
		cache:         map[string]string{},
	}
}

func executeToolFunction[T interface{ CacheKey() string }](cache map[string]string, args map[string]any, runTool func(T) (string, error)) (string, error) {
	var input T
	dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.TextUnmarshallerHookFunc(),
		Result:     &input,
	})

	err := dec.Decode(args)
	if err != nil {
		return "", fmt.Errorf("error decoding input: %w", err)
	}

	cacheVal, ok := cache[input.CacheKey()]
	if ok {
		return cacheVal, nil
	}

	output, err := runTool(input)
	if err != nil {
		return "", err
	}

	cache[input.CacheKey()] = output

	return output, nil
}

func (t Tools) Execute(name string, args map[string]any) (output string, _ error) {
	t.logger.With("name", name, "args", args).Debug("tool call")

	defer func(out *string) {
		if out == nil {
			return
		}
		t.logger.With("name", name, "args", args, "output", *out).Debug("tool call done")
	}(&output)

	switch name {
	case "getAllTours":
		return executeToolFunction(t.cache, args, t.GetAllTours)
	case "getTourDetails":
		return executeToolFunction(t.cache, args, t.GetTourDetails)
	case "getTourAvailability":
		return executeToolFunction(t.cache, args, t.GetAvailability)
	default:
		return "", fmt.Errorf("unknown function: %q", name)
	}
}

type GetAllToursInput struct{}

func (GetAllToursInput) CacheKey() string {
	return "getAllTours"
}

func (t Tools) GetAllTours(GetAllToursInput) (string, error) {
	allTours, err := t.sc.GetAll(context.Background(), url.Values{})
	if err != nil {
		return "", fmt.Errorf("error getting tours: %w", err)
	}

	type tourNameID struct {
		Name string
		ID   string
	}

	results := []tourNameID{}
	for _, td := range allTours {
		results = append(results, tourNameID{td.Name, td.ProductID.String()})
	}

	out, err := json.Marshal(results)
	return string(out), err
}

type GetTourDetailsInput struct {
	TourID string `mapstructure:"tour_id"`
}

func (g GetTourDetailsInput) CacheKey() string {
	return "getTourDetails_" + g.TourID
}

func (t Tools) GetTourDetails(in GetTourDetailsInput) (string, error) {
	tour, err := t.sc.Get(context.Background(), in.TourID)
	if err != nil {
		return "", fmt.Errorf("error getting tour: %w", err)
	}

	desc, err := tour.GetDescription(context.Background(), tour.ApiUrl, t.walksToken)
	if err != nil {
		return "", fmt.Errorf("error getting description for %q: %w", tour.Name, err)
	}

	out, err := json.Marshal(desc)
	return string(out), err
}

type GetAvailabilityInput struct {
	TourID string     `mapstructure:"tour_id"`
	Start  tours.Date `mapstructure:"start"`
	End    tours.Date `mapstructure:"end"`
}

func (g GetAvailabilityInput) CacheKey() string {
	return fmt.Sprintf("getTourDetails_%s_%s_%s", g.TourID, g.Start.String(), g.End.String())
}

func (t Tools) GetAvailability(in GetAvailabilityInput) (string, error) {
	tour, err := t.sc.Get(context.Background(), in.TourID)
	if err != nil {
		return "", fmt.Errorf("error getting tour: %w", err)
	}

	avail, err := tour.GetAvailability(context.Background(), t.ventrataToken, in.Start, in.End)
	if err != nil {
		return "", fmt.Errorf("error getting description for %q: %w", tour.Name, err)
	}

	output, err := json.Marshal(map[string]any{
		"availability": avail,
		"instruction":  "tell the user about availability. do not describe the json structure.",
	})
	if err != nil {
		return "", err
	}

	return string(output), err
}

func (t Tools) Tools() api.Tools {
	return api.Tools{
		{
			Type: "function",
			Function: api.ToolFunction{
				Name:        "getAllTours",
				Description: "Get the name and ID of every walks-of-italy tours",
			},
		},
		{
			Type: "function",
			Function: api.ToolFunction{
				Name:        "getTourDetails",
				Description: "Get more specific details about a tour",
				Parameters: ToolFunctionParameters{
					Type:     "object",
					Required: []string{"tour_id"},
					Properties: ToolFunctionProperties{
						"tour_id": {
							Type:        api.PropertyType{"string"},
							Description: "The UUID for identifying a tour",
						},
					},
				}.ToAPI(),
			},
		},
		{
			Type: "function",
			Function: api.ToolFunction{
				Name:        "getTourAvailability",
				Description: "Get a tour's availability for certain dates",
				Parameters: ToolFunctionParameters{
					Type:     "object",
					Required: []string{"tour_id", "start", "end"},
					Properties: ToolFunctionProperties{
						"tour_id": {
							Type:        api.PropertyType{"string"},
							Description: "The UUID for identifying a tour",
						},
						"start": {
							Type:        api.PropertyType{"date"},
							Description: "The date to start the search in format 2006-01-02",
						},
						"end": {
							Type:        api.PropertyType{"date"},
							Description: "The date to start the end in format 2006-01-02",
						},
					},
				}.ToAPI(),
			},
		},
	}
}

type ToolFunctionParameters struct {
	Type       string                 `json:"type"`
	Defs       any                    `json:"$defs,omitempty"`
	Items      any                    `json:"items,omitempty"`
	Required   []string               `json:"required"`
	Properties ToolFunctionProperties `json:"properties"`
}

func (t ToolFunctionParameters) ToAPI() struct {
	Type       string   `json:"type"`
	Defs       any      `json:"$defs,omitempty"`
	Items      any      `json:"items,omitempty"`
	Required   []string `json:"required"`
	Properties map[string]struct {
		Type        api.PropertyType `json:"type"`
		Items       any              `json:"items,omitempty"`
		Description string           `json:"description"`
		Enum        []any            `json:"enum,omitempty"`
	} `json:"properties"`
} {
	return struct {
		Type       string   `json:"type"`
		Defs       any      `json:"$defs,omitempty"`
		Items      any      `json:"items,omitempty"`
		Required   []string `json:"required"`
		Properties map[string]struct {
			Type        api.PropertyType `json:"type"`
			Items       any              `json:"items,omitempty"`
			Description string           `json:"description"`
			Enum        []any            `json:"enum,omitempty"`
		} `json:"properties"`
	}{
		Type:       t.Type,
		Defs:       t.Defs,
		Items:      t.Items,
		Required:   t.Required,
		Properties: t.Properties.ToAPI(),
	}
}

type ToolFunctionProperties map[string]ToolFunctionProperty

type ToolFunctionProperty struct {
	Type        api.PropertyType `json:"type"`
	Items       any              `json:"items,omitempty"`
	Description string           `json:"description"`
	Enum        []any            `json:"enum,omitempty"`
}

func (t ToolFunctionProperties) ToAPI() map[string]struct {
	Type        api.PropertyType `json:"type"`
	Items       any              `json:"items,omitempty"`
	Description string           `json:"description"`
	Enum        []any            `json:"enum,omitempty"`
} {
	result := map[string]struct {
		Type        api.PropertyType `json:"type"`
		Items       any              `json:"items,omitempty"`
		Description string           `json:"description"`
		Enum        []any            `json:"enum,omitempty"`
	}{}

	for key, val := range t {
		result[key] = val
	}

	return result
}
