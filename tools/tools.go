package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"walks-of-italy/storage"
	"walks-of-italy/tours"

	"github.com/ollama/ollama/api"
)

type Tools struct {
	sc          *storage.Client
	accessToken string
	logger      slog.Logger
	cache       map[string]string
}

func New(sc *storage.Client, accessToken string) Tools {
	return Tools{sc: sc, accessToken: accessToken, logger: *slog.Default(), cache: map[string]string{}}
}

func (t Tools) Execute(name string, args map[string]any) (string, error) {
	t.logger.With("name", name, "args", args).Info("tool call")
	switch name {
	case "getAllTours":
		cacheVal, ok := t.cache["getAllTours"]
		if ok {
			return cacheVal, nil
		}

		output, err := t.GetAllTours()
		if err != nil {
			return "", err
		}

		t.cache["getAllTours"] = output

		return output, nil
	case "getTourDetails":
		id, ok := args["tour_id"].(string)
		if !ok {
			return "", fmt.Errorf("missing tour_id argument: %v", args)
		}

		cacheKey := "getTourDetails_" + id

		cacheVal, ok := t.cache[cacheKey]
		if ok {
			return cacheVal, nil
		}

		output, err := t.GetTourDetails(id)
		if err != nil {
			return "", fmt.Errorf("error getting tour details: %w", err)
		}

		t.cache[cacheKey] = output

		return output, nil
	case "getTourAvailability":
		id, ok := args["tour_id"].(string)
		if !ok {
			return "", fmt.Errorf("missing tour_id argument: %v", args)
		}
		start, ok := args["start"].(string)
		if !ok {
			return "", fmt.Errorf("missing start argument: %v", args)
		}
		end, ok := args["end"].(string)
		if !ok {
			return "", fmt.Errorf("missing end argument: %v", args)
		}

		cacheKey := fmt.Sprintf("getTourDetails_%s_%s_%s", id, start, end)
		cacheVal, ok := t.cache[cacheKey]
		if ok {
			return cacheVal, nil
		}

		startTime, err := time.Parse(time.DateOnly, start)
		if err != nil {
			return "", fmt.Errorf("error parsing start: %w", err)
		}
		endTime, err := time.Parse(time.DateOnly, end)
		if err != nil {
			return "", fmt.Errorf("error parsing end: %w", err)
		}

		output, err := t.GetAvailability(id, tours.DateFromTime(startTime), tours.DateFromTime(endTime))
		if err != nil {
			return "", fmt.Errorf("error getting tour details: %w", err)
		}

		t.cache[cacheKey] = output

		t.logger.With("name", name, "output", string(output)).Info("success")

		return output, nil
	default:
		return "", fmt.Errorf("unknown function: %q", name)
	}
}

func (t Tools) GetAllTours() (string, error) {
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

func (t Tools) GetTourDetails(id string) (string, error) {
	tour, err := t.sc.Get(context.Background(), id)
	if err != nil {
		return "", fmt.Errorf("error getting tour: %w", err)
	}

	desc, err := tour.GetDescription(context.Background(), tour.ApiUrl)
	if err != nil {
		return "", fmt.Errorf("error getting description for %q: %w", tour.Name, err)
	}

	return string(desc), err
}

func (t Tools) GetAvailability(id string, start, end tours.Date) (string, error) {
	tour, err := t.sc.Get(context.Background(), id)
	if err != nil {
		return "", fmt.Errorf("error getting tour: %w", err)
	}

	avail, err := tour.GetAvailability(context.Background(), t.accessToken, start, end)
	if err != nil {
		return "", fmt.Errorf("error getting description for %q: %w", tour.Name, err)
	}

	output, err := json.Marshal(avail)
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
				Parameters: api.ToolFunctionParameters{
					Type:     "object",
					Required: []string{"tour_id"},
					Properties: map[string]api.ToolFunctionProperty{
						"tour_id": {
							Type:        api.PropertyType{"string"},
							Description: "The UUID for identifying a tour",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: api.ToolFunction{
				Name:        "getTourAvailability",
				Description: "Get a tour's availability for certain dates",
				Parameters: api.ToolFunctionParameters{
					Type:     "object",
					Required: []string{"tour_id", "start", "end"},
					Properties: map[string]api.ToolFunctionProperty{
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
				},
			},
		},
	}
}
