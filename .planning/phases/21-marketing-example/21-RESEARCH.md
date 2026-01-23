# Phase 21: Marketing Example - Research

**Researched:** 2026-01-22
**Domain:** Marketing/landing page patterns in Gux framework
**Confidence:** HIGH

## Summary

Phase 21 builds a Marketing Example application demonstrating common marketing website patterns: Home with hero/features/testimonials, Features page, Pricing page with cards, About page, and Contact page with form. The example will use the ui component library (Phases 16-19) and patterns established in the Auth Example (Phase 20).

The framework provides strong foundations for marketing sites:
- **Layout components**: Container, VStack, HStack, Grid, Card all ready in ui package
- **Form components**: Input, Textarea, FormField, Button for contact forms
- **Feedback components**: Alert for form submission feedback
- **State management**: r.StateString for form fields

**Key additions needed**: Marketing-specific layouts (Hero, PricingCard, Footer, responsive Navigation with mobile menu toggle). These can be built as local page components following the AuthLayout pattern from Phase 20.

**Primary recommendation:** Create 5 marketing pages as a standalone example app at `examples/marketing/`. Build marketing-specific layouts as page-local components. Use WASM state for mobile menu toggle (simpler than CSS-only checkbox hack).

## Standard Stack

The established libraries/tools for this domain:

### Core (from Gux framework)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `ui` package | current | Grid, Card, Button, Input, Textarea, FormField, Alert | Phase 16-19 components |
| `core` package | current | Node system, Router, state management | Framework foundation |
| Tailwind CSS | existing | Utility-first styling | Already used throughout framework |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `fetch` package | current | HTTP client with CSRF | Contact form submission |
| core.Section, core.Header, core.Footer | current | Semantic HTML elements | Page structure |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| WASM menu toggle | CSS checkbox hack | CSS-only is complex, requires sibling ordering; WASM state is simpler and matches framework patterns |
| Page-local layouts | ui package components | Marketing layouts are example-specific, not reusable library components |

**Installation:**
```bash
# No new dependencies - all components exist in framework
```

## Architecture Patterns

### Recommended Project Structure
```
examples/marketing/
├── app.go              # App setup, routes
├── pages/
│   ├── layout.go       # MarketingLayout, Nav, Footer, Hero, Section helpers
│   ├── home.go         # Home page with hero, features, testimonials, CTA
│   ├── features.go     # Detailed features page
│   ├── pricing.go      # Pricing cards page
│   ├── about.go        # Team/company story page
│   └── contact.go      # Contact form page
└── .gux/               # Generated (api client, wasm)
```

### Pattern 1: Marketing Page Layout
**What:** Full-page layout with responsive nav and footer
**When to use:** All marketing pages
**Example:**
```go
// Source: examples/auth/pages/layout.go pattern adapted
func MarketingLayout(r *core.Router, children ...core.Node) core.Node {
    return core.Div(core.Class("min-h-screen bg-white dark:bg-gray-900 flex flex-col"),
        Nav(r),
        core.Main(core.Class("flex-1"),
            children...,
        ),
        Footer(),
    )
}
```

### Pattern 2: Responsive Navigation with Mobile Menu
**What:** Desktop horizontal nav, mobile hamburger with slide-out menu
**When to use:** Marketing layout navigation
**Example:**
```go
// Using WASM state for menu toggle (simpler than CSS checkbox hack)
func Nav(r *core.Router) core.Node {
    menuOpen := r.StateBool("menuOpen", false)

    toggleMenu := func() {
        menuOpen.Set(!menuOpen.Get())
    }

    return core.Nav(core.Class("bg-white dark:bg-gray-800 shadow-sm sticky top-0 z-50"),
        core.Div(core.Class("max-w-7xl mx-auto px-4 py-4 flex items-center justify-between"),
            // Logo
            core.A(core.Attrs{Href: "/", Class: "text-xl font-bold text-gray-900 dark:text-white"},
                core.Text("Brand"),
            ),
            // Desktop menu (hidden on mobile)
            core.Div(core.Class("hidden md:flex gap-6"),
                navLinks()...,
            ),
            // Mobile hamburger (hidden on desktop)
            core.Button(core.Attrs{
                Class:   "md:hidden p-2",
                OnClick: toggleMenu,
            }, hamburgerIcon()),
        ),
        // Mobile menu (conditional)
        core.If(menuOpen.Get(),
            core.Div(core.Class("md:hidden bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700"),
                core.Div(core.Class("px-4 py-2 flex flex-col gap-2"),
                    navLinks()...,
                ),
            ),
        ),
    )
}

func navLinks() []core.Node {
    return []core.Node{
        core.A(core.Attrs{Href: "/", Class: "text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"}, core.Text("Home")),
        core.A(core.Attrs{Href: "/features", Class: "text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"}, core.Text("Features")),
        core.A(core.Attrs{Href: "/pricing", Class: "text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"}, core.Text("Pricing")),
        core.A(core.Attrs{Href: "/about", Class: "text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"}, core.Text("About")),
        core.A(core.Attrs{Href: "/contact", Class: "text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"}, core.Text("Contact")),
    }
}
```

### Pattern 3: Hero Section
**What:** Full-width hero with headline, subheadline, CTAs
**When to use:** Home page, landing pages
**Example:**
```go
// Source: examples/auth/pages/home.go hero pattern
func Hero(title, subtitle string, ctaNodes ...core.Node) core.Node {
    return core.Section(core.Class("py-20 px-4 bg-gradient-to-br from-blue-600 to-purple-700 text-white"),
        core.Div(core.Class("max-w-4xl mx-auto text-center"),
            core.H1(core.Class("text-4xl md:text-5xl lg:text-6xl font-bold mb-6"),
                core.Text(title),
            ),
            core.P(core.Class("text-xl md:text-2xl text-blue-100 mb-8 max-w-2xl mx-auto"),
                core.Text(subtitle),
            ),
            core.Div(core.Class("flex flex-col sm:flex-row gap-4 justify-center"),
                ctaNodes...,
            ),
        ),
    )
}
```

### Pattern 4: Feature Section with Grid
**What:** Section header + responsive grid of feature cards
**When to use:** Features page, home page features section
**Example:**
```go
// Source: examples/auth/pages/home.go featureCard pattern + ui/layout.go Grid
func FeaturesSection(title, description string, features []FeatureItem) core.Node {
    featureNodes := make([]core.Node, len(features))
    for i, f := range features {
        featureNodes[i] = featureCard(f.Icon, f.Title, f.Description)
    }

    return core.Section(core.Class("py-20 px-4"),
        core.Div(core.Class("max-w-6xl mx-auto"),
            core.Div(core.Class("text-center mb-12"),
                core.H2(core.Class("text-3xl font-bold text-gray-900 dark:text-white mb-4"),
                    core.Text(title),
                ),
                core.P(core.Class("text-lg text-gray-600 dark:text-gray-400 max-w-2xl mx-auto"),
                    core.Text(description),
                ),
            ),
            ui.Grid(ui.GridProps{
                Cols:     "3",
                Gap:      "8",
                Class:    "md:grid-cols-3 grid-cols-1", // Responsive override
                Children: featureNodes,
            }),
        ),
    )
}

func featureCard(icon, title, description string) core.Node {
    return ui.Card(ui.CardProps{
        Children: []core.Node{
            ui.CardContent(ui.CardContentProps{
                Class: "text-center p-6",
                Children: []core.Node{
                    core.Div(core.Class("text-4xl mb-4"), core.Text(icon)),
                    core.H3(core.Class("text-xl font-semibold text-gray-900 dark:text-white mb-2"),
                        core.Text(title),
                    ),
                    core.P(core.Class("text-gray-600 dark:text-gray-400"),
                        core.Text(description),
                    ),
                },
            }),
        },
    })
}
```

### Pattern 5: Pricing Card
**What:** Pricing tier card with feature list and CTA
**When to use:** Pricing page
**Example:**
```go
func PricingCard(name, price, period, description string, features []string, highlighted bool, ctaText string, ctaAction func()) core.Node {
    cardClass := "relative"
    if highlighted {
        cardClass = "relative ring-2 ring-blue-500 scale-105"
    }

    featureNodes := make([]core.Node, len(features))
    for i, f := range features {
        featureNodes[i] = core.Li(core.Class("flex items-center gap-2"),
            core.Span(core.Class("text-green-500"), core.Text("\u2713")), // Checkmark
            core.Text(f),
        )
    }

    return ui.Card(ui.CardProps{
        Class: cardClass,
        Children: []core.Node{
            core.If(highlighted,
                core.Div(core.Class("absolute -top-3 left-1/2 transform -translate-x-1/2 bg-blue-500 text-white text-sm px-3 py-1 rounded-full"),
                    core.Text("Most Popular"),
                ),
            ),
            ui.CardHeader(ui.CardHeaderProps{
                Class: "text-center",
                Children: []core.Node{
                    core.H3(core.Class("text-xl font-semibold text-gray-900 dark:text-white"),
                        core.Text(name),
                    ),
                    core.Div(core.Class("mt-4"),
                        core.Span(core.Class("text-4xl font-bold text-gray-900 dark:text-white"),
                            core.Text(price),
                        ),
                        core.Span(core.Class("text-gray-500 dark:text-gray-400"),
                            core.Text("/"+period),
                        ),
                    ),
                    core.P(core.Class("text-sm text-gray-600 dark:text-gray-400 mt-2"),
                        core.Text(description),
                    ),
                },
            }),
            ui.CardContent(ui.CardContentProps{
                Children: []core.Node{
                    core.Ul(core.Class("space-y-3 text-gray-600 dark:text-gray-400"),
                        featureNodes...,
                    ),
                },
            }),
            ui.CardFooter(ui.CardFooterProps{
                Children: []core.Node{
                    ui.Button(ui.ButtonProps{
                        Variant: func() ui.ButtonVariant {
                            if highlighted {
                                return ui.ButtonPrimary
                            }
                            return ui.ButtonOutline
                        }(),
                        Class:   "w-full",
                        OnClick: ctaAction,
                        Children: []core.Node{core.Text(ctaText)},
                    }),
                },
            }),
        },
    })
}
```

### Pattern 6: Footer
**What:** Multi-column footer with links and copyright
**When to use:** All marketing pages
**Example:**
```go
func Footer() core.Node {
    return core.Footer(core.Class("bg-gray-900 text-gray-400 py-12"),
        core.Div(core.Class("max-w-6xl mx-auto px-4"),
            ui.Grid(ui.GridProps{
                Cols:  "4",
                Gap:   "8",
                Class: "md:grid-cols-4 grid-cols-2",
                Children: []core.Node{
                    footerColumn("Product", []footerLink{
                        {Text: "Features", Href: "/features"},
                        {Text: "Pricing", Href: "/pricing"},
                        {Text: "Docs", Href: "#"},
                    }),
                    footerColumn("Company", []footerLink{
                        {Text: "About", Href: "/about"},
                        {Text: "Blog", Href: "#"},
                        {Text: "Careers", Href: "#"},
                    }),
                    footerColumn("Support", []footerLink{
                        {Text: "Contact", Href: "/contact"},
                        {Text: "FAQ", Href: "#"},
                        {Text: "Status", Href: "#"},
                    }),
                    footerColumn("Legal", []footerLink{
                        {Text: "Privacy", Href: "#"},
                        {Text: "Terms", Href: "#"},
                    }),
                },
            }),
            core.Div(core.Class("border-t border-gray-800 mt-8 pt-8 text-center text-sm"),
                core.Text("(c) 2026 Brand. All rights reserved."),
            ),
        ),
    )
}
```

### Pattern 7: Contact Form (like Auth forms)
**What:** Contact form with name, email, message fields
**When to use:** Contact page
**Example:**
```go
// Source: examples/auth/pages/login.go form pattern
func Contact(r *core.Router) func() core.Node {
    return func() core.Node {
        nameState := r.StateString("name", "")
        emailState := r.StateString("email", "")
        messageState := r.StateString("message", "")
        errorState := r.StateString("error", "")
        successState := r.StateBool("success", false)
        loadingState := r.StateBool("loading", false)

        submit := func() {
            errorState.Set("")
            if nameState.Get() == "" || emailState.Get() == "" || messageState.Get() == "" {
                errorState.Set("All fields are required")
                return
            }
            loadingState.Set(true)
            // Simulate submission (demo mode)
            loadingState.Set(false)
            successState.Set(true)
        }

        // Success state
        if successState.Get() {
            return MarketingLayout(r,
                // Success message card
            )
        }

        return MarketingLayout(r,
            // Contact form using ui.FormField, ui.Input, ui.Textarea, ui.Button
        )
    }
}
```

### Anti-Patterns to Avoid
- **Complex CSS-only mobile menu:** Use WASM state toggle for simplicity and framework consistency
- **Hardcoded content in layouts:** Pass content as parameters to layout functions
- **Missing responsive classes:** Always include mobile-first responsive variants (e.g., `md:grid-cols-3 grid-cols-1`)
- **Forgetting dark mode:** All Tailwind classes should include dark: variants

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Grid layouts | Custom divs with flexbox | ui.Grid | Handles responsive gaps, columns |
| Card styling | Custom shadow/border classes | ui.Card | Consistent design tokens |
| Form fields | Manual label/input/error structure | ui.FormField | Accessibility, error handling |
| Button variants | Inline classes | ui.Button | Consistent hover, disabled states |
| Multi-line input | Custom textarea styles | ui.Textarea | Proper resize, focus styles |

**Key insight:** Marketing pages should use ui components for all interactive elements. Custom layouts (Hero, Footer, Nav) wrap these components.

## Common Pitfalls

### Pitfall 1: Mobile Menu State on SSR
**What goes wrong:** Menu renders open on server if state initializes to true
**Why it happens:** State default persists to SSR render
**How to avoid:** Initialize menuOpen to false, ensure menu closed by default
**Warning signs:** Mobile menu visible on initial page load before WASM

### Pitfall 2: Missing Responsive Breakpoints
**What goes wrong:** Grid looks bad on mobile (3 tiny columns)
**Why it happens:** Grid cols not responsive
**How to avoid:** Always add `grid-cols-1 md:grid-cols-3` (mobile-first)
**Warning signs:** Squished content on mobile devices

### Pitfall 3: Sticky Nav Z-Index Issues
**What goes wrong:** Nav hides behind hero sections or dropdowns
**Why it happens:** Competing z-index values
**How to avoid:** Use z-50 for nav, ensure hero doesn't have higher z-index
**Warning signs:** Content overlapping navigation

### Pitfall 4: Footer at Bottom
**What goes wrong:** Footer floats in middle of short pages
**Why it happens:** No flex-grow on main content
**How to avoid:** Use flex-col min-h-screen on wrapper, flex-1 on main
**Warning signs:** Floating footer on pages with little content

### Pitfall 5: Contact Form CSRF
**What goes wrong:** Form submission fails with CSRF error
**Why it happens:** Not using fetch package for POST requests
**How to avoid:** Use fetch package for API calls, which automatically includes CSRF
**Warning signs:** 403 Forbidden on form submit

## Code Examples

Verified patterns from existing codebase:

### Home Page Structure
```go
// Source: examples/auth/pages/home.go pattern extended
func Home(r *core.Router) func() core.Node {
    return func() core.Node {
        return MarketingLayout(r,
            // Hero Section
            Hero(
                "Build Amazing Products",
                "The modern framework for web applications",
                ui.Button(ui.ButtonProps{
                    Variant: ui.ButtonPrimary,
                    Size:    ui.ButtonLG,
                    OnClick: func() { r.Navigate("/pricing") },
                    Children: []core.Node{core.Text("Get Started")},
                }),
                ui.Button(ui.ButtonProps{
                    Variant: ui.ButtonOutline,
                    Size:    ui.ButtonLG,
                    Class:   "bg-white/10 border-white text-white hover:bg-white/20",
                    OnClick: func() { r.Navigate("/features") },
                    Children: []core.Node{core.Text("Learn More")},
                }),
            ),

            // Features Section
            FeaturesSection("Why Choose Us", "Everything you need to succeed", []FeatureItem{
                {Icon: "\u26a1", Title: "Fast", Description: "Lightning-fast performance"},
                {Icon: "\ud83d\udd12", Title: "Secure", Description: "Enterprise-grade security"},
                {Icon: "\ud83d\udcca", Title: "Scalable", Description: "Grows with your business"},
            }),

            // Testimonials Section
            TestimonialsSection(),

            // CTA Section
            CTASection("Ready to Get Started?", "Join thousands of happy customers today.", func() { r.Navigate("/contact") }),
        )
    }
}
```

### Pricing Page Structure
```go
func Pricing(r *core.Router) func() core.Node {
    return func() core.Node {
        return MarketingLayout(r,
            core.Section(core.Class("py-20 px-4"),
                core.Div(core.Class("max-w-6xl mx-auto"),
                    // Header
                    core.Div(core.Class("text-center mb-12"),
                        core.H1(core.Class("text-4xl font-bold text-gray-900 dark:text-white mb-4"),
                            core.Text("Simple, Transparent Pricing"),
                        ),
                        core.P(core.Class("text-lg text-gray-600 dark:text-gray-400"),
                            core.Text("Choose the plan that's right for you"),
                        ),
                    ),
                    // Pricing Cards Grid
                    ui.Grid(ui.GridProps{
                        Cols:  "3",
                        Gap:   "8",
                        Class: "md:grid-cols-3 grid-cols-1 items-start",
                        Children: []core.Node{
                            PricingCard("Starter", "$9", "month", "For individuals", []string{
                                "5 projects",
                                "Basic support",
                                "1GB storage",
                            }, false, "Start Free Trial", func() {}),
                            PricingCard("Pro", "$29", "month", "For teams", []string{
                                "Unlimited projects",
                                "Priority support",
                                "10GB storage",
                                "API access",
                            }, true, "Start Free Trial", func() {}),
                            PricingCard("Enterprise", "$99", "month", "For organizations", []string{
                                "Everything in Pro",
                                "Dedicated support",
                                "Unlimited storage",
                                "Custom integrations",
                            }, false, "Contact Sales", func() { r.Navigate("/contact") }),
                        },
                    }),
                ),
            ),
        )
    }
}
```

## SEO Considerations

Marketing pages benefit from SSR-first approach in Gux:

### What Gux Provides Automatically
- Full HTML rendered on server (search engines see complete content)
- WASM hydration for interactivity (doesn't block initial render)
- Semantic HTML elements (core.Section, core.Header, core.Nav, core.Main, core.Footer)

### Best Practices for Marketing Pages
1. **Use semantic elements**: `core.H1`, `core.H2`, `core.Section`, etc. for proper document outline
2. **One H1 per page**: Hero title should be the only H1
3. **Descriptive link text**: "Learn More" with context, not generic "Click Here"
4. **Image alt text**: When adding images, always include alt attributes
5. **Page titles**: Set via `app.SetTitle()` for each route (or implement per-page titles)

### Title/Meta Tags
The framework currently sets a single app title. For production marketing sites, consider:
```go
// Future enhancement: per-page titles
// Current workaround: Use app.SetTitle() with descriptive default
app.SetTitle("Brand - Build Amazing Products")
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| CSS-only checkbox menu | WASM state toggle | Framework pattern | Simpler, more flexible |
| Page-specific components | ui library reuse | Phase 16-19 | Consistent design system |

**Deprecated/outdated:**
- Complex CSS-only interactions - use WASM state for interactive elements
- Inline styles - use Tailwind classes exclusively

## Open Questions

Things that couldn't be fully resolved:

1. **Per-Page SEO Titles**
   - What we know: `app.SetTitle()` sets global title
   - What's unclear: Best pattern for page-specific titles in SSR
   - Recommendation: Use descriptive global title; per-page titles are a future enhancement

2. **Image Handling**
   - What we know: Marketing sites need hero images, team photos, logos
   - What's unclear: Best pattern for static assets in Gux examples
   - Recommendation: Use placeholder text/emoji for demo; real images are out of scope

3. **Contact Form Backend**
   - What we know: Example app, no real email service
   - What's unclear: How to demonstrate form submission
   - Recommendation: Show success message immediately (demo mode), similar to Auth example password reset

4. **Testimonials Data**
   - What we know: Marketing sites show testimonials
   - What's unclear: Should testimonials be hardcoded or from API?
   - Recommendation: Hardcode in page component (it's an example, not a CMS)

## Sources

### Primary (HIGH confidence)
- `/Users/dougbarrett/projects/dbb1dev/goquery/examples/auth/pages/layout.go` - Layout patterns (AuthLayout, Nav, PageLayout)
- `/Users/dougbarrett/projects/dbb1dev/goquery/examples/auth/pages/home.go` - Hero pattern, feature cards
- `/Users/dougbarrett/projects/dbb1dev/goquery/examples/auth/pages/login.go` - Form patterns with state
- `/Users/dougbarrett/projects/dbb1dev/goquery/ui/*.go` - All Phase 16-19 components
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/elements.go` - Semantic HTML elements
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/app.go` - Router, state management

### Secondary (MEDIUM confidence)
- [LogRocket - CSS-only mobile menu](https://blog.logrocket.com/create-responsive-mobile-menu-css-without-javascript/) - Checkbox hack patterns
- [DEV Community - Tailwind mobile menu](https://dev.to/seppegadeyne/crafting-a-mobile-menu-with-tailwind-css-without-javascript-1814) - CSS peer selector approach
- [Design Studio UIUX - SaaS Pricing Page Best Practices](https://www.designstudiouiux.com/blog/saas-pricing-page-design-best-practices/) - Pricing card patterns
- [Tailwind CSS Heroes](https://tailwindcss.com/plus/ui-blocks/marketing/sections/heroes) - Hero section patterns

### Tertiary (LOW confidence)
- Training data on marketing page patterns (general web patterns)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All components verified in codebase
- Architecture: HIGH - Follows established auth example patterns
- Pitfalls: MEDIUM - Based on framework understanding and common marketing site issues

**Research date:** 2026-01-22
**Valid until:** 2026-02-22 (stable framework, unlikely to change)

---

## Appendix: Component Checklist

Components needed for Marketing Example (all available):

| Component | Status | Location | Usage |
|-----------|--------|----------|-------|
| Button | Ready | ui/button.go | CTAs, form submit |
| Card, CardHeader, CardContent, CardFooter | Ready | ui/card.go | Feature cards, pricing cards |
| Container | Ready | ui/layout.go | Page containers |
| Grid | Ready | ui/layout.go | Feature grids, pricing grids, footer columns |
| VStack, HStack | Ready | ui/layout.go | Stacked layouts |
| Input | Ready | ui/input.go | Contact form |
| Textarea | Ready | ui/textarea.go | Contact form message |
| FormField | Ready | ui/form.go | Contact form fields |
| Alert | Ready | ui/alert.go | Form feedback |

## Appendix: Pages to Build

| Page | Route | Key Sections |
|------|-------|--------------|
| Home | `/` | Hero, Features grid, Testimonials, CTA |
| Features | `/features` | Feature sections with details |
| Pricing | `/pricing` | 3 pricing tier cards |
| About | `/about` | Company story, team section |
| Contact | `/contact` | Contact form with name, email, message |

## Appendix: Custom Layout Components (Page-Local)

These should be created in `pages/layout.go`:

| Component | Purpose | Based On |
|-----------|---------|----------|
| MarketingLayout | Full page wrapper with nav + footer | AuthLayout/PageLayout |
| Nav | Responsive navigation with mobile menu | Auth Nav pattern |
| Footer | Multi-column footer | Custom for marketing |
| Hero | Hero section with title, subtitle, CTAs | Auth home hero pattern |
| Section | Generic section wrapper with padding | Common pattern |
| PricingCard | Pricing tier card | Card + custom structure |
