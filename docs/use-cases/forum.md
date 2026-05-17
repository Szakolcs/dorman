# Forum Use Cases

## UC-FM-01 Consume Official and Community Posts

- **Primary actor:** Tenant
- **Goal:** Stay informed via a centralized forum feed
- **Main flow:**
  1. Tenant opens forum feed.
  2. System loads news, events, and announcements.
  3. Tenant opens details for a selected post.
- **Postconditions:** View activity may be recorded for analytics.

## UC-FM-02 Run Community Poll

- **Primary actor:** Student Consult Organizer
- **Goal:** Collect tenant votes for decisions
- **Preconditions:** Poll has options and closing date
- **Main flow:**
  1. Organizer creates poll.
  2. Tenants submit votes.
  3. System validates one vote per tenant (as configured).
  4. System shows results based on visibility rule.
- **Postconditions:** Poll results are persisted and exportable.

## UC-FM-03 Manage Event Participation Intent

- **Primary actor:** Tenant
- **Goal:** Mark planned attendance for event coordination
- **Main flow:**
  1. Tenant opens event post.
  2. Tenant marks going or not going.
  3. System updates counters for organizer dashboard.
- **Postconditions:** Attendance intent is reflected in event metrics.

## UC-FM-04 Operate Forum via HTML Surface

- **Primary actor:** Tenant (or delegated organizer role in v1 staff-stub mode)
- **Goal:** Complete feed, comment, poll, and intent actions from browser pages
- **Main flow:**
  1. Actor opens `/forum/view`.
  2. Actor creates or opens a post in the feed.
  3. Actor submits comment/poll vote/attendance intent forms.
  4. System persists changes and reflects counters/results per policy.
- **Postconditions:** Forum actions are visible in feed and queryable via API.
