package config

import "gorm.io/gorm"

func CreateAuditInfrastructure(db *gorm.DB) error {
	sql := `
CREATE TABLE IF NOT EXISTS audits (
    table_name TEXT NOT NULL,
    operation TEXT NOT NULL,
    old_data JSONB,
    new_data JSONB,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changed_by UUID
);

CREATE OR REPLACE FUNCTION audit_trigger()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    current_app_user UUID;
BEGIN
    BEGIN
        current_app_user :=
            current_setting('app.current_app_user', true)::UUID;
    EXCEPTION
        WHEN OTHERS THEN
            current_app_user := NULL;
    END;

    INSERT INTO audits (
        table_name,
        operation,
        old_data,
        new_data,
        changed_at,
        changed_by
    )
    VALUES (
        TG_TABLE_NAME,
        TG_OP,
        to_jsonb(OLD),
        to_jsonb(NEW),
        CURRENT_TIMESTAMP,
        current_app_user
    );

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS buildings_audit_trigger ON buildings;
CREATE TRIGGER buildings_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON buildings
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS flats_audit_trigger ON flats;
CREATE TRIGGER flats_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON flats
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS rooms_audit_trigger ON rooms;
CREATE TRIGGER rooms_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON rooms
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS shared_areas_audit_trigger ON shared_areas;
CREATE TRIGGER shared_areas_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON shared_areas
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS tenants_audit_trigger ON tenants;
CREATE TRIGGER tenants_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON tenants
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS room_assignments_audit_trigger ON room_assignments;
CREATE TRIGGER room_assignments_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON room_assignments
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS inventory_items_audit_trigger ON inventory_items;
CREATE TRIGGER inventory_items_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON inventory_items
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS operational_jobs_audit_trigger ON operational_jobs;
CREATE TRIGGER operational_jobs_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON operational_jobs
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS users_audit_trigger ON users;
CREATE TRIGGER users_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON users
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS roles_audit_trigger ON roles;
CREATE TRIGGER roles_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON roles
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS permissions_audit_trigger ON permissions;
CREATE TRIGGER permissions_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON permissions
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS operations_audit_trigger ON operations;
CREATE TRIGGER operations_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON operations
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS publications_audit_trigger ON publications;
CREATE TRIGGER publications_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON publications
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS activities_audit_trigger ON activities;
CREATE TRIGGER activities_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON activities
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS events_audit_trigger ON events;
CREATE TRIGGER events_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON events
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS tickets_audit_trigger ON tickets;
CREATE TRIGGER tickets_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON tickets
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

DROP TRIGGER IF EXISTS status_changes_audit_trigger ON status_changes;
CREATE TRIGGER status_changes_audit_trigger
AFTER INSERT OR UPDATE OR DELETE
ON status_changes
FOR EACH ROW
EXECUTE FUNCTION audit_trigger();
`

	return db.Exec(sql).Error
}
