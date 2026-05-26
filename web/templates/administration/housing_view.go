package adminviews

import (
	"fmt"

	"dorm-man/internal/models"

	"github.com/google/uuid"
)

func buildingDetailURL(id uuid.UUID) string {
	return "/administration/housing/building/" + id.String()
}

func flatBackURL(flat models.Flat) string {
	if flat.BuildingID != nil {
		return buildingDetailURL(*flat.BuildingID)
	}
	return "/administration/housing"
}

func flatBackLabel(flat models.Flat) string {
	if flat.Building != nil {
		return "Back to " + flat.Building.Name
	}
	return "Back to housing"
}

func sharedAreaBackURL(area models.SharedArea) string {
	if area.BuildingID != nil {
		return buildingDetailURL(*area.BuildingID)
	}
	return "/administration/housing"
}

func sharedAreaBackLabel(area models.SharedArea) string {
	if area.Building != nil {
		return "Back to " + area.Building.Name
	}
	return "Back to housing"
}

func roomOccupancyLabel(room models.Room) string {
	return fmt.Sprintf("%d / %d occupied", len(room.Assignments), room.Capacity)
}

func roomBackURL(room models.Room) string {
	return "/administration/housing/flat/" + room.FlatID.String()
}

func roomBackLabel(room models.Room) string {
	if room.Flat.ID != uuid.Nil {
		return "Back to " + room.Flat.Name
	}
	return "Back to flat"
}
