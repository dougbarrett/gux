# Audit Logging

Track who changed what and when for regulatory compliance. Gux provides automatic audit logging for any CRUD model with field-level change diffs.

## Quick Start

Enable audit logging on any CRUD model with `WithAuditLog()`:

```go
// Audit all fields
app.CRUD(models.Document{}, core.WithAuditLog())

// Ignore sensitive fields in audit diffs
app.CRUD(models.User{},
    core.WithAuditLog("PasswordHash"),
    core.WithRoles("admin"),
)
```

The `audit_entries` table is auto-migrated when any model enables audit logging — no manual migration needed.

## What Gets Logged

| Action | Changes Field |
|--------|--------------|
| **Create** | Full snapshot of created entity (minus ignored fields) |
| **Update** | Field-level diff: `{"name": {"old": "Alice", "new": "Bob"}}` |
| **Delete** | Entity type and ID recorded; changes is `{}` |

Each audit entry also captures:
- **User ID and email** (from the authenticated session, if available)
- **IP address** (respects `X-Forwarded-For` and `X-Real-IP` headers)
- **Timestamp** (immutable, append-only)

## Configuration

`WithAuditLog()` accepts optional field names to exclude from diffs:

```go
// Exclude password hash and internal tokens
core.WithAuditLog("PasswordHash", "APIToken", "InternalSecret")
```

`UpdatedAt` and `DeletedAt` are always excluded from diffs automatically.

## AuditEntry Schema

```go
type AuditEntry struct {
    ID         uint        `json:"id"`
    CreatedAt  time.Time   `json:"created_at"`
    Action     AuditAction `json:"action"`      // "create", "update", "delete"
    EntityType string      `json:"entity_type"` // Model name (e.g., "User")
    EntityID   uint        `json:"entity_id"`
    UserID     string      `json:"user_id"`
    UserEmail  string      `json:"user_email"`
    IPAddress  string      `json:"ip_address"`
    Changes    string      `json:"changes"`     // JSON
}
```

## Viewing Audit Logs

Register `AuditEntry` as a read-only CRUD endpoint:

```go
app.CRUD(core.AuditEntry{},
    core.WithRoles("admin"),
)
```

This exposes `GET /__gux_api/crud/auditentries` for listing all entries, which you can display in an admin page using a `DataTable`.

## Behavior

- **Auto-migration**: The `audit_entries` table is created automatically when any model enables audit
- **Async logging**: Audit writes happen in a goroutine — they never block the HTTP response
- **Failure handling**: Write failures are logged to stderr but never cause request errors
- **Public endpoints**: Entries are created with empty `UserID`/`UserEmail` fields
- **Field exclusion**: `UpdatedAt` and `DeletedAt` are always excluded from diffs; add custom fields via `WithAuditLog("FieldName")`

## Example

See `examples/audit/` for a complete document management application with:
- CRUD for Users and Documents with audit logging enabled
- Admin page showing the full audit trail with action badges, user info, and change diffs
- Filter by action type and entity type
- Seed data demonstrating create and update audit entries
