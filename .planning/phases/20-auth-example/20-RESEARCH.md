# Phase 20: Auth Example - Research

**Researched:** 2026-01-22
**Domain:** Authentication UI patterns in Gux framework
**Confidence:** HIGH

## Summary

Phase 20 builds an Auth Example application demonstrating core authentication flows: login, registration, password reset, and email verification. The example will use the ui component library (Phases 16-19) and existing Gux patterns from examples/minimal.

The framework already provides strong foundations for auth UIs:
- **Form components**: Input, FormField, Button, Select, Checkbox all ready in ui package
- **State management**: r.StateString for form fields, error/success messages
- **Navigation**: r.Navigate() for post-auth redirects
- **Server infrastructure**: JWT middleware in server/jwt.go, CSRF protection in core/csrf.go, User model with bcrypt password hashing

**Primary recommendation:** Create 5 auth pages as a standalone example app following examples/minimal patterns, using ui components exclusively. Add an Alert component to ui package for auth feedback (success/error messages).

## Standard Stack

The established libraries/tools for this domain:

### Core (from Gux framework)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `ui` package | current | Form components, Button, Card, layout | Phase 16-19 components, tested and ready |
| `core` package | current | Node system, Router, state management | Framework foundation |
| `server` package | current | JWT middleware | Existing authentication infrastructure |
| `models.User` | current | User model with password hashing | Demonstrated in examples/minimal |
| `golang.org/x/crypto/bcrypt` | stdlib | Password hashing | Already used in models/user.go |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `fetch` package | current | HTTP client with CSRF | API calls from WASM |
| GORM | existing | Database ORM | User storage, token storage |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Session cookies | JWT tokens | JWT already implemented in server/jwt.go |
| Custom password reset | Email service | Out of scope per requirements (example only) |

**Installation:**
```bash
# No new dependencies - all components exist in framework
```

## Architecture Patterns

### Recommended Project Structure
```
examples/auth/
├── app.go              # App setup, routes, CRUD
├── models/
│   ├── user.go         # User model (reuse from minimal)
│   └── password_reset.go # Password reset token model
├── dto/
│   └── user.go         # User DTOs (reuse from minimal)
├── pages/
│   ├── nav.go          # Simple nav component
│   ├── login.go        # Login page
│   ├── register.go     # Registration page
│   ├── forgot.go       # Forgot password page
│   ├── reset.go        # Password reset page (with token)
│   └── verify.go       # Email verification landing
└── .gux/               # Generated (api client, wasm)
```

### Pattern 1: Auth Page with Form State
**What:** Standard pattern for auth pages with form validation
**When to use:** All auth pages (login, register, etc.)
**Example:**
```go
// Source: examples/minimal/admin/user_new.go pattern
func Login(r *core.Router) func() core.Node {
    return func() core.Node {
        // Form state
        emailState := r.StateString("email", "")
        passwordState := r.StateString("password", "")
        errorState := r.StateString("error", "")
        loadingState := r.StateBool("loading", false)

        // Submit handler
        submit := func() {
            // Validate
            if emailState.Get() == "" || passwordState.Get() == "" {
                errorState.Set("Email and password are required")
                return
            }
            loadingState.Set(true)

            // API call
            api.Auth.Login(emailState.Get(), passwordState.Get(),
                func(result *dto.User, err error) {
                    loadingState.Set(false)
                    if err != nil {
                        errorState.Set(err.Error())
                        return
                    }
                    // Success - redirect
                    r.Navigate("/dashboard")
                })
        }

        return Card(CardProps{Children: []core.Node{
            // Form fields using ui components
        }})
    }
}
```

### Pattern 2: Centered Auth Layout
**What:** Centered card on full-height page for auth forms
**When to use:** All auth pages
**Example:**
```go
// Source: ui/layout.go Container pattern
func AuthLayout(children ...core.Node) core.Node {
    return core.Div(core.Class("min-h-screen bg-gray-100 dark:bg-gray-900 flex items-center justify-center py-12 px-4"),
        core.Div(core.Class("w-full max-w-md"),
            children...,
        ),
    )
}
```

### Pattern 3: Form Validation Feedback
**What:** Inline errors + form-level alert for auth feedback
**When to use:** All form submissions
**Example:**
```go
// Using FormField with error state + Alert for form-level messages
var alertNode core.Node = core.Frag()
if errorState.Get() != "" {
    alertNode = Alert(AlertProps{
        Variant: AlertError,
        Message: errorState.Get(),
    })
}

FormField(FormFieldProps{
    Label:    "Email",
    LabelFor: "email",
    Required: true,
    Error:    emailError.Get(), // Field-level error
    Children: []core.Node{
        Input(InputProps{
            ID:       "email",
            Type:     InputEmail,
            Name:     "email",
            Value:    emailState.Get(),
            OnChange: func(v string) { emailState.Set(v) },
        }),
    },
})
```

### Pattern 4: Post-Auth Redirect
**What:** Navigate to destination after successful auth
**When to use:** Login success, registration success
**Example:**
```go
// Source: core/app.go Router.Navigate
func (r *Router) Navigate(path string)

// Usage after successful login
r.Navigate("/dashboard")

// With return URL (stored in query param or state)
returnURL := r.StateString("returnURL", "/dashboard")
r.Navigate(returnURL.Get())
```

### Anti-Patterns to Avoid
- **Custom form state management:** Use r.StateString, not manual DOM manipulation
- **Direct js.Value usage in components:** All ui components use core.Node
- **Inline styles:** Use Tailwind classes exclusively
- **Storing passwords in state:** Never expose password values in rendered HTML

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Password hashing | Custom hash | models.User.SetPassword() | bcrypt with proper cost |
| Form field layout | Custom div structure | ui.FormField | Handles label, error, description |
| Button loading state | Manual spinner | Button with disabled + text change | Consistent UX |
| CSRF protection | Manual tokens | Automatic via fetch package | Already built into framework |
| Input validation | Custom regex | HTML5 attributes + JS validation | type="email" gives free validation |

**Key insight:** The auth example should demonstrate the framework, not implement auth infrastructure. Use existing components and patterns.

## Common Pitfalls

### Pitfall 1: Hydration Mismatch with Conditional Rendering
**What goes wrong:** Error/success alerts cause SSR/WASM mismatch if not rendered consistently
**Why it happens:** SSR renders with empty state, WASM may have different state
**How to avoid:** Always render alert container, use CSS visibility or empty content
**Warning signs:** Console errors about hydration mismatch

### Pitfall 2: Password in State Serialization
**What goes wrong:** Password state gets serialized to HTML for hydration
**Why it happens:** r.StateString serializes to __gux_state JSON
**How to avoid:** Don't hydrate password fields - leave them empty on SSR
**Warning signs:** Password visible in page source

### Pitfall 3: Form Submission on SSR
**What goes wrong:** OnClick handlers are no-ops on server
**Why it happens:** Event handlers only work in WASM
**How to avoid:** Forms should work progressively - SSR shows form, WASM enables submission
**Warning signs:** Button clicks do nothing until WASM loads

### Pitfall 4: Redirect Loops
**What goes wrong:** After auth, user keeps getting redirected to login
**Why it happens:** Auth check runs before state/cookie is set
**How to avoid:** Store auth state, check before redirect logic runs
**Warning signs:** Browser shows redirect loop error

### Pitfall 5: Token Expiration Handling
**What goes wrong:** Password reset link shows error after token expires
**Why it happens:** Token TTL not checked before showing form
**How to avoid:** Check token validity in OnLoad, show error if expired
**Warning signs:** User fills form, then gets "token expired" on submit

## Code Examples

Verified patterns from existing codebase:

### Login Page Structure
```go
// Source: examples/minimal/admin/user_new.go pattern
func Login(r *core.Router) func() core.Node {
    return func() core.Node {
        emailState := r.StateString("email", "")
        passwordState := r.StateString("password", "")
        errorState := r.StateString("error", "")

        submit := func() {
            if emailState.Get() == "" || passwordState.Get() == "" {
                errorState.Set("Email and password are required")
                return
            }
            // API call...
        }

        return AuthLayout(
            ui.Card(ui.CardProps{
                Children: []core.Node{
                    ui.CardHeader(ui.CardHeaderProps{
                        Children: []core.Node{
                            core.H2(core.Class("text-xl font-semibold text-center"),
                                core.Text("Sign In")),
                        },
                    }),
                    ui.CardContent(ui.CardContentProps{
                        Children: []core.Node{
                            // Alert for errors
                            conditionalAlert(errorState.Get()),
                            // Form fields
                            ui.FormField(ui.FormFieldProps{
                                Label:    "Email",
                                LabelFor: "email",
                                Required: true,
                                Children: []core.Node{
                                    ui.Input(ui.InputProps{
                                        ID:       "email",
                                        Type:     ui.InputEmail,
                                        Name:     "email",
                                        Value:    emailState.Get(),
                                        OnChange: func(v string) { emailState.Set(v) },
                                        OnEnter:  submit,
                                    }),
                                },
                            }),
                            ui.FormField(ui.FormFieldProps{
                                Label:    "Password",
                                LabelFor: "password",
                                Required: true,
                                Children: []core.Node{
                                    ui.Input(ui.InputProps{
                                        ID:       "password",
                                        Type:     ui.InputPassword,
                                        Name:     "password",
                                        Value:    passwordState.Get(),
                                        OnChange: func(v string) { passwordState.Set(v) },
                                        OnEnter:  submit,
                                    }),
                                },
                            }),
                        },
                    }),
                    ui.CardFooter(ui.CardFooterProps{
                        Children: []core.Node{
                            ui.Button(ui.ButtonProps{
                                Variant:  ui.ButtonPrimary,
                                OnClick:  submit,
                                Class:    "w-full",
                                Children: []core.Node{core.Text("Sign In")},
                            }),
                        },
                    }),
                },
            }),
        )
    }
}
```

### Password Reset Token Handling
```go
// Source: examples/minimal/admin/user_edit.go pattern for route params
func Reset(r *core.Router) func() core.Node {
    var validToken bool
    var userEmail string

    r.OnLoad(func() {
        token := r.GetRouteParams()["token"]
        if token == "" {
            return
        }
        // Verify token via API
        api.Auth.VerifyResetToken(token, func(result *dto.TokenInfo, err error) {
            if err == nil && result != nil {
                validToken = true
                userEmail = result.Email
            }
        })
    })

    return func() core.Node {
        tokenState := r.StateBool("validToken", validToken)

        if !tokenState.Get() {
            return AuthLayout(
                ui.Card(ui.CardProps{
                    Children: []core.Node{
                        // Show "invalid or expired token" message
                    },
                }),
            )
        }

        // Show password reset form
        // ...
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| js.Value components | core.Node components | v2.0 (Phase 16) | Use ui package exclusively |
| Direct GORM in pages | API via generated client | v1.0 | Server-side uses DB, client uses API |

**Deprecated/outdated:**
- components package (uses js.Value directly) - use ui package instead
- Manual CSRF handling - automatic via fetch package

## Open Questions

Things that couldn't be fully resolved:

1. **Alert Component Location**
   - What we know: Alert is needed for auth feedback (success/error messages)
   - What's unclear: Should Alert be added to ui package as part of Phase 20, or is Toast sufficient?
   - Recommendation: Add Alert component to ui package in first plan of Phase 20

2. **Session Storage Pattern**
   - What we know: server/jwt.go provides JWT validation middleware
   - What's unclear: Where/how to store auth token after login (cookie vs localStorage)
   - Recommendation: Use httpOnly cookie for JWT, set by server on successful login

3. **Email Simulation**
   - What we know: Requirements say "example app" - no actual email sending
   - What's unclear: How to simulate email verification flow
   - Recommendation: Show "verification email sent" message, provide direct link to verify page

4. **Return URL Handling**
   - What we know: r.Navigate() works for programmatic navigation
   - What's unclear: Best pattern for "return to original page after login"
   - Recommendation: Use query param ?returnURL=/path, read in login page

## Sources

### Primary (HIGH confidence)
- `/Users/dougbarrett/projects/dbb1dev/goquery/examples/minimal/admin/user_new.go` - Form pattern with state
- `/Users/dougbarrett/projects/dbb1dev/goquery/examples/minimal/admin/user_edit.go` - Route params, OnLoad pattern
- `/Users/dougbarrett/projects/dbb1dev/goquery/ui/*.go` - All Phase 16-19 components
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/app.go` - Router, state management
- `/Users/dougbarrett/projects/dbb1dev/goquery/server/jwt.go` - JWT middleware
- `/Users/dougbarrett/projects/dbb1dev/goquery/models/user.go` - User model with password hashing

### Secondary (MEDIUM confidence)
- `/Users/dougbarrett/projects/dbb1dev/goquery/.planning/REQUIREMENTS.md` - Auth example requirements
- `/Users/dougbarrett/projects/dbb1dev/goquery/.planning/phases/17-form-components/17-VERIFICATION.md` - Form component capabilities

### Tertiary (LOW confidence)
- Training data on auth UI patterns (general web patterns, not framework-specific)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All components and patterns verified in codebase
- Architecture: HIGH - Follows established examples/minimal patterns
- Pitfalls: MEDIUM - Based on framework understanding, not observed issues

**Research date:** 2026-01-22
**Valid until:** 2026-02-22 (stable framework, unlikely to change)

---

## Appendix: Component Checklist

Components needed for Auth Example (all available except Alert):

| Component | Status | Location |
|-----------|--------|----------|
| Button | Ready | ui/button.go |
| Card, CardHeader, CardContent, CardFooter | Ready | ui/card.go |
| Container, VStack, HStack | Ready | ui/layout.go |
| Input (text, email, password) | Ready | ui/input.go |
| FormField | Ready | ui/form.go |
| Checkbox | Ready | ui/checkbox.go |
| Alert | **NEEDED** | Create in ui/alert.go |
| Toast | Ready | ui/toast.go |

## Appendix: Pages to Build

| Page | Route | Components Used |
|------|-------|-----------------|
| Login | `/login` | Card, FormField, Input, Button, Alert, Checkbox |
| Register | `/register` | Card, FormField, Input, Button, Alert |
| Forgot Password | `/forgot` | Card, FormField, Input, Button, Alert |
| Reset Password | `/reset/:token` | Card, FormField, Input, Button, Alert |
| Verify Email | `/verify/:token` | Card, Alert |
