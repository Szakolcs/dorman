package seed

import "gorm.io/gorm"

// TruncateAll removes seed data in dependency-safe order (PostgreSQL CASCADE).
func TruncateAll(db *gorm.DB) error {
	sql := `
TRUNCATE TABLE
	forum_moderation_actions,
	forum_post_views,
	forum_reactions,
	forum_attendance_intents,
	forum_poll_votes,
	forum_poll_options,
	forum_polls,
	forum_comments,
	forum_post_updates,
	forum_post_schedules,
	forum_posts,
	chat_membership_sync_logs,
	chat_room_members,
	chat_messages,
	chat_tenant_profiles,
	chat_rooms,
	access_events,
	guest_access_events,
	guest_visits,
	package_notifications,
	packages,
	item_loans,
	tenant_entry_tokens,
	ticket_status_changes,
	maintenance_tickets,
	operational_jobs,
	inventory_items,
	events,
	activities,
	room_assignments,
	tenants,
	rooms,
	flats,
	user_roles,
	users,
	roles,
	buildings,
	audit_events
RESTART IDENTITY CASCADE`
	return db.Exec(sql).Error
}
