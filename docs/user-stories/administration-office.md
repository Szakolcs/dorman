# Administration and Office User Stories

## AO-000 Tenant registration

- **As a** dorm administrator
- **I want** to register all tenants for the semester
- **So that** I can plan tenant management for the given semester

### Acceptance Criteria

- A room list shows room number, capacity, occupancy, and inventory status.
- Rooms can be filtered by occupied, available, and maintenance-needed status.
- Clicking a room opens a details view with assigned tenants and items.

## AO-001 Room Inventory Overview

- **As a** dorm administrator
- **I want** to view all rooms with occupancy and inventory status
- **So that** I can plan room assignments and procurement

### Acceptance Criteria

- A room list shows room number, capacity, occupancy, and inventory status.
- Rooms can be filtered by occupied, available, and maintenance-needed status.
- Clicking a room opens a details view with assigned tenants and items.

## AO-002 Room Assignment

- **As a** office worker
- **I want** to assign a tenant to an available room
- **So that** move-in workflows are handled quickly

### Acceptance Criteria

- Assignment is blocked if the room is already at full capacity.
- Assignment updates tenant and room occupancy records immediately.
- Assignment history is stored with actor and timestamp.

## AO-003 Maintenance Ticket Creation

- **As a** office worker
- **I want** to create a maintenance ticket for a room or shared area
- **So that** issues are tracked and resolved systematically

### Acceptance Criteria

- Ticket includes category, severity, location, and description.
- Ticket can be assigned to a worker/team and due date.
- Ticket transitions are tracked (open, in progress, resolved, closed).

## AO-003 Maintenance Ticket Approvement

- **As a** office worker
- **I want** to approve a maintenance tickets created by the tenants for a room or shared area
- **So that** issues are tracked and resolved systematically

### Acceptance Criteria

- Tickets created by tenants can be listed and filtered (approved/pending).
- Clicking tickets show their metadata.
- Tickets can be assigned to a worker/team and due date.
- Tickets transitions are tracked (open, in progress, resolved, closed).

## AO-005 Job Scheduling Board

- **As a** administrator
- **I want** to schedule jobs for office and maintenance staff
- **So that** periodic operations are organized and visible

### Acceptance Criteria

- Jobs can be created with assignee, time window, and priority.
- Calendar/list view supports filtering by assignee and date.
- Schedule conflicts are highlighted before saving.

## AO-006 Publish Dorm News

- **As a** office worker
- **I want** to publish news and articles for tenants
- **So that** residents receive official updates in one place

### Acceptance Criteria

- News supports title, body, tags, and publish date.
- Draft and published states are available.
- Published news appears in the tenant forum feed.

## AO-007 Activity Creation

- **As a** administrator
- **I want** to create activity entries (for example ping pong, basketball)
- **So that** tenants can enjoy their free time

### Acceptance Criteria

- Activity has title, location, time, capacity.
- Activity can be linked to required accessories/keys.
- Activity appears in forum/activities after publishing.

## AO-008 Event Creation

- **As a** administrator
- **I want** to create event entries (dorm cup, christmas, midterm party)
- **So that** tenants can indicate, if they want to participate in said events

### Acceptance Criteria

- Event has title, location, time, capacity.
- Event can be reacted to by tenants.
- Events can be canceled or postponed.
- Event appears in forum/events after publishing.

