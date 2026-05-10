# Doorman Use Cases

## UC-DM-01 Register Package Arrival

- **Primary actor:** Doorman
- **Goal:** Record incoming package for tenant pickup
- **Preconditions:** Tenant exists in system
- **Main flow:**
  1. Doorman enters the name from package details.
  2. System links package to tenant.
  3. System sets status to received.
  4. System offers notification action.
- **Alternate flows:**
  - Student can't be identified by name, thus no notification is sent out.
- **Postconditions:** Package appears in pending pickup queue.

## UC-DM-02 Validate Guest Entry

- **Primary actor:** Doorman
- **Goal:** Allow only approved guest access
- **Preconditions:** Guest request exists and is valid in time window
- **Main flow:**
  1. Guest arrives and provides identity.
  2. Doorman registers guest to the host tenant.
  3. System confirms validity and logs check-in.
- **Alternate flows:**
  - No out of visit window -> access denied and event logged.
- **Postconditions:** Entry decision is auditable.

## UC-DM-03 Lend and Return Activity Accessories

- **Primary actor:** Doorman
- **Goal:** Track loaned keys/equipment for activities
- **Preconditions:** Item exists and is available
- **Main flow:**
  1. Tenant requests item.
  2. Doorman issues item in lending interface.
  3. System marks item as checked out.
  4. Upon return, doorman closes lending record.
- **Postconditions:** Inventory availability and history are updated.

## UC-DM-04 Validate Access to the Dorm

- **Primary actor:** Doorman
- **Goal:** Allow access only to tenants by default
- **Preconditions:** Tenant is registered
- **Main flow:**
  1. Tenant shows their QR identifier.
  2. The QR code reader reads the code.
  3. Doorman's workspace is updated with the tenant's details.
  4. System confirms and registers the entry.
- **Postconditions:** Every entry is logged and auditable.



