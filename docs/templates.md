# Templates

This guide provides complete, ready-to-use templates for common application patterns. Each template demonstrates Gux best practices and can be adapted for your needs.

## Single Page Marketing Site

A clean, conversion-focused landing page with hero section, features, and call-to-action.

```go
//go:build js && wasm

package main

import (
    "syscall/js"
    "github.com/dougbarrett/gux/components"
)

func main() {
    app := components.NewApp("app")
    components.InitToasts()

    // Build the marketing page
    page := components.Div("min-h-screen",
        navbar(),
        hero(),
        features(),
        testimonials(),
        cta(),
        footer(),
    )

    app.Mount(page)
    app.Run()
}

func navbar() js.Value {
    return components.Div("fixed top-0 left-0 right-0 z-50 surface-base border-b border-subtle",
        components.Div("max-w-6xl mx-auto px-4 py-4 flex justify-between items-center",
            components.HeadingWithClass(1, "Acme", "text-2xl font-bold text-primary"),
            components.Div("flex items-center gap-6",
                components.Link("#features", "Features", "text-secondary hover:text-primary"),
                components.Link("#testimonials", "Testimonials", "text-secondary hover:text-primary"),
                components.Link("#pricing", "Pricing", "text-secondary hover:text-primary"),
                components.PrimaryButton("Get Started", func() {
                    components.Toast("Welcome aboard!", components.ToastSuccess)
                }),
            ),
        ),
    )
}

func hero() js.Value {
    return components.Div("pt-32 pb-20 px-4",
        components.Div("max-w-4xl mx-auto text-center",
            components.HeadingWithClass(1, "Build Better Apps Faster",
                "text-5xl md:text-6xl font-bold text-primary mb-6"),
            components.TextWithClass("The modern framework for building lightning-fast web applications with Go and WebAssembly.",
                "text-xl text-secondary mb-8 max-w-2xl mx-auto"),
            components.Div("flex justify-center gap-4",
                components.Button(components.ButtonProps{
                    Text:      "Start Free Trial",
                    Variant:   components.ButtonPrimary,
                    Size:      components.ButtonLG,
                    ClassName: "px-8",
                    OnClick:   func() { /* navigate to signup */ },
                }),
                components.Button(components.ButtonProps{
                    Text:      "View Demo",
                    Variant:   components.ButtonSecondary,
                    Size:      components.ButtonLG,
                    ClassName: "px-8",
                    OnClick:   func() { /* show demo */ },
                }),
            ),
        ),
    )
}

func features() js.Value {
    featureItems := []struct {
        icon, title, desc string
    }{
        {"🚀", "Lightning Fast", "Compiled to WebAssembly for near-native performance in the browser."},
        {"🔒", "Type Safe", "Full Go type safety catches errors at compile time, not runtime."},
        {"📦", "Small Bundles", "TinyGo builds produce bundles under 500KB gzipped."},
        {"🔄", "Real-time", "Built-in WebSocket support for live updates and collaboration."},
        {"🎨", "Beautiful UI", "45+ production-ready components with Tailwind styling."},
        {"🛠️", "Great DX", "Hot reload, code generation, and powerful CLI tools."},
    }

    cards := make([]js.Value, len(featureItems))
    for i, f := range featureItems {
        cards[i] = components.Card(
            components.Div("text-center p-4",
                components.TextWithClass(f.icon, "text-4xl mb-4"),
                components.HeadingWithClass(3, f.title, "text-lg font-semibold text-primary mb-2"),
                components.TextWithClass(f.desc, "text-secondary"),
            ),
        )
    }

    return components.Div("py-20 px-4 surface-overlay",
        components.Div("max-w-6xl mx-auto",
            components.HeadingWithClass(2, "Everything You Need",
                "text-3xl font-bold text-primary text-center mb-12"),
            components.Div("grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6", cards...),
        ),
    )
}

func testimonials() js.Value {
    quotes := []struct {
        text, author, role string
    }{
        {"Gux transformed how we build web apps. The productivity gains are incredible.",
            "Sarah Chen", "CTO at TechCorp"},
        {"Finally, a framework that lets our Go backend team build frontends too.",
            "Marcus Johnson", "Lead Engineer at StartupXYZ"},
        {"The type safety alone has saved us countless hours of debugging.",
            "Emily Rodriguez", "Senior Developer at BigCo"},
    }

    cards := make([]js.Value, len(quotes))
    for i, q := range quotes {
        cards[i] = components.Card(
            components.Div("p-6",
                components.TextWithClass("\""+q.text+"\"", "text-primary italic mb-4"),
                components.Div("flex items-center gap-3",
                    components.Avatar(components.AvatarProps{Name: q.author, Size: components.AvatarMD}),
                    components.Div("",
                        components.TextWithClass(q.author, "font-semibold text-primary"),
                        components.TextWithClass(q.role, "text-sm text-secondary"),
                    ),
                ),
            ),
        )
    }

    return components.Div("py-20 px-4",
        components.Div("max-w-6xl mx-auto",
            components.HeadingWithClass(2, "Loved by Developers",
                "text-3xl font-bold text-primary text-center mb-12"),
            components.Div("grid grid-cols-1 md:grid-cols-3 gap-6", cards...),
        ),
    )
}

func cta() js.Value {
    return components.Div("py-20 px-4 bg-gradient-to-r from-blue-600 to-purple-600",
        components.Div("max-w-4xl mx-auto text-center",
            components.HeadingWithClass(2, "Ready to Get Started?",
                "text-3xl font-bold text-white mb-4"),
            components.TextWithClass("Join thousands of developers building with Gux.",
                "text-white/80 mb-8"),
            components.Button(components.ButtonProps{
                Text:      "Start Building Today",
                Variant:   components.ButtonSecondary,
                Size:      components.ButtonLG,
                ClassName: "px-8 bg-white text-blue-600 hover:bg-gray-100",
                OnClick:   func() { /* navigate to signup */ },
            }),
        ),
    )
}

func footer() js.Value {
    return components.Div("py-8 px-4 surface-base border-t border-subtle",
        components.Div("max-w-6xl mx-auto flex justify-between items-center",
            components.TextWithClass("© 2024 Acme Inc. All rights reserved.", "text-secondary"),
            components.Div("flex gap-6",
                components.Link("#", "Privacy", "text-secondary hover:text-primary"),
                components.Link("#", "Terms", "text-secondary hover:text-primary"),
                components.Link("#", "Contact", "text-secondary hover:text-primary"),
            ),
        ),
    )
}
```

## Admin Layout

A full-featured admin dashboard with sidebar navigation, header with user menu, and content area. This pattern is used in the [example app](https://github.com/dougbarrett/gux/tree/main/example).

```go
//go:build js && wasm

package main

import (
    "syscall/js"
    "github.com/dougbarrett/gux/components"
)

var (
    app    *components.App
    layout *components.Layout
    router *components.Router
)

func main() {
    app = components.NewApp("app")
    components.InitToasts()

    // Setup router
    router = components.NewRouter()
    components.SetGlobalRouter(router)

    // Create user menu
    userMenu := components.NewUserMenu(components.UserMenuProps{
        Name:  "Admin User",
        Email: "admin@example.com",
        OnProfile: func() {
            router.Navigate("/profile")
        },
        OnSettings: func() {
            router.Navigate("/settings")
        },
        OnLogout: func() {
            components.Toast("Logged out", components.ToastSuccess)
        },
    })

    // Create notification center
    notificationCenter := components.NewNotificationCenter(components.NotificationCenterProps{
        Notifications: []components.Notification{
            {ID: "1", Title: "New Order", Message: "Order #1234 received", Time: "2 min ago", Type: "success"},
            {ID: "2", Title: "Low Stock", Message: "Widget inventory below threshold", Time: "1 hour ago", Type: "warning"},
        },
        OnMarkRead:    func(id string) { /* mark read */ },
        OnMarkAllRead: func() { /* mark all read */ },
    })

    // Create layout with sidebar and header
    layout = components.NewLayout(components.LayoutProps{
        Sidebar: components.SidebarProps{
            Title: "Admin Panel",
            Items: []components.NavItem{
                {Label: "Dashboard", Icon: "📊", Path: "/"},
                {Label: "Users", Icon: "👥", Path: "/users"},
                {Label: "Products", Icon: "📦", Path: "/products"},
                {Label: "Orders", Icon: "🛒", Path: "/orders"},
                {Label: "Analytics", Icon: "📈", Path: "/analytics"},
                {Label: "Settings", Icon: "⚙️", Path: "/settings"},
            },
        },
        Header: components.HeaderProps{
            Title:              "Dashboard",
            NotificationCenter: notificationCenter,
            UserMenu:           userMenu,
            Actions: []components.HeaderAction{
                {Label: "Refresh", OnClick: func() {
                    js.Global().Get("location").Call("reload")
                }},
            },
        },
    })

    // Register routes
    router.Register("/", showDashboard)
    router.Register("/users", showUsers)
    router.Register("/products", showProducts)
    router.Register("/orders", showOrders)
    router.Register("/analytics", showAnalytics)
    router.Register("/settings", showSettings)

    // Update sidebar active state on navigation
    router.OnNavigate(func(path string) {
        layout.Sidebar().SetActive(path)
    })

    // Register keyboard shortcut for sidebar collapse (Cmd/Ctrl+B)
    layout.Sidebar().RegisterKeyboardShortcut()

    app.Mount(layout.Element())
    router.Start()
    app.Run()
}

func showDashboard() {
    layout.SetContent(
        components.Div("space-y-6",
            // Stats row
            components.Div("grid grid-cols-1 md:grid-cols-4 gap-4",
                statCard("Total Revenue", "$45,231", "+20.1%", components.BadgeSuccess),
                statCard("Orders", "2,350", "+12.5%", components.BadgeSuccess),
                statCard("Customers", "1,247", "+8.2%", components.BadgePrimary),
                statCard("Avg Order", "$36.28", "-3.1%", components.BadgeWarning),
            ),
            // Recent activity
            components.TitledCard("Recent Activity",
                "Latest orders and updates from your store.",
            ),
        ),
    )
}

func statCard(label, value, change string, variant components.BadgeVariant) js.Value {
    return components.Card(
        components.Div("p-2",
            components.Div("flex justify-between items-start",
                components.TextWithClass(label, "text-sm text-secondary"),
                components.Badge(components.BadgeProps{Text: change, Variant: variant, Rounded: true}),
            ),
            components.HeadingWithClass(3, value, "text-2xl font-bold text-primary mt-2"),
        ),
    )
}

func showUsers() {
    table := components.NewTable(components.TableProps{
        Columns: []components.TableColumn{
            {Header: "Name", Key: "name", Sortable: true},
            {Header: "Email", Key: "email", Sortable: true},
            {Header: "Role", Key: "role"},
            {Header: "Status", Key: "status", Render: func(row map[string]any, value any) js.Value {
                status := value.(string)
                variant := components.BadgeSuccess
                if status == "Inactive" {
                    variant = components.BadgeSecondary
                }
                return components.Badge(components.BadgeProps{Text: status, Variant: variant})
            }},
        },
        Data: []map[string]any{
            {"name": "John Doe", "email": "john@example.com", "role": "Admin", "status": "Active"},
            {"name": "Jane Smith", "email": "jane@example.com", "role": "Editor", "status": "Active"},
            {"name": "Bob Wilson", "email": "bob@example.com", "role": "Viewer", "status": "Inactive"},
        },
        Striped:    true,
        Hoverable:  true,
        Filterable: true,
        Paginated:  true,
        PageSize:   10,
    })

    layout.SetPage("Users", "Manage user accounts and permissions.",
        components.Div("mt-4",
            components.Div("flex justify-end mb-4",
                components.PrimaryButton("Add User", func() {
                    components.Toast("Add user modal", components.ToastInfo)
                }),
            ),
            table.Element(),
        ),
    )
}

func showProducts()  { layout.SetPage("Products", "Manage your product catalog.", components.Text("Products content")) }
func showOrders()    { layout.SetPage("Orders", "View and manage orders.", components.Text("Orders content")) }
func showAnalytics() { layout.SetPage("Analytics", "View business insights.", components.Text("Analytics content")) }
func showSettings()  { layout.SetPage("Settings", "Configure your account.", components.Text("Settings content")) }
```

## Product Landing Page

A focused product page with pricing tiers and feature comparison.

```go
//go:build js && wasm

package main

import (
    "syscall/js"
    "github.com/dougbarrett/gux/components"
)

func main() {
    app := components.NewApp("app")
    components.InitToasts()

    page := components.Div("min-h-screen surface-base",
        productHero(),
        pricing(),
        featureComparison(),
        faq(),
    )

    app.Mount(page)
    app.Run()
}

func productHero() js.Value {
    return components.Div("py-20 px-4",
        components.Div("max-w-6xl mx-auto grid grid-cols-1 lg:grid-cols-2 gap-12 items-center",
            // Left: Product info
            components.Div("",
                components.Badge(components.BadgeProps{
                    Text:    "New Release",
                    Variant: components.BadgePrimary,
                }),
                components.HeadingWithClass(1, "The Ultimate Developer Tool",
                    "text-4xl md:text-5xl font-bold text-primary mt-4 mb-6"),
                components.TextWithClass(
                    "Streamline your workflow with our powerful development platform. Build, test, and deploy faster than ever.",
                    "text-lg text-secondary mb-8"),
                components.Div("flex flex-wrap gap-4",
                    components.Button(components.ButtonProps{
                        Text:    "Start Free Trial",
                        Variant: components.ButtonPrimary,
                        Size:    components.ButtonLG,
                        OnClick: func() { /* signup */ },
                    }),
                    components.Button(components.ButtonProps{
                        Text:    "Watch Demo",
                        Variant: components.ButtonGhost,
                        Size:    components.ButtonLG,
                        OnClick: func() { /* demo modal */ },
                    }),
                ),
                // Trust badges
                components.Div("mt-8 flex items-center gap-6",
                    components.TextWithClass("Trusted by:", "text-sm text-secondary"),
                    components.AvatarGroup([]components.AvatarProps{
                        {Name: "Google"},
                        {Name: "Meta"},
                        {Name: "Amazon"},
                        {Name: "Netflix"},
                    }, 4),
                ),
            ),
            // Right: Product screenshot/preview
            components.Div("surface-overlay rounded-xl p-8 border border-subtle",
                components.TextWithClass("[Product Screenshot]", "text-center text-secondary py-32"),
            ),
        ),
    )
}

func pricing() js.Value {
    plans := []struct {
        name, price, desc string
        features          []string
        popular           bool
    }{
        {
            name: "Starter", price: "$9", desc: "Perfect for side projects",
            features: []string{"5 projects", "1GB storage", "Community support", "Basic analytics"},
        },
        {
            name: "Pro", price: "$29", desc: "For professional developers",
            features: []string{"Unlimited projects", "10GB storage", "Priority support", "Advanced analytics", "Team collaboration", "Custom domains"},
            popular: true,
        },
        {
            name: "Enterprise", price: "$99", desc: "For large teams",
            features: []string{"Everything in Pro", "100GB storage", "Dedicated support", "SLA guarantee", "SSO/SAML", "Audit logs"},
        },
    }

    cards := make([]js.Value, len(plans))
    for i, plan := range plans {
        featureItems := make([]js.Value, len(plan.features))
        for j, f := range plan.features {
            featureItems[j] = components.Div("flex items-center gap-2",
                components.TextWithClass("✓", "text-green-500"),
                components.Text(f),
            )
        }

        cardClass := "p-6 rounded-xl border"
        if plan.popular {
            cardClass += " border-blue-500 ring-2 ring-blue-500/20"
        } else {
            cardClass += " border-subtle"
        }

        var popularBadge js.Value
        if plan.popular {
            popularBadge = components.Badge(components.BadgeProps{
                Text: "Most Popular", Variant: components.BadgePrimary,
            })
        } else {
            popularBadge = components.Div("h-6") // Spacer
        }

        cards[i] = components.Div(cardClass,
            popularBadge,
            components.HeadingWithClass(3, plan.name, "text-xl font-bold text-primary mt-4"),
            components.Div("flex items-baseline gap-1 mt-2",
                components.HeadingWithClass(2, plan.price, "text-4xl font-bold text-primary"),
                components.TextWithClass("/month", "text-secondary"),
            ),
            components.TextWithClass(plan.desc, "text-secondary mt-2 mb-6"),
            components.Button(components.ButtonProps{
                Text:      "Get Started",
                Variant:   components.ButtonPrimary,
                ClassName: "w-full",
                OnClick:   func() { /* checkout */ },
            }),
            components.Div("mt-6 space-y-3", featureItems...),
        )
    }

    return components.Div("py-20 px-4 surface-overlay",
        components.Div("max-w-5xl mx-auto",
            components.HeadingWithClass(2, "Simple, Transparent Pricing",
                "text-3xl font-bold text-primary text-center mb-4"),
            components.TextWithClass("No hidden fees. Cancel anytime.",
                "text-secondary text-center mb-12"),
            components.Div("grid grid-cols-1 md:grid-cols-3 gap-8", cards...),
        ),
    )
}

func featureComparison() js.Value {
    table := components.NewTable(components.TableProps{
        Columns: []components.TableColumn{
            {Header: "Feature", Key: "feature"},
            {Header: "Starter", Key: "starter"},
            {Header: "Pro", Key: "pro"},
            {Header: "Enterprise", Key: "enterprise"},
        },
        Data: []map[string]any{
            {"feature": "Projects", "starter": "5", "pro": "Unlimited", "enterprise": "Unlimited"},
            {"feature": "Storage", "starter": "1GB", "pro": "10GB", "enterprise": "100GB"},
            {"feature": "Team Members", "starter": "1", "pro": "5", "enterprise": "Unlimited"},
            {"feature": "API Access", "starter": "—", "pro": "✓", "enterprise": "✓"},
            {"feature": "Custom Domains", "starter": "—", "pro": "✓", "enterprise": "✓"},
            {"feature": "SSO", "starter": "—", "pro": "—", "enterprise": "✓"},
        },
        Striped: true,
    })

    return components.Div("py-20 px-4",
        components.Div("max-w-4xl mx-auto",
            components.HeadingWithClass(2, "Feature Comparison",
                "text-3xl font-bold text-primary text-center mb-12"),
            table.Element(),
        ),
    )
}

func faq() js.Value {
    accordion := components.NewAccordion(components.AccordionProps{
        Items: []components.AccordionItem{
            {Title: "How does the free trial work?",
                Content: components.Text("Start your 14-day free trial with full access to all Pro features. No credit card required.")},
            {Title: "Can I change plans later?",
                Content: components.Text("Yes, you can upgrade or downgrade your plan at any time. Changes take effect immediately.")},
            {Title: "What payment methods do you accept?",
                Content: components.Text("We accept all major credit cards, PayPal, and wire transfers for Enterprise plans.")},
            {Title: "Is there a refund policy?",
                Content: components.Text("Yes, we offer a 30-day money-back guarantee for all paid plans.")},
        },
        AllowMultiple: false,
    })

    return components.Div("py-20 px-4 surface-overlay",
        components.Div("max-w-2xl mx-auto",
            components.HeadingWithClass(2, "Frequently Asked Questions",
                "text-3xl font-bold text-primary text-center mb-12"),
            accordion.Element(),
        ),
    )
}
```

## Contact Form

A complete contact form with validation, file upload, and submission handling.

```go
//go:build js && wasm

package main

import (
    "syscall/js"
    "github.com/dougbarrett/gux/components"
)

func main() {
    app := components.NewApp("app")
    components.InitToasts()

    page := components.Div("min-h-screen surface-base py-12 px-4",
        components.Div("max-w-2xl mx-auto",
            // Header
            components.Div("text-center mb-8",
                components.HeadingWithClass(1, "Get in Touch",
                    "text-3xl font-bold text-primary mb-4"),
                components.TextWithClass(
                    "Have a question or want to work together? Fill out the form below and we'll get back to you within 24 hours.",
                    "text-secondary"),
            ),
            // Contact form card
            components.Card(contactForm()),
            // Alternative contact methods
            contactInfo(),
        ),
    )

    app.Mount(page)
    app.Run()
}

func contactForm() js.Value {
    // Create form fields
    nameInput := components.TextInput("Full Name", "John Doe")
    emailInput := components.EmailInput("Email Address", "john@example.com")

    phoneInput := components.NewInput(components.InputProps{
        Label:       "Phone Number",
        Type:        components.InputText,
        Placeholder: "+1 (555) 123-4567",
    })

    subjectSelect := components.SimpleSelectWithPlaceholder(
        "Subject",
        "Select a topic...",
        "General Inquiry",
        "Technical Support",
        "Sales",
        "Partnership",
        "Other",
    )

    messageArea := components.NewTextArea(components.TextAreaProps{
        Label:       "Message",
        Placeholder: "Tell us how we can help...",
        Rows:        5,
    })

    // Priority selector with radio-like badges
    var selectedPriority string = "normal"
    priorityBadges := components.Div("space-y-2",
        components.TextWithClass("Priority", "text-sm font-medium text-primary"),
        components.Div("flex gap-2",
            priorityOption("Low", "low", &selectedPriority),
            priorityOption("Normal", "normal", &selectedPriority),
            priorityOption("High", "high", &selectedPriority),
        ),
    )

    // File attachment
    fileUpload := components.FileUpload(components.FileUploadProps{
        Label:    "Attachments (optional)",
        Multiple: true,
        OnChange: func(files []components.FileInfo) {
            if len(files) > 0 {
                components.Toast("Files attached: "+string(rune('0'+len(files))), components.ToastInfo)
            }
        },
    })

    // Consent checkbox
    consentCheckbox := components.NewCheckbox(components.CheckboxProps{
        Label: "I agree to the privacy policy and terms of service",
    })

    // Newsletter checkbox
    newsletterCheckbox := components.NewCheckbox(components.CheckboxProps{
        Label:   "Subscribe to our newsletter for updates and tips",
        Checked: true,
    })

    // Submit handler
    submitForm := func() {
        // Validate required fields
        if nameInput.Value() == "" {
            components.Toast("Please enter your name", components.ToastError)
            nameInput.Focus()
            return
        }
        if emailInput.Value() == "" {
            components.Toast("Please enter your email", components.ToastError)
            emailInput.Focus()
            return
        }
        if messageArea.Value() == "" {
            components.Toast("Please enter a message", components.ToastError)
            messageArea.Focus()
            return
        }
        if !consentCheckbox.IsChecked() {
            components.Toast("Please agree to the terms", components.ToastWarning)
            return
        }

        // Show loading state
        components.Toast("Sending message...", components.ToastInfo)

        // In a real app, you would send this to your API
        // go func() {
        //     err := api.SendContactForm(...)
        //     if err != nil { ... }
        // }()

        // Simulate success
        js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
            components.Toast("Message sent successfully! We'll be in touch soon.", components.ToastSuccess)
            // Clear form
            nameInput.Clear()
            emailInput.Clear()
            phoneInput.Clear()
            messageArea.Clear()
            return nil
        }), 1000)
    }

    return components.Div("p-6 space-y-6",
        // Two-column row for name and email
        components.Div("grid grid-cols-1 md:grid-cols-2 gap-4",
            nameInput.Element(),
            emailInput.Element(),
        ),
        // Two-column row for phone and subject
        components.Div("grid grid-cols-1 md:grid-cols-2 gap-4",
            phoneInput.Element(),
            subjectSelect.Element(),
        ),
        priorityBadges,
        messageArea.Element(),
        fileUpload.Element(),
        components.Div("space-y-3",
            consentCheckbox.Element(),
            newsletterCheckbox.Element(),
        ),
        // Submit button
        components.Div("pt-4",
            components.Button(components.ButtonProps{
                Text:      "Send Message",
                Variant:   components.ButtonPrimary,
                Size:      components.ButtonLG,
                ClassName: "w-full",
                OnClick:   submitForm,
            }),
        ),
    )
}

func priorityOption(label, value string, selected *string) js.Value {
    variant := components.BadgeSecondary
    if *selected == value {
        variant = components.BadgePrimary
    }

    return components.Div("cursor-pointer",
        components.Badge(components.BadgeProps{
            Text:    label,
            Variant: variant,
        }),
    )
}

func contactInfo() js.Value {
    infoItems := []struct {
        icon, label, value string
    }{
        {"📧", "Email", "hello@example.com"},
        {"📞", "Phone", "+1 (555) 123-4567"},
        {"📍", "Address", "123 Main St, San Francisco, CA"},
    }

    items := make([]js.Value, len(infoItems))
    for i, info := range infoItems {
        items[i] = components.Div("flex items-center gap-3",
            components.TextWithClass(info.icon, "text-2xl"),
            components.Div("",
                components.TextWithClass(info.label, "text-sm text-secondary"),
                components.TextWithClass(info.value, "text-primary font-medium"),
            ),
        )
    }

    return components.Div("mt-8 grid grid-cols-1 md:grid-cols-3 gap-6", items...)
}
```

### Using FormBuilder for Simpler Forms

For basic contact forms, you can use the `FormBuilder` component for automatic validation:

```go
func simpleContactForm() js.Value {
    form := components.NewFormBuilder(components.FormBuilderProps{
        Fields: []components.BuilderField{
            {
                Name:        "name",
                Type:        components.BuilderFieldText,
                Label:       "Full Name",
                Placeholder: "John Doe",
                Rules:       []components.ValidationRule{components.Required},
            },
            {
                Name:        "email",
                Type:        components.BuilderFieldEmail,
                Label:       "Email",
                Placeholder: "john@example.com",
                Rules:       []components.ValidationRule{components.Required, components.Email},
            },
            {
                Name:        "subject",
                Type:        components.BuilderFieldSelect,
                Label:       "Subject",
                Placeholder: "Select a topic",
                Options: []components.SelectOption{
                    {Label: "General Inquiry", Value: "general"},
                    {Label: "Support", Value: "support"},
                    {Label: "Sales", Value: "sales"},
                },
            },
            {
                Name:        "message",
                Type:        components.BuilderFieldTextarea,
                Label:       "Message",
                Placeholder: "How can we help?",
                Rows:        5,
                Rules:       []components.ValidationRule{components.Required, components.MinLength(20)},
            },
            {
                Name:  "newsletter",
                Type:  components.BuilderFieldCheckbox,
                Label: "Subscribe to newsletter",
            },
        },
        SubmitText: "Send Message",
        ShowCancel: true,
        CancelText: "Clear",
        OnSubmit: func(values map[string]any) error {
            // Send to API
            components.Toast("Message sent!", components.ToastSuccess)
            return nil
        },
        OnCancel: func() {
            components.Toast("Form cleared", components.ToastInfo)
        },
    })

    return form.Element()
}
```

## Next Steps

- [Components](components.md) — Full component API reference
- [State Management](state-management.md) — Manage application state
- [Getting Started](getting-started.md) — Project setup guide
