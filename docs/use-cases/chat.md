# Chat Use Cases

## UC-CM-01 Use Flat Auto-Created Room

- **Primary actor:** Tenant
- **Goal:** Communicate with flatmates in default channel
- **Preconditions:** Tenant has active room assignment
- **Main flow:**
  1. Tenant opens chat module.
  2. System displays flat room channel.
  3. Tenant sends and receives messages.
- **Postconditions:** Messages are stored in flat room history.

## UC-CM-02 Exchange Private Messages

- **Primary actor:** Tenant
- **Goal:** Communicate privately with another tenant
- **Main flow:**
  1. Tenant selects another tenant profile.
  2. System opens or creates private thread.
  3. Tenant sends message.
- **Postconditions:** Thread is available in conversation list.

## UC-CM-03 Create Group Chat

- **Primary actor:** Tenant
- **Goal:** Organize communication for activity/friend group
- **Main flow:**
  1. Tenant creates group chat.
  2. Tenant adds members.
  3. Group messages flow in real time.
- **Postconditions:** Group metadata and membership are persisted.

## UC-CM-04 Operate Chat via HTML Surface

- **Primary actor:** Tenant (staff-stub actor in current implementation)
- **Goal:** Open room timelines, send messages, and update profile in browser
- **Main flow:**
  1. Actor opens `/chat/view`.
  2. Actor selects room or creates group/DM as allowed.
  3. Actor sends messages and edits profile fields.
  4. System persists conversation and membership/profile changes.
- **Postconditions:** Chat data is accessible through both API and HTML surfaces.
