package seed

import (
	plat "dorm-man/internal/models/cross-cutting"
	"errors"
	"fmt"
	"time"

	adm "dorm-man/internal/models/administration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DevAccount describes a fixed local-development login inserted after SQL seed.
type DevAccount struct {
	Email   string
	Name    string
	UniCode string
	Roles   []adm.RoleName
}

// DevAccounts are always available after seeding (password: seedPassword).
var DevAccounts = []DevAccount{
	{
		Email:   DevAdminEmail,
		Name:    "Admin User",
		UniCode: "ADMIN001",
		Roles: []adm.RoleName{
			adm.RoleAdministrator,
			adm.RoleDoorman,
			adm.RoleOfficeWorker,
			adm.RoleDirector,
		},
	},
	{Email: "staff1@dorm.local", Name: "Staff Administrator", UniCode: "STF0001", Roles: []adm.RoleName{adm.RoleAdministrator}},
	{Email: "staff2@dorm.local", Name: "Staff Director", UniCode: "STF0002", Roles: []adm.RoleName{adm.RoleDirector}},
	{Email: "staff3@dorm.local", Name: "Staff Office", UniCode: "STF0003", Roles: []adm.RoleName{adm.RoleOfficeWorker}},
	{Email: "staff4@dorm.local", Name: "Staff Doorman", UniCode: "STF0004", Roles: []adm.RoleName{adm.RoleDoorman}},
}

// NormalizeRoleNames maps fabricate role slugs to application role names.
func NormalizeRoleNames(db *gorm.DB) error {
	mappings := []struct{ from, to string }{
		{"admin", string(adm.RoleAdministrator)},
		{"manager", string(adm.RoleDirector)},
		{"reception", string(adm.RoleOfficeWorker)},
		{"security", string(adm.RoleDoorman)},
	}
	for _, m := range mappings {
		if err := db.Exec(`UPDATE roles SET name = ? WHERE name = ?`, m.to, m.from).Error; err != nil {
			return fmt.Errorf("normalize role %q: %w", m.from, err)
		}
	}
	return nil
}

// EnsureDevAccounts upserts development users with plaintext password (login scaffold).
func EnsureDevAccounts(db *gorm.DB) error {
	roleIDs, err := loadRoleIDs(db)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, acct := range DevAccounts {
		user, err := upsertDevUser(db, acct, now)
		if err != nil {
			return err
		}
		if err := ensureUserRoles(db, user.ID, acct.Roles, roleIDs, now); err != nil {
			return fmt.Errorf("roles for %s: %w", acct.Email, err)
		}
	}
	return nil
}

func loadRoleIDs(db *gorm.DB) (map[adm.RoleName]uuid.UUID, error) {
	var roles []adm.Role
	if err := db.Find(&roles).Error; err != nil {
		return nil, err
	}
	out := make(map[adm.RoleName]uuid.UUID, len(roles))
	for _, r := range roles {
		out[r.Name] = r.ID
	}
	return out, nil
}

func upsertDevUser(db *gorm.DB, acct DevAccount, now time.Time) (adm.User, error) {
	var user adm.User
	err := db.Where("email = ?", acct.Email).First(&user).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		user = adm.User{
			BaseModel:     plat.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now},
			UniCode:       acct.UniCode,
			Email:         acct.Email,
			PasswordHash:  seedPassword,
			Name:          acct.Name,
			PrincipalType: adm.PrincipalTypeStaff,
			IsActive:      true,
		}
		if err := db.Create(&user).Error; err != nil {
			return adm.User{}, fmt.Errorf("create dev user %s: %w", acct.Email, err)
		}
		return user, nil
	case err != nil:
		return adm.User{}, fmt.Errorf("lookup dev user %s: %w", acct.Email, err)
	default:
		user.PasswordHash = seedPassword
		user.Name = acct.Name
		user.UniCode = acct.UniCode
		user.PrincipalType = adm.PrincipalTypeStaff
		user.IsActive = true
		user.UpdatedAt = now
		if err := db.Save(&user).Error; err != nil {
			return adm.User{}, fmt.Errorf("update dev user %s: %w", acct.Email, err)
		}
		return user, nil
	}
}

func ensureUserRoles(db *gorm.DB, userID uuid.UUID, want []adm.RoleName, roleIDs map[adm.RoleName]uuid.UUID, now time.Time) error {
	for _, name := range want {
		roleID, ok := roleIDs[name]
		if !ok {
			return fmt.Errorf("role %q not in database", name)
		}
		var existing adm.UserRole
		err := db.Where("user_id = ? AND role_id = ? AND revoked_at IS NULL", userID, roleID).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		ur := adm.UserRole{
			BaseModel:  plat.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now},
			UserID:     userID,
			RoleID:     roleID,
			AssignedBy: &userID,
			AssignedAt: now,
		}
		if err := db.Create(&ur).Error; err != nil {
			return err
		}
	}
	return nil
}
