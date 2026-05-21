package seed

import (
	plat "dorm-man/internal/models/cross-cutting"
	"fmt"
	"time"

	adm "dorm-man/internal/models/administration"
	dm "dorm-man/internal/models/doorman"

	"github.com/google/uuid"
)

func (s *Seeder) seedAssignments() error {
	activeRooms := make([]adm.Room, 0, len(s.rooms))
	for _, r := range s.rooms {
		if !r.IsArchived {
			activeRooms = append(activeRooms, r)
		}
	}
	creator := s.staff[0].ID
	var assignments []adm.RoomAssignment
	tenantIdx := 0

	// ~400 active assignments (one active per tenant max)
	for tenantIdx < len(s.tenants) && tenantIdx < 400 {
		room := activeRooms[tenantIdx%len(activeRooms)]
		at := randTime(s.rng, s.start, s.now.AddDate(0, -1, 0))
		assignments = append(assignments, adm.RoomAssignment{
			BaseModel:       plat.BaseModel{ID: uuid.New(), CreatedAt: at, UpdatedAt: s.now},
			TenantID:        s.tenants[tenantIdx].ID,
			RoomID:          room.ID,
			EffectiveAt:     at,
			CreatedByUserID: creator,
		})
		tenantIdx++
	}

	// ~120 historical
	for i := 0; i < 120 && i < len(s.tenants); i++ {
		t := s.tenants[i]
		room := pick(s.rng, activeRooms)
		start := randTime(s.rng, s.start, s.now.AddDate(0, -6, 0))
		end := start.AddDate(0, 3, 0)
		assignments = append(assignments, adm.RoomAssignment{
			BaseModel:       plat.BaseModel{ID: uuid.New(), CreatedAt: start, UpdatedAt: s.now},
			TenantID:        t.ID,
			RoomID:          room.ID,
			EffectiveAt:     start,
			EndedAt:         &end,
			CreatedByUserID: creator,
		})
	}
	return batchCreate(s.db, assignments)
}

func (s *Seeder) seedPublications() error {
	states := []adm.PublicationState{
		adm.PublicationStatePublished,
		adm.PublicationStateDraft,
		adm.PublicationStateArchived,
	}
	for i := 0; i < 50; i++ {
		var bid *uuid.UUID
		if s.rng.Float64() < 0.7 {
			b := s.buildings[s.rng.IntN(len(s.buildings))].ID
			bid = &b
		}
		s.activities = append(s.activities, adm.Activity{
			BaseModel:   plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			Title:       fmt.Sprintf("Activity %d", i+1),
			Description: "Seed activity program",
			BuildingID:  bid,
			Location:    "Common area",
			Capacity:    20 + s.rng.IntN(30),
			State:       pick(s.rng, states),
		})
	}
	if err := batchCreate(s.db, s.activities); err != nil {
		return err
	}

	for i := 0; i < 200; i++ {
		var bid *uuid.UUID
		if s.rng.Float64() < 0.8 {
			b := s.buildings[s.rng.IntN(len(s.buildings))].ID
			bid = &b
		}
		start := randTime(s.rng, s.start, s.now.AddDate(0, 1, 0))
		end := start.Add(2 * time.Hour)
		s.events = append(s.events, adm.Event{
			BaseModel:       plat.BaseModel{ID: uuid.New(), CreatedAt: start, UpdatedAt: s.now},
			Title:           fmt.Sprintf("Event %d", i+1),
			Description:     "Seed dorm event",
			BuildingID:      bid,
			OrganizerUserID: s.staff[s.rng.IntN(len(s.staff))].ID,
			Location:        "Main hall",
			StartsAt:        start,
			EndsAt:          end,
			Capacity:        30 + s.rng.IntN(70),
			State:           pick(s.rng, states),
		})
	}
	return batchCreate(s.db, s.events)
}

func (s *Seeder) seedInventoryMaintenance() error {
	locTypes := []adm.InventoryLocationType{
		adm.InventoryLocationRoom, adm.InventoryLocationSharedArea,
	}
	conditions := []adm.InventoryCondition{
		adm.InventoryConditionGood, adm.InventoryConditionUsed, adm.InventoryConditionNew,
	}
	statuses := []adm.InventoryStatus{
		adm.InventoryStatusInUse, adm.InventoryStatusInStock,
	}

	for i := 0; i < 800; i++ {
		item := adm.InventoryItem{
			BaseModel:    plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			Name:         fmt.Sprintf("Item %d", i+1),
			Description:  "Seed inventory",
			LocationType: pick(s.rng, locTypes),
			Condition:    pick(s.rng, conditions),
			Status:       pick(s.rng, statuses),
			PurchaseDate: randTime(s.rng, s.start, s.now),
		}
		if i < 600 {
			r := s.rooms[s.rng.IntN(len(s.rooms))]
			item.RoomID = &r.ID
			item.FlatID = &s.flats[s.rng.IntN(len(s.flats))].ID
		} else {
			b := s.buildings[s.rng.IntN(len(s.buildings))].ID
			item.BuildingID = &b
		}
		s.inventory = append(s.inventory, item)
	}
	if err := batchCreate(s.db, s.inventory); err != nil {
		return err
	}

	ticketStatuses := []adm.MaintenanceStatus{
		adm.MaintenanceStatusReported, adm.MaintenanceStatusInProgress,
		adm.MaintenanceStatusResolved, adm.MaintenanceStatusClosed,
	}
	var tickets []adm.MaintenanceTicket
	var transitions []adm.TicketStatusChange
	for i := 0; i < 600; i++ {
		rid := s.rooms[s.rng.IntN(len(s.rooms))].ID
		creator := s.staff[s.rng.IntN(len(s.staff))].ID
		st := pick(s.rng, ticketStatuses)
		if s.rng.Float64() < 0.7 {
			st = adm.MaintenanceStatusClosed
		}
		t := adm.MaintenanceTicket{
			BaseModel:       plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			RoomID:          &rid,
			Category:        adm.MaintenanceCategoryIncident,
			Severity:        adm.MaintenanceSeverityMedium,
			Impact:          adm.MaintenanceImpactInconvenience,
			Status:          st,
			Description:     "Seed maintenance issue",
			CreatedByUserID: creator,
		}
		if s.rng.Float64() < 0.6 {
			a := s.staff[s.rng.IntN(len(s.staff))].ID
			t.AssigneeUserID = &a
		}
		tickets = append(tickets, t)
		nTrans := 2 + s.rng.IntN(2)
		prev := adm.MaintenanceStatusReported
		for j := 0; j < nTrans && len(transitions) < 1500; j++ {
			transitions = append(transitions, adm.TicketStatusChange{
				BaseModel:   plat.BaseModel{ID: uuid.New(), CreatedAt: t.CreatedAt, UpdatedAt: s.now},
				TicketID:    t.ID,
				ActorUserID: creator,
				FromStatus:  &prev,
				ToStatus:    st,
			})
			prev = st
		}
	}
	if err := batchCreate(s.db, tickets); err != nil {
		return err
	}
	if err := batchCreate(s.db, transitions); err != nil {
		return err
	}

	var jobs []adm.OperationalJob
	for i := 0; i < 300; i++ {
		rid := s.rooms[s.rng.IntN(len(s.rooms))].ID
		start := randTime(s.rng, s.start, s.now)
		jobs = append(jobs, adm.OperationalJob{
			BaseModel:       plat.BaseModel{ID: uuid.New(), CreatedAt: start, UpdatedAt: s.now},
			Title:           fmt.Sprintf("Job %d", i+1),
			Description:     "Seed operational job",
			RoomID:          &rid,
			AssigneeUserID:  s.staff[s.rng.IntN(len(s.staff))].ID,
			CreatedByUserID: &s.staff[0].ID,
			StartsAt:        start,
			EndsAt:          start.Add(4 * time.Hour),
			Priority:        adm.JobPriorityMedium,
			Status:          adm.JobStatusFinished,
		})
	}
	return batchCreate(s.db, jobs)
}

func (s *Seeder) seedDoorman() error {
	var packages []dm.Package
	for i := 0; i < 1000; i++ {
		var tid *uuid.UUID
		if s.rng.Float64() < 0.85 {
			t := s.tenants[s.rng.IntN(len(s.tenants))].ID
			tid = &t
		}
		reg := s.staff[s.rng.IntN(len(s.staff))].ID
		st := dm.PackageStatusReceived
		if s.rng.Float64() < 0.85 {
			st = dm.PackageStatusPickedUp
		} else if s.rng.Float64() < 0.5 {
			st = dm.PackageStatusNotified
		}
		pkg := dm.Package{
			BaseModel:          plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			RecipientLabel:     fmt.Sprintf("Recipient %d", i+1),
			Description:        "Seed package",
			TenantID:           tid,
			Status:             st,
			ReceivedAt:         randTime(s.rng, s.start, s.now),
			RegisteredByUserID: &reg,
		}
		if st == dm.PackageStatusPickedUp {
			t := pkg.ReceivedAt.Add(24 * time.Hour)
			pkg.PickedUpAt = &t
			pkg.PickedUpByUserID = &reg
		}
		packages = append(packages, pkg)
	}
	if err := batchCreate(s.db, packages); err != nil {
		return err
	}

	var notifs []dm.PackageNotification
	for _, p := range packages {
		n := 1 + s.rng.IntN(2)
		if p.TenantID == nil {
			continue
		}
		for j := 0; j < n && len(notifs) < 1500; j++ {
			ch := dm.PackageNotificationChannelInApp
			if j%2 == 1 {
				ch = dm.PackageNotificationChannelEmail
			}
			st := dm.PackageNotificationStatusSent
			sent := p.ReceivedAt
			notifs = append(notifs, dm.PackageNotification{
				BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: p.CreatedAt, UpdatedAt: s.now},
				PackageID: p.ID,
				TenantID:  *p.TenantID,
				Channel:   ch,
				Status:    st,
				SentAt:    &sent,
			})
		}
	}
	if err := batchCreate(s.db, notifs); err != nil {
		return err
	}

	var tokens []dm.TenantEntryToken
	for i := 0; i < 500; i++ {
		t := s.tenants[s.rng.IntN(len(s.tenants))].ID
		tokens = append(tokens, dm.TenantEntryToken{
			BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			TenantID:  t,
			PublicRef: fmt.Sprintf("QR-%08d", i+1),
		})
	}
	if err := batchCreate(s.db, tokens); err != nil {
		return err
	}

	var accessEvents []dm.AccessEvent
	for i := 0; i < 8000; i++ {
		var tid, aid *uuid.UUID
		if s.rng.Float64() < 0.9 {
			t := s.tenants[s.rng.IntN(len(s.tenants))].ID
			tid = &t
		}
		if s.rng.Float64() < 0.3 {
			a := s.staff[s.rng.IntN(len(s.staff))].ID
			aid = &a
		}
		outcome := dm.AccessEventOutcomeGranted
		if s.rng.Float64() < 0.05 {
			outcome = dm.AccessEventOutcomeDenied
		}
		accessEvents = append(accessEvents, dm.AccessEvent{
			BaseModel:   plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			TenantID:    tid,
			ActorUserID: aid,
			Outcome:     outcome,
			Source:      dm.AccessEventSourceQRScan,
			OccurredAt:  randTime(s.rng, s.start, s.now),
		})
	}
	if err := batchCreate(s.db, accessEvents); err != nil {
		return err
	}

	var visits []dm.GuestEntry
	for i := 0; i < 800; i++ {
		host := s.tenants[s.rng.IntN(len(s.tenants))].ID
		from := randTime(s.rng, s.start, s.now)
		to := from.Add(4 * time.Hour)
		reg := s.staff[s.rng.IntN(len(s.staff))].ID
		visits = append(visits, dm.GuestEntry{
			BaseModel:          plat.BaseModel{ID: uuid.New(), CreatedAt: from, UpdatedAt: s.now},
			HostTenantID:       host,
			GuestName:          fmt.Sprintf("Guest %d", i+1),
			ValidFrom:          from,
			ValidTo:            to,
			Status:             dm.GuestCheckedOut,
			RegisteredByUserID: &reg,
		})
	}
	if err := batchCreate(s.db, visits); err != nil {
		return err
	}

	var guestEvents []dm.GuestAccessEvent
	for _, v := range visits {
		actor := s.staff[s.rng.IntN(len(s.staff))].ID
		guestEvents = append(guestEvents, dm.GuestAccessEvent{
			BaseModel:    plat.BaseModel{ID: uuid.New(), CreatedAt: v.ValidFrom, UpdatedAt: s.now},
			GuestVisitID: v.ID,
			ActorUserID:  actor,
			EventType:    dm.GuestAccessEventCheckIn,
			OccurredAt:   v.ValidFrom,
		})
		if len(guestEvents) < 1600 {
			guestEvents = append(guestEvents, dm.GuestAccessEvent{
				BaseModel:    plat.BaseModel{ID: uuid.New(), CreatedAt: v.ValidTo, UpdatedAt: s.now},
				GuestVisitID: v.ID,
				ActorUserID:  actor,
				EventType:    dm.GuestAccessEventCheckOut,
				OccurredAt:   v.ValidTo,
			})
		}
	}
	if err := batchCreate(s.db, guestEvents); err != nil {
		return err
	}

	var loans []dm.ItemLoan
	for i := 0; i < 500; i++ {
		tenant := s.tenants[s.rng.IntN(len(s.tenants))].ID
		item := s.inventory[s.rng.IntN(len(s.inventory))].ID
		by := s.staff[s.rng.IntN(len(s.staff))].ID
		out := randTime(s.rng, s.start, s.now)
		loan := dm.ItemLoan{
			BaseModel:          plat.BaseModel{ID: uuid.New(), CreatedAt: out, UpdatedAt: s.now},
			TenantID:           tenant,
			InventoryItemID:    item,
			CheckedOutByUserID: by,
			CheckedOutAt:       out,
		}
		if i >= 150 {
			ret := out.Add(7 * 24 * time.Hour)
			loan.ReturnedAt = &ret
			loan.ReturnedByUserID = &by
		}
		loans = append(loans, loan)
	}
	return batchCreate(s.db, loans)
}
