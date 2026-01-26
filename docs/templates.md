# Templates

This guide provides patterns and examples for building Gux applications with the `core/` framework.

## Complete Example

See [examples/minimal](../examples/minimal) for a complete reference implementation demonstrating:

- **Hybrid rendering** — SSR + WASM hydration
- **Route groups** — Public and admin routes with separate WASM bundles
- **CRUD with DTOs** — Users and Posts with secure field filtering
- **State management** — Server-side data loading with client hydration
- **Security hooks** — Password hashing in create/update hooks

## Page Pattern

Pages follow a loader + component pattern:

```go
import "github.com/dougbarrett/gux/core"

func MyPage(r *core.Router) func() core.Node {
    // LOADER: Runs on server (SSR) and via API for client navigation
    var items []Item

    r.OnLoad(func() {
        // Fetch data from database or API
        if db := r.DB(); db != nil {
            // Server-side: direct database access
            db.(*gorm.DB).Find(&items)
        } else {
            // Client-side: API call
            api.Items.List(func(result []Item, err error) {
                if err == nil {
                    items = result
                }
            })
        }
    })

    // COMPONENT: Returns the UI, re-runs on state changes
    return func() core.Node {
        // Store data in state for hydration
        itemsState := r.StateString("items", serializeItems(items))

        return core.Div(core.Class("container"),
            core.H1(core.Attrs{}, core.Text("Items")),
            renderItemList(parseItems(itemsState.Get())),
        )
    }
}
```

## State Management

```go
func CounterPage(r *core.Router) func() core.Node {
    return func() core.Node {
        // Typed state helpers
        count := r.StateInt("count", 0)
        name := r.StateString("name", "")
        active := r.StateBool("active", false)

        // Generic state for any type
        user := core.UseState(r, "user", User{Name: "Guest"})

        // Read state
        current := count.Get()

        // Update state (triggers re-render)
        onIncrement := func() { count.Set(current + 1) }

        // Update without re-render
        onNameChange := func(value string) { name.SetQuiet(value) }

        return core.Div(core.Class("container"),
            core.H1(core.Attrs{}, core.Text("Counter")),
            core.P(core.Attrs{}, core.Text(fmt.Sprintf("Count: %d", current))),
            core.Button(core.Attrs{OnClick: onIncrement}, core.Text("Increment")),
        )
    }
}
```

## Form Example

```go
func UserForm(r *core.Router) func() core.Node {
    return func() core.Node {
        email := r.StateString("email", "")
        password := r.StateString("password", "")

        onSubmit := func() {
            // Handle form submission
            api.Users.Create(func(user User, err error) {
                if err != nil {
                    // Handle error
                    return
                }
                // Success - redirect or update state
            })
        }

        return core.Form(core.Attrs{OnSubmit: onSubmit},
            core.Div(core.Class("space-y-4"),
                core.Input(core.Attrs{
                    Type:        "email",
                    Placeholder: "Email",
                    Value:       email.Get(),
                    OnChange: func(value string) {
                        email.SetQuiet(value)
                    },
                }),
                core.Input(core.Attrs{
                    Type:        "password",
                    Placeholder: "Password",
                    Value:       password.Get(),
                    OnChange: func(value string) {
                        password.SetQuiet(value)
                    },
                }),
                core.Button(core.Attrs{Type: "submit"}, core.Text("Submit")),
            ),
        )
    }
}
```

## CRUD with DTOs

```go
// In your app setup (app.go)
func SetupApp() *core.App {
    app := core.New()
    app.SetDB(db)

    // Basic CRUD
    app.CRUD(models.Counter{})

    // With DTOs to hide sensitive fields
    app.CRUD(models.User{},
        core.WithListDTO(dto.UserList{}),
        core.WithDetailDTO(dto.UserDetail{}, "Posts"),
    )

    // With security hooks
    app.CRUD(models.User{},
        core.WithCreateHook(func(data map[string]interface{}) (interface{}, error) {
            user := &models.User{}
            user.Email = data["email"].(string)
            if password, ok := data["password"].(string); ok {
                user.SetPassword(password) // Hash before storing
            }
            return user, nil
        }),
    )

    return app
}
```

## Routing

```go
app := core.New()

// Simple routes (SSR only)
app.Routes().
    GET("/", pages.Home).
    POST("/submit", pages.Submit)

// Hybrid routes (SSR + WASM hydration)
app.Routes().
    Hybrid("/", pages.Home).
    Hybrid("/about", pages.About)

// Route groups with separate WASM bundles
app.RouteGroup("/admin", core.WithBundle("admin")).
    Hybrid("/", admin.Dashboard).
    Hybrid("/users", admin.Users).
    Hybrid("/users/:id", admin.UserDetail)

// External links (cross-bundle navigation)
core.A(core.Attrs{Href: "/admin", External: true}, core.Text("Admin"))
```

## Best Practices

1. **Use `r.OnLoad()` for data fetching** — Automatically skipped when hydrated
2. **Serialize complex state** — Use JSON for lists/objects in state
3. **Keep loaders simple** — Move complex logic to separate functions
4. **Use DTOs to hide sensitive data** — Never expose password hashes
5. **Use hooks for validation** — Create/update hooks for business logic
6. **Mark cross-bundle links as External** — Prevents WASM from handling them
7. **Use `SetQuiet()` for input tracking** — Avoids unnecessary re-renders

## Next Steps

- [API Generation](api-generation.md) — Type-safe HTTP clients and servers
- [Server Utilities](server.md) — Middleware and backend helpers
- [Deployment](deployment.md) — Docker and production setup

For complete working examples, explore [examples/minimal](../examples/minimal).
