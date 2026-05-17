# Tenant Chat User Stories

## CM-001 Automatic Flat Room

- **As a** tenant
- **I want** to be automatically added to my flat chat room
- **So that** I can coordinate with roommates immediately

### Acceptance Criteria

- One system-managed flat chat exists per flat (see [glossary](../specifications/glossary.md)).
- **Membership rule:** a tenant is in the flat chat if and only if they have an **active** room assignment to any room in that flat; add/remove follows assignment create/update/end (not a separate manual join list).
- The flat chat row may be created when the flat is provisioned, or lazily when the first room in that flat receives an assignment (product picks one approach and documents it in MVP).
- Flat room membership is not manually editable by non-admin users.

## CM-002 Private Messaging

- **As a** tenant
- **I want** to send private messages to other tenants
- **So that** I can communicate one-to-one

### Acceptance Criteria

- Private chats can be initiated from tenant profile/search.
- Message history persists and loads in chronological order.
- Unread count is shown until conversation is opened.

## CM-003 Group Chats

- **As a** tenant
- **I want** to create group chats for activities or friends
- **So that** I can organize conversations around shared interests

### Acceptance Criteria

- Group creator can add/remove participants.
- Group supports name and avatar customization.
- Leave-group action preserves message history for remaining members.

## CM-004 Profile Personalization

- **As a** tenant
- **I want** to edit avatar, nickname, and bio
- **So that** my profile reflects my identity

### Acceptance Criteria

- Default nickname is full name from tenant record.
- Avatar upload validates type/size constraints.
- Profile updates propagate to chat headers and member lists.

## CM-005 Real-Time Message Delivery

- **As a** tenant
- **I want** messages to appear in near real-time
- **So that** chats feel responsive and reliable

### Acceptance Criteria

- New messages appear without full-page reload.
- Temporary delivery failure surfaces a retry state.
- Read receipts/status indicators are updated per conversation.
