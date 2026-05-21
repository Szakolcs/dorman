package seed

import (
	plat "dorm-man/internal/models/cross-cutting"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	adm "dorm-man/internal/models/administration"
	chatm "dorm-man/internal/models/chat"
	fm "dorm-man/internal/models/forum"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Seeder populates the database per docs/database/seed.md.
type Seeder struct {
	db    *gorm.DB
	rng   *rand.Rand
	now   time.Time
	start time.Time

	buildings  []adm.Building
	flats      []adm.Flat
	rooms      []adm.Room
	roles      []adm.Role
	users      []adm.User
	staff      []adm.User
	tenants    []adm.Tenant
	activities []adm.Activity
	events     []adm.Event
	inventory  []adm.InventoryItem

	flatRooms   []chatm.ChatRoom
	directRooms []chatm.ChatRoom
	forumPosts  []fm.ForumPost
	pollPosts   []fm.ForumPost
}

func New(db *gorm.DB, seed uint64) *Seeder {
	now := time.Now().UTC()
	return &Seeder{
		db:    db,
		rng:   rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)),
		now:   now,
		start: now.AddDate(0, -18, 0),
	}
}

func (s *Seeder) Run() error {
	steps := []struct {
		name string
		fn   func() error
	}{
		{"foundation", s.seedFoundation},
		{"assignments", s.seedAssignments},
		{"publications", s.seedPublications},
		{"inventory & maintenance", s.seedInventoryMaintenance},
		{"doorman", s.seedDoorman},
		{"chat", s.seedChat},
		{"forum", s.seedForum},
		{"audit", s.seedAudit},
	}
	for _, step := range steps {
		log.Printf("seed: %s", step.name)
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}
	}
	return s.printCounts()
}

func (s *Seeder) printCounts() error {
	tables := []string{
		"buildings", "roles", "users", "flats", "rooms", "tenants", "user_roles",
		"room_assignments", "activities", "events", "inventory_items",
		"maintenance_tickets", "ticket_status_changes", "operational_jobs",
		"packages", "package_notifications", "tenant_entry_tokens",
		"access_events", "guest_visits", "guest_access_events", "item_loans",
		"chat_rooms", "chat_tenant_profiles", "chat_messages", "chat_room_members",
		"chat_membership_sync_logs", "forum_posts", "forum_post_schedules",
		"forum_post_updates", "forum_polls", "forum_poll_options", "forum_poll_votes",
		"forum_attendance_intents", "forum_comments", "forum_reactions",
		"forum_post_views", "forum_moderation_actions", "audit_events",
	}
	for _, t := range tables {
		var n int64
		if err := s.db.Table(t).Count(&n).Error; err != nil {
			return err
		}
		log.Printf("  %s: %d", t, n)
	}
	return nil
}

func (s *Seeder) seedFoundation() error {
	names := []struct{ name, code string }{
		{"North Hall", "A"}, {"Central House", "B"}, {"South Wing", "C"},
	}
	for i, b := range names {
		s.buildings = append(s.buildings, adm.Building{
			BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: s.start, UpdatedAt: s.now},
			Name:      b.name,
			Code:      b.code,
		})
		_ = i
	}
	if err := batchCreate(s.db, s.buildings); err != nil {
		return err
	}

	roleDefs := []struct {
		name adm.RoleName
		desc string
	}{
		{adm.RoleAdministrator, "System administrator"},
		{adm.RoleDirector, "Building / dorm manager"},
		{adm.RoleOfficeWorker, "Reception and office staff"},
		{adm.RoleDoorman, "Security and entry desk"},
		{adm.RoleTenant, "Resident tenant"},
	}
	for _, r := range roleDefs {
		s.roles = append(s.roles, adm.Role{
			BaseModel:   plat.BaseModel{ID: uuid.New(), CreatedAt: s.start, UpdatedAt: s.now},
			Name:        r.name,
			Description: r.desc,
		})
	}
	if err := batchCreate(s.db, s.roles); err != nil {
		return err
	}
	roleByName := map[adm.RoleName]uuid.UUID{}
	for _, r := range s.roles {
		roleByName[r.Name] = r.ID
	}

	// Fixed dev admin (administration + doorman); password matches seedPassword.
	adminUser := adm.User{
		BaseModel:     plat.BaseModel{ID: uuid.New(), CreatedAt: s.start, UpdatedAt: s.now},
		UniCode:       "ADMIN001",
		Email:         DevAdminEmail,
		PasswordHash:  seedPassword,
		Name:          "Admin User",
		PrincipalType: adm.PrincipalTypeStaff,
		IsActive:      true,
	}
	s.staff = append(s.staff, adminUser)
	s.users = append(s.users, adminUser)

	// 25 staff
	staffRoles := []adm.RoleName{
		adm.RoleAdministrator, adm.RoleDirector, adm.RoleOfficeWorker, adm.RoleDoorman,
	}
	for i := 0; i < 25; i++ {
		pt := adm.PrincipalTypeStaff
		u := adm.User{
			BaseModel:     plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			UniCode:       uniCode("STF", i),
			Email:         email("staff", i),
			PasswordHash:  seedPassword,
			Name:          fmt.Sprintf("Staff Member %d", i+1),
			PrincipalType: pt,
			IsActive:      true,
		}
		s.staff = append(s.staff, u)
		s.users = append(s.users, u)
	}
	// 400 tenant users (created before tenants; linked below)
	tenantUsers := make([]adm.User, 0, 400)
	for i := 0; i < 400; i++ {
		u := adm.User{
			BaseModel:     plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			UniCode:       uniCode("TNT", i),
			Email:         email("tenant", i),
			PasswordHash:  seedPassword,
			Name:          fmt.Sprintf("Tenant User %d", i+1),
			PrincipalType: adm.PrincipalTypeTenant,
			IsActive:      true,
		}
		tenantUsers = append(tenantUsers, u)
		s.users = append(s.users, u)
	}
	if err := batchCreate(s.db, s.users); err != nil {
		return err
	}

	var userRoles []adm.UserRole
	assigner := s.staff[0].ID
	type roleKey struct{ userID, roleID uuid.UUID }
	seenRoles := map[roleKey]struct{}{}
	addRole := func(userID, roleID uuid.UUID) {
		k := roleKey{userID, roleID}
		if _, ok := seenRoles[k]; ok {
			return
		}
		seenRoles[k] = struct{}{}
		userRoles = append(userRoles, adm.UserRole{
			BaseModel:  plat.BaseModel{ID: uuid.New(), CreatedAt: s.now, UpdatedAt: s.now},
			UserID:     userID,
			RoleID:     roleID,
			AssignedBy: &assigner,
			AssignedAt: randTime(s.rng, s.start, s.now),
		})
	}
	for i, u := range s.staff {
		if u.Email == DevAdminEmail {
			addRole(u.ID, roleByName[adm.RoleAdministrator])
			addRole(u.ID, roleByName[adm.RoleDoorman])
			addRole(u.ID, roleByName[adm.RoleOfficeWorker])
			addRole(u.ID, roleByName[adm.RoleDirector])
			continue
		}
		addRole(u.ID, roleByName[staffRoles[i%len(staffRoles)]])
		if i < 5 {
			addRole(u.ID, roleByName[adm.RoleOfficeWorker])
		}
	}
	tenantRoleID := roleByName[adm.RoleTenant]
	for _, u := range tenantUsers {
		addRole(u.ID, tenantRoleID)
	}
	for i := 0; len(userRoles) < 450 && i < len(s.staff); i++ {
		addRole(s.staff[i].ID, roleByName[adm.RoleDirector])
	}
	if err := batchCreate(s.db, userRoles); err != nil {
		return err
	}

	// 90 flats (~30 per building)
	flatIdx := 0
	for _, b := range s.buildings {
		for f := 0; f < 30; f++ {
			floor := f % 6
			s.flats = append(s.flats, adm.Flat{
				BaseModel:  plat.BaseModel{ID: uuid.New(), CreatedAt: s.start, UpdatedAt: s.now},
				BuildingID: &b.ID,
				Name:       fmt.Sprintf("%s-%d%02d", b.Code, floor, f%10+1),
				Floor:      floor,
			})
			flatIdx++
		}
	}
	if err := batchCreate(s.db, s.flats); err != nil {
		return err
	}

	// ~300 rooms
	roomNo := 0
	for _, fl := range s.flats {
		nRooms := 2 + s.rng.IntN(3) // 2-4
		for r := 0; r < nRooms; r++ {
			cap := 1 + s.rng.IntN(3)
			archived := s.rng.Float64() < 0.05
			s.rooms = append(s.rooms, adm.Room{
				BaseModel:  plat.BaseModel{ID: uuid.New(), CreatedAt: s.start, UpdatedAt: s.now},
				FlatID:     fl.ID,
				Number:     fmt.Sprintf("%d", r+1),
				Capacity:   cap,
				IsArchived: archived,
			})
			roomNo++
			if len(s.rooms) >= 300 {
				break
			}
		}
		if len(s.rooms) >= 300 {
			break
		}
	}
	if err := batchCreate(s.db, s.rooms); err != nil {
		return err
	}

	degrees := []adm.DegreeType{adm.DegreeBSc, adm.DegreeBA, adm.DegreeMSc, adm.DegreeMA, adm.DegreePhD}
	faculties := []adm.FacultyType{adm.FacultyScience, adm.FacultyHumanities, adm.FacultyEngineering, adm.FacultyMedicine}
	sexes := []adm.SexType{adm.SexFemale, adm.SexMale, adm.SexOther}
	nationalities := []adm.NationalityType{adm.NationalityHungarian, adm.NationalityInternational}

	for i := 0; i < 400; i++ {
		uid := tenantUsers[i].ID
		age := 18 + s.rng.IntN(13)
		deg := pick(s.rng, degrees)
		fac := pick(s.rng, faculties)
		sex := pick(s.rng, sexes)
		nat := pick(s.rng, nationalities)
		s.tenants = append(s.tenants, adm.Tenant{
			BaseModel:    plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			UserID:       &uid,
			StudentCode:  studentCode(i),
			Name:         fmt.Sprintf("Resident %d", i+1),
			Email:        email("resident", i),
			Degree:       &deg,
			Faculty:      &fac,
			Age:          &age,
			Sex:          &sex,
			Nationality:  &nat,
			IsActive:     true,
			RegisteredAt: randTime(s.rng, s.start, s.now),
		})
	}
	if err := batchCreate(s.db, s.tenants); err != nil {
		return err
	}

	var profiles []chatm.ChatTenantProfile
	for _, t := range s.tenants {
		nick := fmt.Sprintf("nick_%s", t.StudentCode)
		profiles = append(profiles, chatm.ChatTenantProfile{
			BaseModel:        plat.BaseModel{ID: uuid.New(), CreatedAt: s.now, UpdatedAt: s.now},
			TenantID:         t.ID,
			Nickname:         &nick,
			Bio:              "Seed profile",
			ProfileUpdatedAt: s.now,
		})
	}
	return batchCreate(s.db, profiles)
}
