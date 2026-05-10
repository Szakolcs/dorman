# Doorman Module User Stories

## DM-001 Package Check-In

- **As a** doorman
- **I want** to register incoming packages by tenant
- **So that** deliveries are traceable and easy to hand over

### Acceptance Criteria

- Package entry stores tenant (by name), arrival time, and package type.
- Package status transitions: received, notified (optional), picked up.

## DM-002 Tenant Pickup Notification

- **As a** doorman
- **I want** to notify tenants when packages arrive
- **So that** pickup delays are reduced

### Acceptance Criteria

- Notification can be sent from the package record from tenants full name.
- If the tenant can't be identified certainly, no notification will be sent out.
- Notification log stores channel, timestamp.
- Tenant can see notification in their app feed.

## DM-003 Guest Access Management

- **As a** doorman
- **I want** to register and validate tenant guests
- **So that** building access is secure and auditable

### Acceptance Criteria

- Guest entry requires host tenant, guest name, and visit window.
- Check-in/check-out timestamps are recorded.
- Expired guest approvals are automatically denied.

## DM-004 QR Access Verification

- **As a** doorman
- **I want** to verify tenant QR codes at entry
- **So that** only authorized residents enter the dorm

### Acceptance Criteria

- Valid code grants access and logs entry time.
- Invalid/expired code is denied and logged.
- Access logs can be filtered by tenant and date.

## DM-005 Accessory Lending

- **As a** doorman
- **I want** to lend keys and activity accessories to tenants
- **So that** shared resources are tracked and returned

### Acceptance Criteria

- Lending record captures item, tenant, checkout time, expected return.
- Return action updates item availability status.
- Overdue items are flagged in a dedicated view.
