# Model Scaffolding

The `gux model` command generates complete CRUD scaffolding for your models, including GORM models, DTOs, and admin pages.

## Quick Start

```bash
# Interactive model builder
gux model add

# Generate models from config
gux model add --from-config

# Regenerate a specific model
gux model regen Client
```

## Commands

| Command | Description |
|---------|-------------|
| `gux model add` | Interactive model builder |
| `gux model add <Name>` | Start with model name pre-filled |
| `gux model add --from-config` | Generate all models from gux.config.json |
| `gux model list` | List models defined in config |
| `gux model regen <Name>` | Regenerate files for a model |
| `gux model export <Name>` | Export existing model to config (coming soon) |

## Generated Files

For a model named `Client`, the following files are generated:

```
models/client_gen.go          # GORM model
dto/client_gen.go             # List and Detail DTOs
admin/clients_gen.go          # List page
admin/client_new_gen.go       # Create form
admin/client_detail_gen.go    # Detail view
```

**Note**: Files end with `_gen.go` to indicate they are generated and should not be manually edited.

---

## Interactive Model Builder

Run `gux model add` to start the interactive builder:

```bash
gux model add
```

The wizard guides you through:

1. **Model name** - PascalCase, singular (e.g., `Client`, `Product`)
2. **Fields** - Define each field with type and attributes
3. **Sections** - Group related fields together
4. **Generate** - Create all files automatically

### Field Types

| Type | Description | Input |
|------|-------------|-------|
| `string` | Text field | text, email, tel, url, textarea, select |
| `int`, `uint` | Integer | number |
| `bool` | Boolean | checkbox |
| `float64` | Decimal | number |
| `time.Time` | Date/time | date |
| `*time.Time` | Nullable date | date |
| `*uint` | Foreign key | select (dropdown of related items) |
| `[]string` | Tags/multi-select | multiselect (stored as JSON) |

### Field Attributes

| Attribute | Description |
|-----------|-------------|
| Required | Field is not null in database |
| Show in table | Display in list view |
| Sortable | Column can be sorted |
| Filterable | Can filter by this field |
| Inline editable | Edit directly in table |

### Input Types

For string fields, you can specify the input type:

| Input | Description |
|-------|-------------|
| text | Default text input |
| email | Email validation |
| tel | Phone number |
| url | URL validation |
| textarea | Multi-line text |
| select | Dropdown with options |

---

## Config-Based Models

Models can be defined in `gux.config.json` for version control and team sharing.

### Basic Structure

```json
{
  "module": "github.com/youruser/yourapp",
  "auth": "private",
  "admin": true,
  "models": {
    "Client": {
      "sections": {
        "Contact": [
          { "name": "FirstName", "type": "string", "required": true, "table": true },
          { "name": "LastName", "type": "string", "required": true, "table": true },
          { "name": "Email", "type": "string", "input": "email", "table": true },
          { "name": "Phone", "type": "string", "input": "tel" }
        ],
        "Business": [
          { "name": "Company", "type": "string" },
          { "name": "Notes", "type": "string", "input": "textarea" }
        ]
      }
    }
  }
}
```

### Field Configuration

```json
{
  "name": "Status",           // Field name (PascalCase)
  "type": "string",           // Go type
  "required": true,           // Not null constraint
  "input": "select",          // Input type override
  "label": "Current Status",  // Display label
  "table": true,              // Show in list view
  "priority": 1,              // Column priority (1=high)
  "sortable": true,           // Sortable column
  "filterable": true,         // Can filter
  "editable": true,           // Inline editable
  "default": "active",        // Default value
  "options": [                // For select fields
    { "value": "active", "label": "Active" },
    { "value": "inactive", "label": "Inactive" }
  ]
}
```

---

## Select Fields

### Static Options

Define dropdown options inline:

```json
{
  "name": "Status",
  "type": "string",
  "input": "select",
  "options": [
    { "value": "draft", "label": "Draft" },
    { "value": "active", "label": "Active" },
    { "value": "archived", "label": "Archived" }
  ]
}
```

### Foreign Key Select

Use `*uint` type with a `relation` to create a dropdown of related records:

```json
{
  "name": "SalespersonID",
  "type": "*uint",
  "relation": "User",
  "label": "Salesperson",
  "table": true
}
```

This generates:
- A foreign key field in the model
- A select dropdown populated from the User table
- Proper preloading for display

### Reusable Option Sets

Define options once, use in multiple fields:

```json
{
  "optionSets": {
    "priorities": {
      "low": "Low",
      "medium": "Medium",
      "high": "High",
      "urgent": "Urgent"
    },
    "usStates": {
      "CA": "California",
      "NY": "New York",
      "TX": "Texas"
    }
  },
  "models": {
    "Task": {
      "sections": {
        "Details": [
          { "name": "Priority", "type": "string", "input": "select", "optionsRef": "priorities" }
        ]
      }
    },
    "Client": {
      "sections": {
        "Address": [
          { "name": "State", "type": "string", "input": "select", "optionsRef": "usStates" }
        ]
      }
    }
  }
}
```

---

## Boolean Badges

Display boolean fields as colored badges in the table:

```json
{
  "name": "Active",
  "type": "bool",
  "table": true,
  "badge": {
    "true": "success",      // Badge variant when true
    "false": "secondary",   // Badge variant when false
    "trueLabel": "Active",  // Label when true
    "falseLabel": "Inactive" // Label when false
  }
}
```

Available badge variants: `success`, `secondary`, `warning`, `error`, `primary`

---

## Many-to-Many Relations

For many-to-many relationships (e.g., a Product has many Categories):

```json
{
  "name": "Categories",
  "type": "[]Category",
  "relation": "Category",
  "many2many": "product_categories",
  "input": "multiselect"
}
```

This generates:
- A slice field with GORM many2many tag
- A multi-select input populated from the related table
- Proper join table configuration

---

## Preloads

Specify which relations to preload for API responses:

```json
{
  "models": {
    "Client": {
      "sections": { ... },
      "preloads": ["Salesperson", "Company"]
    }
  }
}
```

---

## Access Control

### Public CRUD

Allow unauthenticated access:

```json
{
  "models": {
    "Product": {
      "public": true,
      "sections": { ... }
    }
  }
}
```

### Role-Based Access

Restrict to specific roles:

```json
{
  "models": {
    "User": {
      "roles": ["admin"],
      "sections": { ... }
    }
  }
}
```

---

## Workflow

### New Model

1. Run `gux model add` or define in config
2. Generate files: `gux model add --from-config`
3. Update API client: `gux gen`
4. Start dev server: `gux dev`
5. Register routes in your app if needed

### Modify Model

1. Edit the model in `gux.config.json`
2. Regenerate: `gux model regen <Name>`
3. Run `gux gen` to update API client
4. Restart with `gux dev`

### Generated DTO Structure

For each model, two DTO types are generated:

**List DTO** - Used for list responses, includes only table-visible fields:
```go
type ClientList struct {
    ID        uint       `json:"id"`
    FirstName string     `json:"first_name"`
    Email     string     `json:"email"`
    Salesperson *UserBrief `json:"salesperson_id" preload:"Salesperson"`
}
```

**Detail DTO** - Used for single-item responses, includes all fields:
```go
type ClientDetail struct {
    ID          uint       `json:"id"`
    FirstName   string     `json:"first_name"`
    LastName    string     `json:"last_name"`
    Email       string     `json:"email"`
    Phone       string     `json:"phone"`
    Description string     `json:"description"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}
```

---

## Example: Complete Model

```json
{
  "models": {
    "Client": {
      "sections": {
        "Contact": [
          { "name": "FirstName", "type": "string", "required": true, "table": true, "priority": 1 },
          { "name": "LastName", "type": "string", "required": true, "table": true, "priority": 1 },
          { "name": "Email", "type": "string", "input": "email", "table": true, "priority": 2 },
          { "name": "Phone", "type": "string", "input": "tel" }
        ],
        "Lead Info": [
          { "name": "SalespersonID", "type": "*uint", "relation": "User", "label": "Salesperson", "table": true, "priority": 2 },
          { "name": "ClosedLead", "type": "bool", "table": true, "badge": { "true": "success", "false": "secondary", "trueLabel": "Closed", "falseLabel": "Open" } }
        ],
        "Business": [
          { "name": "State", "type": "string", "input": "select", "options": [{"value": "CA", "label": "California"}, {"value": "NY", "label": "New York"}, {"value": "TX", "label": "Texas"}] },
          { "name": "Description", "type": "string", "input": "textarea" }
        ]
      },
      "preloads": ["Salesperson"]
    }
  }
}
```

This generates:
- A GORM model with all fields properly typed
- DTOs with foreign key relations mapped to brief structs
- Admin pages with forms, tables, and proper validation
- Automatic preloading for related data
