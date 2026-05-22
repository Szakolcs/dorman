package maintenance

import (
	"time"

	"dorm-man/internal/models"

	"github.com/labstack/echo/v4"
)

func ParseTicketFilter(c echo.Context) (TicketFilter, error) {
	filter := TicketFilter{}
	for _, v := range c.QueryParams()["status"] {
		if v != "" {
			filter.Statuses = append(filter.Statuses, models.Status(v))
		}
	}
	for _, v := range c.QueryParams()["category"] {
		if v != "" {
			filter.Categories = append(filter.Categories, models.Category(v))
		}
	}
	for _, v := range c.QueryParams()["severity"] {
		if v != "" {
			filter.Severities = append(filter.Severities, models.Severity(v))
		}
	}
	if from := c.QueryParam("from"); from != "" {
		t, err := time.Parse(time.DateOnly, from)
		if err != nil {
			return TicketFilter{}, ErrValidation
		}
		filter.From = t
	}
	if to := c.QueryParam("to"); to != "" {
		t, err := time.Parse(time.DateOnly, to)
		if err != nil {
			return TicketFilter{}, ErrValidation
		}
		filter.To = t
	}
	return filter, nil
}
