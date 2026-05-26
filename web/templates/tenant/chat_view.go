package tenantviews

import (
	"fmt"

	"dorm-man/internal/models"

	"github.com/google/uuid"
)

func ChatRoomURL(roomID uuid.UUID) string {
	return chatRoomURL(roomID)
}

func chatRoomURL(roomID uuid.UUID) string {
	return "/tenant/chat?room=" + roomID.String()
}

func chatPanelURL(roomID uuid.UUID) string {
	return "/tenant/chat/panel?room=" + roomID.String()
}

func chatMessagesPollURL(roomID uuid.UUID) string {
	return "/tenant/chat/messages?room=" + roomID.String()
}

func chatSidebarURL(query string, activeRoomID uuid.UUID) string {
	url := "/tenant/chat/sidebar"
	if query != "" {
		url += "?q=" + query
	}
	if activeRoomID != uuid.Nil {
		if query != "" {
			url += "&"
		} else {
			url += "?"
		}
		url += "room=" + activeRoomID.String()
	}
	return url
}

func activeRoomValue(activeRoomID uuid.UUID) string {
	if activeRoomID == uuid.Nil {
		return ""
	}
	return activeRoomID.String()
}

func tenantDirectChatVals(tenantID uuid.UUID) string {
	return fmt.Sprintf(`{"tenant_id": "%s"}`, tenantID.String())
}

func tenantDisplayName(tenant models.Tenant) string {
	if tenant.User != nil && tenant.User.Name != "" {
		return tenant.User.Name
	}
	return "Tenant"
}

func chatRoomDisplayTitle(room models.ChatRoom, displayTitle string) string {
	if displayTitle != "" {
		return displayTitle
	}
	return room.Title
}

func messageSenderLabel(message models.Message) string {
	if message.Sender != nil && message.Sender.User != nil && message.Sender.User.Name != "" {
		return message.Sender.User.Name
	}
	if message.Sender != nil && message.Sender.StudentCode != "" {
		return message.Sender.StudentCode
	}
	return "Tenant"
}

func roomKindLabel(kind models.RoomKind) string {
	switch kind {
	case models.RoomKindDirect:
		return "Direct"
	case models.RoomKindGroup:
		return "Group"
	case models.RoomKindEvent:
		return "Event"
	default:
		return string(kind)
	}
}
