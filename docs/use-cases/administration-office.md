# Administration and Office Use Cases

## UC-AO-01 Manage Room Allocation

- **Primary actor:** Office Worker
- **Goal:** Assign tenant to room while respecting occupancy constraints
- **Preconditions:** Tenant is active; room exists
- **Main flow:**
  1. User opens room allocation view.
  2. User selects tenant and room.
  3. System validates room capacity.
  4. System persists assignment and updates occupancy.
- **Alternate flows:**
  - Room full -> system blocks assignment and suggests alternatives.
- **Postconditions:** Tenant-room relationship is updated and auditable.

## UC-AO-02 Manage Room Allocation for All

- **Primary actor:** Office Worker
- **Goal:** Assign tenant to room while respecting occupancy constraints
- **Preconditions:** Semester is inactive, Tenants are active; rooms are empty
- **Main flow:**
  1. User opens room allocation view.
  2. User selects the plan room allocation.
  3. System creates the room allocation report.
  4. User can modify the results by swapping tenants.
  5. User approves the allocation report.
  6. System persists assignment and updates occupancy.

- **Postconditions:** Tenant-room relationship is updated and auditable.

## UC-AO-03 Track Maintenance Ticket Lifecycle

- **Primary actor:** Director
- **Goal:** Ensure maintenance requests are resolved with visibility
- **Preconditions:** Ticket created with location and severity
- **Main flow:**
  1. User reviews open tickets.
  2. User assigns ticket to worker/team.
  3. Worker updates status over time.
  4. User closes ticket upon verification.
- **Postconditions:** Ticket history contains timestamps and actors.

## UC-AO-04 Publish Official Dorm News

- **Primary actor:** Office Worker
- **Goal:** Inform tenants about important updates
- **Preconditions:** User has publishing permission
- **Main flow:**
  1. User drafts article/news post.
  2. User sets publish schedule.
  3. System publishes to tenant forum feed.
- **Postconditions:** News is visible in forum and linked to author.

## UC-AO-05 Creates Dorm Activities

- **Primary actor:** Office Worker
- **Goal:** Inform tenants about available free time activities
- **Preconditions:** User has publishing permission
- **Main flow:**
  1. User drafts activity description (title, capacity, availability).
  2. User selects the relevant building.
  3. User links the necessary inventory items (can be none).
  4. System publishes to tenant forum activities.
- **Postconditions:** Activities is visible in forum.

## UC-AO-06 Creates Dorm Event

- **Primary actor:** Office Worker
- **Goal:** Inform tenants about upcoming events
- **Preconditions:** User has publishing permission
- **Main flow:**
  1. User drafts event description (title, capacity, availability).
  2. User selects the relevant building.
  3. System publishes to tenant forum activities.
- **Postconditions:** Event is visible in forum and tenants can react to it.

