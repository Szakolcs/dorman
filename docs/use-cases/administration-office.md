# Administration and Office Use Cases

## UC-AO-00 Manage Tenants

- **Primary actor:** Office Worker
- **Goal:** Administer tenants into the system
- **Preconditions:** Semester is inactive;
- **Main flow:**
  1. User opens the tenants view.
  2. User selects tenants that left the dormatory.
  3. User sets these tenants to inactive.
  4. User selects tenants that are alreadyt registered and stay in the dorm for the next semester.
  5. User sets them to be active for the next semester.
  6. System validates the students' status.
  7. User opens registration view.
  8. User registers the newly arrived tenants.
  9. System validates the new students' status.
  10. System persists assignment and updates occupancy.
- **Postconditions:** Tenants are updated and auditable.

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

## UC-AO-04 Inventory Lifecycle

- **Primary actor:** Director
- **Goal:** Ensure inventory is managed and can be audited
- **Preconditions:** Inventory item is purchased or registered
- **Main flow:**
  1. User lists the inventory items.
  2. User registers any new items purchased.
  3. User generates an inventory report on all the statuses of inventory items.
  4. User matches the report, with the real life audit.
  5. User updates missing, destroyed or withdrawned items.
- **Postconditions:** Inventory history contains timestamps and actors.

## UC-AO-05 Publish Official Dorm News

- **Primary actor:** Office Worker
- **Goal:** Inform tenants about important updates
- **Preconditions:** User has publishing permission
- **Main flow:**
  1. User drafts article/news post.
  2. User sets publish schedule.
  3. System publishes to tenant forum feed.
- **Postconditions:** News is visible in forum and linked to author.

## UC-AO-06 Creates Dorm Activities

- **Primary actor:** Office Worker
- **Goal:** Inform tenants about available free time activities
- **Preconditions:** User has publishing permission
- **Main flow:**
  1. User drafts activity description (title, capacity, availability).
  2. User selects the relevant building.
  3. User links the necessary inventory items (can be none).
  4. System publishes to tenant forum activities.
- **Postconditions:** Activities is visible in forum.

## UC-AO-07 Creates Dorm Event

- **Primary actor:** Office Worker
- **Goal:** Inform tenants about upcoming events
- **Preconditions:** User has publishing permission
- **Main flow:**
  1. User drafts event description (title, capacity, availability).
  2. User selects the relevant building.
  3. System publishes to tenant forum activities.
- **Postconditions:** Event is visible in forum and tenants can react to it.

