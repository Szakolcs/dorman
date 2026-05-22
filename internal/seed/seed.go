package seed

import (
	"fmt"
	"time"

	"dorm-man/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const markerNickname = "dev"

// Run migrates and inserts demo data for every application table.
// When reset is false and seed data already exists, the run is skipped.
func Run(db *gorm.DB, reset bool) error {
	if reset {
		if err := clearAll(db); err != nil {
			return fmt.Errorf("clear database: %w", err)
		}
		if err := db.AutoMigrate(models.All()...); err != nil {
			return fmt.Errorf("migrate after reset: %w", err)
		}
	} else {
		var count int64
		if err := db.Model(&models.User{}).Where("nickname = ?", markerNickname).Count(&count).Error; err != nil {
			return fmt.Errorf("check existing seed: %w", err)
		}
		if count > 0 {
			return nil
		}
	}

	now := time.Now().UTC().Truncate(time.Second)
	ago := func(d time.Duration) time.Time { return now.Add(-d) }
	later := func(d time.Duration) time.Time { return now.Add(d) }

	// --- RBAC chain (Operation -> Permission -> Role) ---
	ops := []models.Operation{
		{Name: "platform.full_access"},
		{Name: "administration.manage"},
		{Name: "maintenance.manage"},
		{Name: "doorman.manage"},
		{Name: "tenant.self_service"},
	}
	if err := db.Create(&ops).Error; err != nil {
		return fmt.Errorf("seed operations: %w", err)
	}

	perms := []models.Permission{
		{Name: "platform.all", Read: true, Write: true, OperationID: ops[0].ID},
		{Name: "administration.all", Read: true, Write: true, OperationID: ops[1].ID},
		{Name: "maintenance.all", Read: true, Write: true, OperationID: ops[2].ID},
		{Name: "doorman.all", Read: true, Write: true, OperationID: ops[3].ID},
		{Name: "tenant.basic", Read: true, Write: false, OperationID: ops[4].ID},
	}
	if err := db.Create(&perms).Error; err != nil {
		return fmt.Errorf("seed permissions: %w", err)
	}

	roles := []models.Role{
		{Name: "dev", PermissionID: perms[0].ID},
		{Name: "admin", PermissionID: perms[1].ID},
		{Name: "maintainer", PermissionID: perms[2].ID},
		{Name: "doorman", PermissionID: perms[3].ID},
		{Name: "tenant", PermissionID: perms[4].ID},
	}
	if err := db.Create(&roles).Error; err != nil {
		return fmt.Errorf("seed roles: %w", err)
	}

	roleByName := map[string]models.Role{}
	for _, r := range roles {
		roleByName[r.Name] = r
	}

	type account struct {
		nickname string
		password string
		name     string
		email    string
		role     string
	}

	accounts := []account{
		{
			nickname: "dev",
			password: "dev",
			name:     "Development Superuser Account",
			email:    "dev@dorm.local",
			role:     "dev",
		},
		{
			nickname: "admin",
			password: "admin",
			name:     "Campus Housing Administrator",
			email:    "admin@dorm.local",
			role:     "admin",
		},
		{
			nickname: "maintainer",
			password: "maintainer",
			name:     "Facilities and Maintenance Coordinator",
			email:    "maintainer@dorm.local",
			role:     "maintainer",
		},
		{
			nickname: "doorman",
			password: "doorman",
			name:     "Front Desk and Security Officer",
			email:    "doorman@dorm.local",
			role:     "doorman",
		},
		{
			nickname: "tenant1",
			password: "tenant1",
			name:     "Anna Kovacs",
			email:    "tenant1@student.uni.local",
			role:     "tenant",
		},
		{
			nickname: "tenant2",
			password: "tenant2",
			name:     "Mate Nagy",
			email:    "tenant2@student.uni.local",
			role:     "tenant",
		},
	}

	users := make(map[string]models.User, len(accounts))
	for _, acct := range accounts {
		hash, err := bcrypt.GenerateFromPassword([]byte(acct.password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password for %s: %w", acct.nickname, err)
		}
		user := models.User{
			Name:         acct.name,
			Email:        acct.email,
			Nickname:     acct.nickname,
			PasswordHash: string(hash),
			AvatarURL:    "https://example.local/avatars/" + acct.nickname + ".png",
			PhotoUrl:     "https://example.local/photos/" + acct.nickname + "-profile.jpg",
			RoleID:       roleByName[acct.role].ID,
		}
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("seed user %s: %w", acct.nickname, err)
		}
		users[acct.nickname] = user
	}

	// --- Housing ---
	building := models.Building{
		Name: "Central Dormitory Complex — North Wing",
		Code: "CDCNW",
	}
	if err := db.Create(&building).Error; err != nil {
		return fmt.Errorf("seed building: %w", err)
	}

	sharedAreas := []models.SharedArea{
		{
			BuildingID: &building.ID,
			Name:       "Ground Floor Common Lounge and Study Area",
			Code:       "LOUNGE-A",
		},
		{
			BuildingID: &building.ID,
			Name:       "Basement Fitness and Recreation Gymnasium",
			Code:       "GYM-01",
		},
	}
	if err := db.Create(&sharedAreas).Error; err != nil {
		return fmt.Errorf("seed shared areas: %w", err)
	}

	flats := []models.Flat{
		{BuildingID: &building.ID, Name: "Flat 101 — East Corridor", Floor: 1},
		{BuildingID: &building.ID, Name: "Flat 102 — West Corridor", Floor: 1},
		{BuildingID: &building.ID, Name: "Flat 201 — Upper Level Suite", Floor: 2},
	}
	if err := db.Create(&flats).Error; err != nil {
		return fmt.Errorf("seed flats: %w", err)
	}

	rooms := []models.Room{
		{FlatID: flats[0].ID, Number: "101A", Capacity: 2},
		{FlatID: flats[0].ID, Number: "101B", Capacity: 2},
		{FlatID: flats[1].ID, Number: "102A", Capacity: 1},
		{FlatID: flats[1].ID, Number: "102B", Capacity: 2},
		{FlatID: flats[2].ID, Number: "201A", Capacity: 3},
	}
	if err := db.Create(&rooms).Error; err != nil {
		return fmt.Errorf("seed rooms: %w", err)
	}

	degreeBSc := models.DegreeBSc
	facultyEng := models.FacultyEngineering
	age21 := 21
	age22 := 22
	sexFemale := models.SexFemale
	sexMale := models.SexMale
	natHU := models.NationalityHungarian

	tenant1UserID := users["tenant1"].ID
	tenant2UserID := users["tenant2"].ID

	tenants := []models.Tenant{
		{
			UserID:       &tenant1UserID,
			StudentCode:  "STU-2024-00142",
			Degree:       &degreeBSc,
			Faculty:      &facultyEng,
			Age:          &age21,
			Sex:          &sexFemale,
			Nationality:  &natHU,
			IsActive:     true,
			RegisteredAt: ago(180 * 24 * time.Hour),
		},
		{
			UserID:       &tenant2UserID,
			StudentCode:  "STU-2024-00287",
			Degree:       &degreeBSc,
			Faculty:      &facultyEng,
			Age:          &age22,
			Sex:          &sexMale,
			Nationality:  &natHU,
			IsActive:     true,
			RegisteredAt: ago(175 * 24 * time.Hour),
		},
		{
			StudentCode:  "STU-2023-00991",
			Degree:       &degreeBSc,
			Faculty:      &facultyEng,
			Age:          &age22,
			Sex:          &sexFemale,
			Nationality:  &natHU,
			IsActive:     true,
			RegisteredAt: ago(400 * 24 * time.Hour),
		},
	}
	if err := db.Create(&tenants).Error; err != nil {
		return fmt.Errorf("seed tenants: %w", err)
	}

	assignments := []models.RoomAssignment{
		{
			TenantID:    tenants[0].ID,
			RoomID:      rooms[0].ID,
			EffectiveAt: ago(170 * 24 * time.Hour),
		},
		{
			TenantID:    tenants[1].ID,
			RoomID:      rooms[2].ID,
			EffectiveAt: ago(165 * 24 * time.Hour),
		},
		{
			TenantID:    tenants[2].ID,
			RoomID:      rooms[4].ID,
			EffectiveAt: ago(360 * 24 * time.Hour),
		},
	}
	if err := db.Create(&assignments).Error; err != nil {
		return fmt.Errorf("seed room assignments: %w", err)
	}

	inUse := ago(200 * 24 * time.Hour)
	purchase := ago(730 * 24 * time.Hour)
	inventory := []models.InventoryItem{
		{
			Name:         "Ergonomic Desk Chair with Adjustable Lumbar Support",
			Description:  "Fabric-upholstered swivel chair purchased for the 101A bedroom. The gas lift was serviced last semester and the casters were replaced to reduce noise on the laminate flooring.",
			RoomID:       &rooms[0].ID,
			FlatID:       &flats[0].ID,
			BuildingID:   &building.ID,
			Condition:    models.InventoryConditionGood,
			Status:       models.InventoryStatusInUse,
			PurchaseDate: purchase,
			InUseDate:    &inUse,
		},
		{
			Name:         "Compact Refrigerator — Energy Class A",
			Description:  "Shared kitchen appliance assigned to Flat 101. Tenants are reminded to defrost monthly and report unusual compressor noise to the maintenance desk.",
			FlatID:       &flats[0].ID,
			BuildingID:   &building.ID,
			Condition:    models.InventoryConditionUsed,
			Status:       models.InventoryStatusInUse,
			PurchaseDate: purchase,
			InUseDate:    &inUse,
		},
		{
			Name:         "Wall-Mounted Bulletin Board with Cork Surface",
			Description:  "Installed in the ground floor lounge for official notices, event flyers, and community announcements. Surface was resurfaced during the summer renovation project.",
			SharedAreaID: &sharedAreas[0].ID,
			BuildingID:   &building.ID,
			Condition:    models.InventoryConditionGood,
			Status:       models.InventoryStatusInUse,
			PurchaseDate: purchase,
			InUseDate:    &inUse,
		},
		{
			Name:         "Adjustable Dumbbell Set (5–25 kg)",
			Description:  "Stored in the gymnasium equipment rack. Users must wipe down handles after use and return weights to the marked positions to keep the area safe for everyone.",
			SharedAreaID: &sharedAreas[1].ID,
			BuildingID:   &building.ID,
			Condition:    models.InventoryConditionGood,
			Status:       models.InventoryStatusInUse,
			PurchaseDate: purchase,
			InUseDate:    &inUse,
		},
	}
	if err := db.Create(&inventory).Error; err != nil {
		return fmt.Errorf("seed inventory: %w", err)
	}

	jobs := []models.OperationalJob{
		{
			Title:       "Annual Fire Safety Inspection — All Residential Floors",
			Description: "Coordinate with the municipal fire department to verify extinguishers, emergency lighting, and evacuation signage throughout the north wing corridors and stairwells.",
			StartsAt:    later(7 * 24 * time.Hour),
			EndsAt:      later(8 * 24 * time.Hour),
			Priority:    models.JobPriorityHigh,
			Status:      models.JobStatusScheduled,
		},
		{
			Title:       "Replace Worn Carpeting in Flat 102 Common Hallway",
			Description: "Remove existing carpet tiles, inspect subfloor moisture levels, and install new commercial-grade carpet suitable for high foot traffic near the shared bathroom entrance.",
			StartsAt:    ago(14 * 24 * time.Hour),
			EndsAt:      ago(10 * 24 * time.Hour),
			Priority:    models.JobPriorityMedium,
			Status:      models.JobStatusFinished,
		},
		{
			Title:       "HVAC Filter Replacement — Basement Gym Ventilation",
			Description: "Planned maintenance on the gym air handling unit. Access to the fitness area will be restricted for approximately four hours while filters and belt tension are checked.",
			StartsAt:    later(3 * 24 * time.Hour),
			EndsAt:      later(3*24*time.Hour + 6*time.Hour),
			Priority:    models.JobPriorityLow,
			Status:      models.JobStatusPlanned,
		},
	}
	if err := db.Create(&jobs).Error; err != nil {
		return fmt.Errorf("seed operational jobs: %w", err)
	}

	auditOld := fmt.Sprintf(`{"status":"%s"}`, models.InventoryStatusInStock)
	auditNew := fmt.Sprintf(`{"status":"%s","room_id":"%s"}`, models.InventoryStatusInUse, rooms[0].ID)
	adminUserID := users["admin"].ID
	audits := []models.Audit{
		{
			TableName: "inventory_items",
			Operation: "UPDATE",
			OldData:   auditOld,
			NewData:   auditNew,
			ChangedAt: ago(200 * 24 * time.Hour).Format(time.RFC3339),
			ChangedBy: &adminUserID,
		},
		{
			TableName: "room_assignments",
			Operation: "INSERT",
			OldData:   `{}`,
			NewData:   fmt.Sprintf(`{"tenant_id":"%s","room_id":"%s"}`, tenants[0].ID, rooms[0].ID),
			ChangedAt: assignments[0].EffectiveAt.Format(time.RFC3339),
			ChangedBy: &adminUserID,
		},
	}
	if err := db.Create(&audits).Error; err != nil {
		return fmt.Errorf("seed audits: %w", err)
	}

	// --- Forum content ---
	pub := models.Publication{
		Post: models.Post{
			Title: "Welcome to the New Academic Year — Important Housing Guidelines",
			Description: "The housing office published an extended guide covering quiet hours, guest registration procedures, " +
				"recycling schedules, and emergency contact numbers. Please read the full document before inviting visitors " +
				"or requesting maintenance assistance through the tenant portal.",
			State:    models.PublicationStatePublished,
			AuthorID: users["admin"].ID,
		},
	}
	if err := db.Create(&pub).Error; err != nil {
		return fmt.Errorf("seed publication: %w", err)
	}

	activity := models.Activity{
		Post: models.Post{
			Title: "Morning Yoga and Stretching Session in the Gymnasium",
			Description: "Join the weekly guided session led by a certified instructor. Mats are available on a first-come basis; " +
				"please arrive ten minutes early to sign the safety waiver at the front desk.",
			State:    models.PublicationStatePublished,
			AuthorID: users["admin"].ID,
		},
		SharedAreaID: &sharedAreas[1].ID,
		Capacity:     20,
		BookedCount:  2,
	}
	if err := db.Create(&activity).Error; err != nil {
		return fmt.Errorf("seed activity: %w", err)
	}

	eventChat := models.ChatRoom{
		Kind:  models.RoomKindEvent,
		Title: "Movie Night Discussion Room",
		Topic: "Coordination thread for snacks, seating, and accessibility requests for the Friday screening.",
	}
	if err := db.Create(&eventChat).Error; err != nil {
		return fmt.Errorf("seed event chat room: %w", err)
	}

	event := models.Event{
		Post: models.Post{
			Title: "Community Movie Night — Classic Cinema in the Lounge",
			Description: "We will project a subtitled classic film in the ground floor lounge. Popcorn and tea will be provided; " +
				"bring a blanket if you prefer the floor seating area near the windows.",
			State:    models.PublicationStatePublished,
			AuthorID: users["admin"].ID,
		},
		SharedAreaID: &sharedAreas[0].ID,
		StartsAt:     later(5 * 24 * time.Hour),
		EndsAt:       later(5*24*time.Hour + 3*time.Hour),
		ChatRoomID:   &eventChat.ID,
	}
	if err := db.Create(&event).Error; err != nil {
		return fmt.Errorf("seed event: %w", err)
	}

	bookings := []models.ActivityBooking{
		{ActivityID: activity.ID, UserID: users["tenant1"].ID},
		{ActivityID: activity.ID, UserID: users["tenant2"].ID},
	}
	if err := db.Create(&bookings).Error; err != nil {
		return fmt.Errorf("seed activity bookings: %w", err)
	}

	attendances := []models.EventAttendance{
		{EventID: event.ID, UserID: users["tenant1"].ID, Intent: models.EventAttendanceInterested},
		{EventID: event.ID, UserID: users["tenant2"].ID, Intent: models.EventAttendanceBusy},
		{EventID: event.ID, UserID: users["admin"].ID, Intent: models.EventAttendanceInterested},
	}
	if err := db.Create(&attendances).Error; err != nil {
		return fmt.Errorf("seed event attendances: %w", err)
	}

	comments := []models.EventComment{
		{
			EventID:  event.ID,
			AuthorID: users["tenant1"].ID,
			Body:     "Could we enable subtitles in both Hungarian and English? Several flatmates asked after last semester's screening.",
		},
		{
			EventID:  event.ID,
			AuthorID: users["admin"].ID,
			Body:     "Absolutely — the AV team confirmed dual subtitles and an induction loop for hearing aid users near the front row.",
		},
	}
	if err := db.Create(&comments).Error; err != nil {
		return fmt.Errorf("seed event comments: %w", err)
	}

	// --- Chat ---
	pairKey := directPairKey(tenants[0].ID, tenants[1].ID)
	directChat := models.ChatRoom{
		Kind:          models.RoomKindDirect,
		Title:         "Direct conversation",
		Topic:         "Private messages between assigned tenants coordinating shared kitchen cleaning and laundry schedules.",
		DirectPairKey: &pairKey,
	}
	groupChat := models.ChatRoom{
		Kind:  models.RoomKindGroup,
		Title: "Flat 101 — Kitchen Duty Reminders",
		Topic: "Informal group for flatmates on the east corridor to agree on weekly chores and grocery runs.",
	}
	if err := db.Create(&directChat).Error; err != nil {
		return fmt.Errorf("seed direct chat room: %w", err)
	}
	if err := db.Create(&groupChat).Error; err != nil {
		return fmt.Errorf("seed group chat room: %w", err)
	}

	lastRead := ago(2 * time.Hour)
	memberships := []models.Membership{
		{RoomID: directChat.ID, TenantID: tenants[0].ID, Role: models.MembershipRoleMember, JoinedAt: ago(30 * 24 * time.Hour), LastReadAt: &lastRead},
		{RoomID: directChat.ID, TenantID: tenants[1].ID, Role: models.MembershipRoleMember, JoinedAt: ago(30 * 24 * time.Hour)},
		{RoomID: groupChat.ID, TenantID: tenants[0].ID, Role: models.MembershipRoleAdmin, JoinedAt: ago(60 * 24 * time.Hour)},
		{RoomID: eventChat.ID, TenantID: tenants[0].ID, Role: models.MembershipRoleMember, JoinedAt: ago(3 * 24 * time.Hour)},
		{RoomID: eventChat.ID, TenantID: tenants[1].ID, Role: models.MembershipRoleMember, JoinedAt: ago(3 * 24 * time.Hour)},
	}
	if err := db.Create(&memberships).Error; err != nil {
		return fmt.Errorf("seed memberships: %w", err)
	}

	messages := []models.Message{
		{
			RoomID:         directChat.ID,
			SenderTenantID: tenants[0].ID,
			Body:           "Hi Mate — are you free Sunday afternoon to deep-clean the kitchen? The inspection is scheduled for Monday morning.",
		},
		{
			RoomID:         directChat.ID,
			SenderTenantID: tenants[1].ID,
			Body:           "Sunday works for me after 15:00. I can take the oven and you handle the fridge if that split sounds fair.",
		},
		{
			RoomID:         groupChat.ID,
			SenderTenantID: tenants[0].ID,
			Body:           "Reminder: recycling pickup is Tuesday. Please flatten cardboard boxes and rinse plastic containers before placing them in the bin room.",
		},
		{
			RoomID:         eventChat.ID,
			SenderTenantID: tenants[0].ID,
			Body:           "I can bring a large thermos of tea — let me know if anyone has allergies we should avoid.",
		},
	}
	if err := db.Create(&messages).Error; err != nil {
		return fmt.Errorf("seed messages: %w", err)
	}

	ignore := models.Ignore{
		TenantID:        tenants[1].ID,
		IgnoredTenantID: tenants[2].ID,
	}
	if err := db.Create(&ignore).Error; err != nil {
		return fmt.Errorf("seed ignore: %w", err)
	}

	// --- Maintenance ---
	reported := models.StatusReported
	inProgress := models.StatusInProgress
	resolved := models.StatusResolved
	closed := models.StatusClosed

	ticket := models.Ticket{
		FlatID:          &flats[0].ID,
		Category:        models.CategoryIncident,
		Severity:        models.SeverityMedium,
		Impact:          models.ImpactAffectsDaily,
		Status:          closed,
		Description:     "Intermittent loss of hot water in Flat 101 shared bathroom between 07:00 and 09:00 on weekdays. Tenants report the issue clears after thirty minutes but returns the next morning.",
		CreatedByUserID: users["tenant1"].ID,
	}
	if err := db.Create(&ticket).Error; err != nil {
		return fmt.Errorf("seed ticket: %w", err)
	}

	statusChanges := []models.StatusChange{
		{TicketID: ticket.ID, FromStatus: nil, ToStatus: reported, Note: "Ticket opened automatically from tenant maintenance portal submission."},
		{TicketID: ticket.ID, FromStatus: &reported, ToStatus: inProgress, Note: "Maintainer assigned and initial thermostatic valve inspection scheduled."},
		{TicketID: ticket.ID, FromStatus: &inProgress, ToStatus: resolved, Note: "Faulty circulation pump replaced; hot water stable during morning peak."},
		{TicketID: ticket.ID, FromStatus: &resolved, ToStatus: closed, Note: "Tenant confirmed satisfactory operation for five consecutive days."},
	}
	if err := db.Create(&statusChanges).Error; err != nil {
		return fmt.Errorf("seed status changes: %w", err)
	}

	// --- Doorman ---
	tenantEntries := []models.TenantEntry{
		{
			UserID:      users["tenant1"].ID,
			Status:      models.CheckedIn,
			TimeOfEntry: ago(6 * time.Hour),
		},
		{
			UserID:      users["tenant1"].ID,
			Status:      models.CheckedOut,
			TimeOfEntry: ago(18 * time.Hour),
		},
		{
			UserID:      users["tenant2"].ID,
			Status:      models.CheckedIn,
			TimeOfEntry: ago(1 * time.Hour),
		},
	}
	if err := db.Create(&tenantEntries).Error; err != nil {
		return fmt.Errorf("seed tenant entries: %w", err)
	}

	guests := []models.GuestEntry{
		{
			HostTenantID: tenants[0].ID,
			GuestName:    "Eva Szabo — visiting family member",
			IDNotes:      "National ID card presented at reception; host confirmed overnight stay within guest policy limits for exam week.",
		},
		{
			HostTenantID: tenants[1].ID,
			GuestName:    "Peter Horvath — university project collaborator",
			IDNotes:      "Passport and student ID verified; expected departure same day before 22:00 quiet hours.",
		},
	}
	if err := db.Create(&guests).Error; err != nil {
		return fmt.Errorf("seed guest entries: %w", err)
	}

	return nil
}

func directPairKey(a, b uuid.UUID) string {
	lo, hi := a.String(), b.String()
	if lo > hi {
		lo, hi = hi, lo
	}
	return lo + ":" + hi
}

func clearAll(db *gorm.DB) error {
	if err := db.Exec("DROP SCHEMA public CASCADE").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE SCHEMA public").Error; err != nil {
		return err
	}
	return db.Exec("GRANT ALL ON SCHEMA public TO public").Error
}
