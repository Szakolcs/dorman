package maintenance

import (
	"net/url"
	"time"

	"dorm-man/internal/models"

	"github.com/labstack/echo/v4"
)

func ParseTicketFilter(c echo.Context) (TicketFilter, error) {
	return ParseTicketFilterQuery(c.QueryParams())
}

func ParseTicketFilterQuery(q url.Values) (TicketFilter, error) {
	filter := TicketFilter{}
	for _, v := range q["status"] {
		if v != "" {
			filter.Statuses = append(filter.Statuses, models.Status(v))
		}
	}
	for _, v := range q["category"] {
		if v != "" {
			filter.Categories = append(filter.Categories, models.Category(v))
		}
	}
	for _, v := range q["severity"] {
		if v != "" {
			filter.Severities = append(filter.Severities, models.Severity(v))
		}
	}
	if from := q.Get("from"); from != "" {
		t, err := time.Parse(time.DateOnly, from)
		if err != nil {
			return TicketFilter{}, ErrValidation
		}
		filter.From = t
	}
	if to := q.Get("to"); to != "" {
		t, err := time.Parse(time.DateOnly, to)
		if err != nil {
			return TicketFilter{}, ErrValidation
		}
		filter.To = t
	}
	return filter, nil
}
