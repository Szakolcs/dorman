package adminviews

import (
	"fmt"

	"dorm-man/internal/models"
)

func inventoryStatusFilterURL(status string, activeStatus string) string {
	if activeStatus == status {
		return "/administration/inventory"
	}
	return "/administration/inventory?status=" + status
}

func inventoryStatusLabel(status models.InventoryStatus) string {
	switch status {
	case models.InventoryStatusInStock:
		return "In stock"
	case models.InventoryStatusInUse:
		return "In use"
	case models.InventoryStatusWithdrawn:
		return "Withdrawn"
	case models.InventoryStatusDestroyed:
		return "Destroyed"
	default:
		return string(status)
	}
}

func inventoryConditionLabel(condition models.InventoryCondition) string {
	switch condition {
	case models.InventoryConditionNew:
		return "New"
	case models.InventoryConditionGood:
		return "Good"
	case models.InventoryConditionUsed:
		return "Used"
	case models.InventoryConditionDamaged:
		return "Damaged"
	case models.InventoryConditionBroken:
		return "Broken"
	default:
		return string(condition)
	}
}

func inventoryStatusClass(status models.InventoryStatus) string {
	switch status {
	case models.InventoryStatusInStock:
		return "dm-inventory-status-in-stock"
	case models.InventoryStatusInUse:
		return "dm-inventory-status-in-use"
	case models.InventoryStatusWithdrawn:
		return "dm-inventory-status-withdrawn"
	case models.InventoryStatusDestroyed:
		return "dm-inventory-status-destroyed"
	default:
		return ""
	}
}

func inventoryConditionClass(condition models.InventoryCondition) string {
	switch condition {
	case models.InventoryConditionNew:
		return "dm-inventory-condition-new"
	case models.InventoryConditionGood:
		return "dm-inventory-condition-good"
	case models.InventoryConditionUsed:
		return "dm-inventory-condition-used"
	case models.InventoryConditionDamaged:
		return "dm-inventory-condition-damaged"
	case models.InventoryConditionBroken:
		return "dm-inventory-condition-broken"
	default:
		return ""
	}
}

func inventoryLocationLabel(item models.InventoryItem) string {
	switch {
	case item.Room != nil && item.Room.Number != "":
		if item.Flat != nil && item.Flat.Name != "" {
			return fmt.Sprintf("Room %s · %s", item.Room.Number, item.Flat.Name)
		}
		return "Room " + item.Room.Number
	case item.Flat != nil && item.Flat.Name != "":
		return item.Flat.Name
	case item.Building != nil && item.Building.Name != "":
		return item.Building.Name
	case item.SharedArea != nil && item.SharedArea.Name != "":
		return item.SharedArea.Name
	default:
		return ""
	}
}
