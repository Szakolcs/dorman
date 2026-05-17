package seed

import (
	"fmt"
	"time"

	adm "dorm-man/internal/models/administration"
	chatm "dorm-man/internal/models/chat"
	fm "dorm-man/internal/models/forum"
	plat "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

func (s *Seeder) seedChat() error {
	// 90 flat rooms
	for _, fl := range s.flats {
		fid := fl.ID
		title := fl.Name + " chat"
		s.flatRooms = append(s.flatRooms, chatm.ChatRoom{
			BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: s.start, UpdatedAt: s.now},
			Kind:      chatm.ChatRoomKindFlat,
			Title:     title,
			FlatID:    &fid,
		})
	}
	if err := batchCreate(s.db, s.flatRooms); err != nil {
		return err
	}

	// 200 direct rooms (400 tenants paired)
	for i := 0; i < 200; i++ {
		a := s.tenants[i*2].ID
		b := s.tenants[i*2+1].ID
		low, high := canonicalPair(a, b)
		s.directRooms = append(s.directRooms, chatm.ChatRoom{
			BaseModel:    plat.BaseModel{ID: uuid.New(), CreatedAt: s.start, UpdatedAt: s.now},
			Kind:         chatm.ChatRoomKindDirect,
			TenantLowID:  &low,
			TenantHighID: &high,
		})
	}
	if err := batchCreate(s.db, s.directRooms); err != nil {
		return err
	}

	allRooms := append(append([]chatm.ChatRoom{}, s.flatRooms...), s.directRooms...)

	var members []chatm.ChatRoomMember
	type memberKey struct{ room, tenant uuid.UUID }
	activeMember := map[memberKey]struct{}{}
	// flat: ~4 active + historical per room
	for _, room := range s.flatRooms {
		added := 0
		for added < 4 {
			t := s.tenants[s.rng.IntN(len(s.tenants))].ID
			k := memberKey{room.ID, t}
			if _, ok := activeMember[k]; ok {
				continue
			}
			activeMember[k] = struct{}{}
			added++
			members = append(members, chatm.ChatRoomMember{
				BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: s.start, UpdatedAt: s.now},
				RoomID:    room.ID,
				TenantID:  t,
				Role:      chatm.ChatRoomMemberRoleMember,
				JoinedAt:  randTime(s.rng, s.start, s.now),
				Source:    chatm.ChatRoomMemberSourceDerived,
			})
		}
		for h := 0; h < 18 && len(members) < 2400; h++ {
			t := s.tenants[s.rng.IntN(len(s.tenants))].ID
			left := randTime(s.rng, s.start, s.now)
			members = append(members, chatm.ChatRoomMember{
				BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: s.start, UpdatedAt: s.now},
				RoomID:    room.ID,
				TenantID:  t,
				Role:      chatm.ChatRoomMemberRoleMember,
				JoinedAt:  left.Add(-30 * 24 * time.Hour),
				LeftAt:    &left,
				Source:    chatm.ChatRoomMemberSourceDerived,
			})
		}
	}
	// direct: 2 each
	for _, room := range s.directRooms {
		if room.TenantLowID != nil {
			members = append(members, chatm.ChatRoomMember{
				BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: s.start, UpdatedAt: s.now},
				RoomID:    room.ID,
				TenantID:  *room.TenantLowID,
				Role:      chatm.ChatRoomMemberRoleMember,
				JoinedAt:  s.start,
				Source:    chatm.ChatRoomMemberSourceDerived,
			})
		}
		if room.TenantHighID != nil {
			members = append(members, chatm.ChatRoomMember{
				BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: s.start, UpdatedAt: s.now},
				RoomID:    room.ID,
				TenantID:  *room.TenantHighID,
				Role:      chatm.ChatRoomMemberRoleMember,
				JoinedAt:  s.start,
				Source:    chatm.ChatRoomMemberSourceDerived,
			})
		}
	}
	if err := batchCreate(s.db, members); err != nil {
		return err
	}

	var messages []chatm.ChatMessage
	for len(messages) < 15000 {
		room := allRooms[s.rng.IntN(len(allRooms))]
		author := s.tenants[s.rng.IntN(len(s.tenants))].ID
		at := randTime(s.rng, s.start, s.now)
		messages = append(messages, chatm.ChatMessage{
			BaseModel:      plat.BaseModel{ID: uuid.New(), CreatedAt: at, UpdatedAt: s.now},
			RoomID:         room.ID,
			AuthorTenantID: author,
			Body:           fmt.Sprintf("Seed message %d", len(messages)+1),
		})
	}
	if err := batchCreate(s.db, messages); err != nil {
		return err
	}

	var syncLogs []chatm.ChatMembershipSyncLog
	for i := 0; i < 600; i++ {
		t := s.tenants[s.rng.IntN(len(s.tenants))].ID
		fl := s.flats[s.rng.IntN(len(s.flats))].ID
		et := "room_assignment.created"
		if s.rng.Float64() < 0.3 {
			et = "room_assignment.ended"
		}
		syncLogs = append(syncLogs, chatm.ChatMembershipSyncLog{
			BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			TenantID:  t,
			FlatID:    fl,
			EventType: et,
			AppliedAt: randTime(s.rng, s.start, s.now),
			Payload:   []byte(`{"action":"seed"}`),
		})
	}
	return batchCreate(s.db, syncLogs)
}

func (s *Seeder) seedForum() error {
	kinds := []fm.ForumPostKind{
		fm.ForumPostKindAnnouncement, fm.ForumPostKindOfficialNews,
		fm.ForumPostKindCommunityActivity, fm.ForumPostKindEvent, fm.ForumPostKindPoll,
	}
	states := []fm.ForumPostState{
		fm.ForumPostStatePublished, fm.ForumPostStateDraft, fm.ForumPostStateArchived,
	}

	for i := 0; i < 800; i++ {
		author := s.staff[s.rng.IntN(len(s.staff))].ID
		var tenantID *uuid.UUID
		if s.rng.Float64() < 0.4 {
			t := s.tenants[s.rng.IntN(len(s.tenants))].ID
			tenantID = &t
		}
		kind := pick(s.rng, kinds)
		st := fm.ForumPostStatePublished
		if s.rng.Float64() < 0.15 {
			st = pick(s.rng, states)
		}
		var pub *time.Time
		if st == fm.ForumPostStatePublished {
			t := randTime(s.rng, s.start, s.now)
			pub = &t
		}
		post := fm.ForumPost{
			BaseModel:      plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			AuthorUserID:   author,
			AuthorTenantID: tenantID,
			Kind:           kind,
			Title:          fmt.Sprintf("Post %d", i+1),
			Body:           "Seed forum post body",
			State:          st,
			Source:         fm.ForumPostSourceForum,
			PublishedAt:    pub,
		}
		if kind == fm.ForumPostKindEvent && len(s.events) > 0 {
			eid := s.events[s.rng.IntN(len(s.events))].ID
			post.EventID = &eid
		}
		s.forumPosts = append(s.forumPosts, post)
		if kind == fm.ForumPostKindPoll {
			s.pollPosts = append(s.pollPosts, post)
		}
	}
	if err := batchCreate(s.db, s.forumPosts); err != nil {
		return err
	}

	var schedules []fm.ForumPostSchedule
	eventPosts := 0
	for _, p := range s.forumPosts {
		if p.Kind != fm.ForumPostKindEvent && p.Kind != fm.ForumPostKindCommunityActivity {
			continue
		}
		if eventPosts >= 250 {
			break
		}
		start := randTime(s.rng, s.start, s.now)
		end := start.Add(2 * time.Hour)
		cap := 50
		schedules = append(schedules, fm.ForumPostSchedule{
			BaseModel:   plat.BaseModel{ID: uuid.New(), CreatedAt: p.CreatedAt, UpdatedAt: s.now},
			ForumPostID: p.ID,
			StartsAt:    &start,
			EndsAt:      &end,
			Location:    "Campus",
			Capacity:    &cap,
		})
		eventPosts++
	}
	if err := batchCreate(s.db, schedules); err != nil {
		return err
	}

	var updates []fm.ForumPostUpdate
	for i := 0; i < 300; i++ {
		p := s.forumPosts[s.rng.IntN(len(s.forumPosts))]
		updates = append(updates, fm.ForumPostUpdate{
			BaseModel:    plat.BaseModel{ID: uuid.New(), CreatedAt: p.CreatedAt, UpdatedAt: s.now},
			ParentPostID: p.ID,
			AuthorUserID: s.staff[s.rng.IntN(len(s.staff))].ID,
			Body:         "Organizer update (seed)",
		})
	}
	if err := batchCreate(s.db, updates); err != nil {
		return err
	}

	// polls: 100
	pollPosts := s.pollPosts
	if len(pollPosts) > 100 {
		pollPosts = pollPosts[:100]
	}
	var polls []fm.ForumPoll
	var options []fm.ForumPollOption
	for _, p := range pollPosts {
		poll := fm.ForumPoll{
			BaseModel:         plat.BaseModel{ID: uuid.New(), CreatedAt: p.CreatedAt, UpdatedAt: s.now},
			ForumPostID:       p.ID,
			ChoiceMode:        fm.ForumPollChoiceModeSingle,
			ResultsVisibility: fm.ForumPollResultsVisibilityLive,
			ClosesAt:          s.now.Add(7 * 24 * time.Hour),
			CreatedByUserID:   p.AuthorUserID,
		}
		polls = append(polls, poll)
		nOpt := 3 + s.rng.IntN(2)
		for o := 0; o < nOpt; o++ {
			options = append(options, fm.ForumPollOption{
				BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: p.CreatedAt, UpdatedAt: s.now},
				PollID:    poll.ID,
				Label:     fmt.Sprintf("Option %c", 'A'+o),
				SortOrder: o,
			})
		}
	}
	if err := batchCreate(s.db, polls); err != nil {
		return err
	}
	if err := batchCreate(s.db, options); err != nil {
		return err
	}

	// rebuild poll->options map
	optsByPoll := map[uuid.UUID][]fm.ForumPollOption{}
	for _, o := range options {
		optsByPoll[o.PollID] = append(optsByPoll[o.PollID], o)
	}

	var votes []fm.ForumPollVote
	for len(votes) < 2500 {
		poll := polls[s.rng.IntN(len(polls))]
		opts := optsByPoll[poll.ID]
		if len(opts) == 0 {
			continue
		}
		opt := opts[s.rng.IntN(len(opts))]
		tenant := s.tenants[s.rng.IntN(len(s.tenants))].ID
		votes = append(votes, fm.ForumPollVote{
			BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: s.now, UpdatedAt: s.now},
			PollID:    poll.ID,
			OptionID:  opt.ID,
			TenantID:  tenant,
			VotedAt:   randTime(s.rng, s.start, s.now),
		})
	}
	// dedupe votes by unique index - use map
	type voteKey struct{ poll, option, tenant uuid.UUID }
	seen := map[voteKey]struct{}{}
	uniqueVotes := make([]fm.ForumPollVote, 0, len(votes))
	for _, v := range votes {
		k := voteKey{v.PollID, v.OptionID, v.TenantID}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		uniqueVotes = append(uniqueVotes, v)
	}
	for len(uniqueVotes) < 2500 {
		poll := polls[s.rng.IntN(len(polls))]
		opts := optsByPoll[poll.ID]
		if len(opts) == 0 {
			continue
		}
		opt := opts[s.rng.IntN(len(opts))]
		tenant := s.tenants[s.rng.IntN(len(s.tenants))].ID
		k := voteKey{poll.ID, opt.ID, tenant}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		uniqueVotes = append(uniqueVotes, fm.ForumPollVote{
			BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: s.now, UpdatedAt: s.now},
			PollID:    poll.ID,
			OptionID:  opt.ID,
			TenantID:  tenant,
			VotedAt:   randTime(s.rng, s.start, s.now),
		})
	}
	if err := batchCreate(s.db, uniqueVotes); err != nil {
		return err
	}

	eventPostIDs := make([]uuid.UUID, 0)
	for _, p := range s.forumPosts {
		if p.Kind == fm.ForumPostKindEvent || p.Kind == fm.ForumPostKindCommunityActivity {
			eventPostIDs = append(eventPostIDs, p.ID)
		}
	}
	var intents []fm.ForumAttendanceIntent
	seenAtt := map[[2]uuid.UUID]struct{}{}
	for len(intents) < 1500 && len(eventPostIDs) > 0 {
		pid := eventPostIDs[s.rng.IntN(len(eventPostIDs))]
		tid := s.tenants[s.rng.IntN(len(s.tenants))].ID
		key := [2]uuid.UUID{pid, tid}
		if _, ok := seenAtt[key]; ok {
			continue
		}
		seenAtt[key] = struct{}{}
		intent := fm.AttendanceIntentGoing
		if s.rng.Float64() < 0.25 {
			intent = fm.AttendanceIntentNotGoing
		}
		intents = append(intents, fm.ForumAttendanceIntent{
			BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: s.now, UpdatedAt: s.now},
			PostID:    pid,
			TenantID:  tid,
			Intent:    intent,
		})
	}
	if err := batchCreate(s.db, intents); err != nil {
		return err
	}

	var comments []fm.ForumComment
	for len(comments) < 3000 {
		p := s.forumPosts[s.rng.IntN(len(s.forumPosts))]
		comments = append(comments, fm.ForumComment{
			BaseModel:        plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			PostID:           p.ID,
			AuthorUserID:     s.staff[s.rng.IntN(len(s.staff))].ID,
			Body:             "Seed comment",
			ModerationState:  fm.CommentModerationStateVisible,
		})
	}
	if err := batchCreate(s.db, comments); err != nil {
		return err
	}

	type reactKey struct {
		user                           uuid.UUID
		tt                             fm.ReactionTargetType
		tid                            uuid.UUID
		rt                             fm.ReactionType
	}
	seenReact := map[reactKey]struct{}{}
	var reactions []fm.ForumReaction
	for len(reactions) < 5000 {
		u := s.users[s.rng.IntN(len(s.users))].ID
		tt := fm.ReactionTargetPost
		var tid uuid.UUID
		if s.rng.Float64() < 0.8 {
			tid = s.forumPosts[s.rng.IntN(len(s.forumPosts))].ID
		} else if len(comments) > 0 {
			tt = fm.ReactionTargetComment
			tid = comments[s.rng.IntN(len(comments))].ID
		} else {
			continue
		}
		k := reactKey{u, tt, tid, fm.ReactionTypeLike}
		if _, ok := seenReact[k]; ok {
			continue
		}
		seenReact[k] = struct{}{}
		reactions = append(reactions, fm.ForumReaction{
			BaseModel:    plat.BaseModel{ID: uuid.New(), CreatedAt: s.now, UpdatedAt: s.now},
			UserID:       u,
			TargetType:   tt,
			TargetID:     tid,
			ReactionType: fm.ReactionTypeLike,
		})
	}
	if err := batchCreate(s.db, reactions); err != nil {
		return err
	}

	var views []fm.ForumPostView
	for len(views) < 20000 {
		p := s.forumPosts[s.rng.IntN(len(s.forumPosts))]
		v := fm.ForumPostView{
			BaseModel: plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			PostID:    p.ID,
			ViewedAt:  randTime(s.rng, s.start, s.now),
		}
		if s.rng.Float64() < 0.5 {
			u := s.users[s.rng.IntN(len(s.users))].ID
			v.ViewerUserID = &u
		} else {
			t := s.tenants[s.rng.IntN(len(s.tenants))].ID
			v.ViewerTenantID = &t
		}
		views = append(views, v)
	}
	if err := batchCreate(s.db, views); err != nil {
		return err
	}

	var mods []fm.ForumModerationAction
	for i := 0; i < 100; i++ {
		p := s.forumPosts[s.rng.IntN(len(s.forumPosts))]
		mods = append(mods, fm.ForumModerationAction{
			BaseModel:       plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			ModeratorUserID: s.staff[0].ID,
			TargetType:      fm.ForumModerationTargetPost,
			TargetID:        p.ID,
			Action:          fm.ForumModerationActionHide,
			OccurredAt:      randTime(s.rng, s.start, s.now),
		})
	}
	return batchCreate(s.db, mods)
}

func (s *Seeder) seedAudit() error {
	actions := []string{
		"tenant.register", "room.assignment", "inventory.create",
		"news.publish", "forum.poll.create", "package.create", "access.denied",
	}
	targets := []string{"tenant", "room_assignment", "forum_post", "package", "inventory_item"}
	var events []adm.AuditEvent
	for i := 0; i < 10000; i++ {
		actor := s.staff[s.rng.IntN(len(s.staff))].ID
		events = append(events, adm.AuditEvent{
			BaseModel:  plat.BaseModel{ID: uuid.New(), CreatedAt: randTime(s.rng, s.start, s.now), UpdatedAt: s.now},
			ActorUserID: &actor,
			Action:     pick(s.rng, actions),
			TargetType: pick(s.rng, targets),
			TargetID:   uuid.New().String(),
			Outcome:    adm.AuditOutcomeSuccess,
			OccurredAt: randTime(s.rng, s.start, s.now),
		})
	}
	return batchCreate(s.db, events)
}
