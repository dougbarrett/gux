# Features Research: Component Library + Example Apps

**Domain:** Go/WASM Component Library with SSR + Hydration
**Researched:** 2026-01-22
**Overall Confidence:** HIGH (verified against shadcn/ui, Radix UI, Tailwind UI patterns)

---

## Component Library

### Table Stakes

Components users expect from any modern component library. Missing these = incomplete library.

| Component | Why Expected | Complexity | Existing? | Notes |
|-----------|--------------|------------|-----------|-------|
| **Button** | Every UI needs clickable actions | Low | Yes | Needs variants (primary, secondary, outline, ghost, destructive) |
| **Input** | Text entry is fundamental | Low | Yes | Needs states (error, disabled, focus) |
| **Textarea** | Multi-line text entry | Low | Yes | Needs resize control, character count |
| **Select** | Dropdown selection | Medium | Yes | Needs searchable variant |
| **Checkbox** | Binary toggles | Low | Yes | Needs indeterminate state |
| **Radio Group** | Single selection from options | Low | Partial | Needs grouping with labels |
| **Switch/Toggle** | On/off states | Low | Yes | Needs label positioning |
| **Label** | Accessible form labels | Low | Partial | Must link to inputs properly |
| **Form** | Form container with validation | Medium | Yes | Needs server error handling |
| **Card** | Content container | Low | Yes | Needs header/content/footer structure |
| **Badge** | Status indicators | Low | Yes | Needs variants (success, warning, error, info) |
| **Alert** | User notifications | Low | Yes | Needs dismissible variant |
| **Modal/Dialog** | Overlay content | Medium | Yes | Needs focus trap, escape handling |
| **Dropdown Menu** | Action menus | Medium | Yes | Needs submenus, keyboard navigation |
| **Tabs** | Content organization | Medium | Yes | Needs controlled/uncontrolled modes |
| **Accordion** | Collapsible sections | Medium | Yes | Needs single/multiple expand modes |
| **Table** | Data display | Medium | Yes | Needs sorting, selection headers |
| **Pagination** | Data navigation | Low | Yes | Needs page size selector |
| **Tooltip** | Hover information | Low | Yes | Needs positioning, delay |
| **Toast/Sonner** | Transient notifications | Medium | Yes | Needs stack management, actions |
| **Skeleton** | Loading placeholders | Low | Yes | Needs various shapes |
| **Spinner/Progress** | Loading indicators | Low | Yes | Needs determinate/indeterminate |
| **Avatar** | User representation | Low | Yes | Needs fallback initials |
| **Breadcrumb** | Navigation hierarchy | Low | Yes | Needs separator customization |
| **Separator** | Visual dividers | Low | Partial | Horizontal and vertical |
| **Scroll Area** | Custom scrollbars | Medium | Partial | Cross-browser consistency |

### Differentiators

Components that set a library apart. Not expected but highly valued.

| Component | Value Proposition | Complexity | Existing? | Notes |
|-----------|-------------------|------------|-----------|-------|
| **Command Palette** | Power-user search/actions (Cmd+K) | High | Yes | Rare in component libraries, major DX win |
| **Data Table** | Full-featured table with sort/filter/select | High | Partial | TanStack Table patterns |
| **Date Picker** | Date/range selection | High | Yes | Calendar integration |
| **Combobox** | Searchable select with autocomplete | High | Yes | Complex keyboard handling |
| **Sidebar** | Application navigation | Medium | Yes | Collapsible, mobile responsive |
| **Sheet/Drawer** | Slide-in panels | Medium | Yes | Multiple positions |
| **Form Builder** | Declarative form generation | High | Yes | Reduces boilerplate |
| **Charts** | Data visualization | High | Yes | Multiple chart types |
| **Rich Text Editor** | WYSIWYG editing | Very High | Yes | Complex, requires external lib |
| **Virtual List** | Large dataset rendering | High | Yes | Performance critical |
| **Focus Trap** | Accessibility utility | Medium | Yes | Modal support |
| **Skip Links** | Accessibility navigation | Low | Yes | WCAG requirement |
| **Theme Provider** | Dark/light mode | Medium | Yes | System preference detection |
| **Notification Center** | Notification management | Medium | Yes | Aggregates toasts |
| **Empty State** | No-data handling | Low | Yes | Reduces blank screens |
| **Stepper** | Multi-step flows | Medium | Yes | Wizard patterns |

### Anti-Features

Features to deliberately NOT build. Common mistakes in component libraries.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **CSS-in-JS runtime** | WASM + runtime CSS = performance hit | Use Tailwind utility classes (compile-time) |
| **Overly opinionated styling** | Limits customization | Provide composable primitives with Tailwind classes |
| **Global state in components** | Breaks isolation | Components should accept props, not manage global state |
| **Complex prop drilling** | Hard to use | Use composition patterns (children, slots) |
| **Non-accessible defaults** | Excludes users | ARIA, keyboard nav, focus management built-in |
| **JavaScript-only interactions** | SSR breaks | Progressive enhancement, works without JS |
| **Tight coupling to routing** | Framework lock-in | Components agnostic to router implementation |
| **Building what Radix provides** | Wasted effort | Focus on Go/WASM-specific value, not reimplementing Dialog |
| **Animation libraries** | Bundle size, complexity | CSS transitions/animations where possible |
| **Icon library bundling** | Huge bundles | Tree-shakeable or external icon system |

---

## Marketing Site Example

### Expected Pages/Features

Marketing sites are SEO-focused, content-heavy, minimal interactivity.

| Page/Feature | Purpose | Complexity | Notes |
|--------------|---------|------------|-------|
| **Home/Landing** | First impression, value prop | Medium | Hero, features, social proof, CTA |
| **Features** | Detail product capabilities | Low | Feature grid, icons, descriptions |
| **Pricing** | Convert visitors | Medium | Pricing cards, feature comparison, FAQ |
| **About** | Build trust | Low | Team, mission, story |
| **Contact** | Capture leads | Low | Form, contact info, map optional |
| **Blog/Resources** | SEO, thought leadership | Medium | List page + detail pages |

### Page Sections (Table Stakes)

| Section | Where Used | Components Needed |
|---------|------------|-------------------|
| **Hero** | Home | Heading, Text, Button, possibly Image/Video |
| **Feature Grid** | Home, Features | Card, Icon, Heading, Text |
| **Pricing Cards** | Pricing | Card, Badge, Button, List |
| **Testimonials** | Home, About | Card, Avatar, Text, Quote |
| **CTA Banner** | Multiple | Card, Heading, Button |
| **FAQ Accordion** | Pricing, Home | Accordion |
| **Contact Form** | Contact | Form, Input, Textarea, Button |
| **Footer** | All | Links, Text, possibly Newsletter |
| **Navigation** | All | Header, Link, Dropdown (mobile menu) |

### Marketing Site Differentiators

| Feature | Value | Complexity |
|---------|-------|------------|
| **Responsive navigation** | Mobile hamburger menu with smooth animation | Medium |
| **Announcement banner** | Highlight promotions/news | Low |
| **Animated hero** | Visual appeal | Medium |
| **Social proof logos** | Trust building | Low |
| **Newsletter signup** | Lead capture | Low |

### Marketing Site Anti-Features

| Anti-Feature | Why Avoid |
|--------------|-----------|
| **Heavy JavaScript** | Kills SEO, slow load times |
| **Client-side routing** | Unnecessary for static content |
| **Complex state management** | Marketing pages are static content |
| **User authentication** | Not needed for public marketing |

---

## SaaS Dashboard Example

### Expected Pages/Features

SaaS dashboards are data-rich, interactive, authenticated.

| Page/Feature | Purpose | Complexity | Notes |
|--------------|---------|------------|-------|
| **Dashboard/Overview** | Key metrics at a glance | High | Stats cards, charts, recent activity |
| **Resource List** | View all items | Medium | Table with search, filter, pagination |
| **Resource Detail** | View single item | Medium | Detail layout, related data |
| **Resource Create/Edit** | CRUD operations | Medium | Form with validation |
| **Settings** | User preferences | Medium | Tabbed form sections |
| **Profile** | Account management | Low | Avatar, form, password change |
| **Billing** (optional) | Subscription management | Medium | Plan cards, payment info |

### Dashboard Components (Table Stakes)

| Component | Purpose | Complexity |
|-----------|---------|------------|
| **Stats Card** | KPI display (number + trend) | Low |
| **Chart** | Data visualization | High |
| **Data Table** | List resources with actions | High |
| **Search Input** | Filter data | Low |
| **Filter Controls** | Multi-criteria filtering | Medium |
| **Empty State** | No data messaging | Low |
| **Loading States** | Data fetching feedback | Low |
| **Action Menu** | Row/item actions | Medium |
| **Confirmation Dialog** | Destructive action confirmation | Medium |

### Dashboard Layout (Table Stakes)

| Element | Purpose | Complexity |
|---------|---------|------------|
| **Sidebar Navigation** | Primary navigation | Medium |
| **Header** | Breadcrumb, user menu, search | Medium |
| **Main Content Area** | Page content | Low |
| **User Menu** | Profile, settings, logout | Medium |

### Dashboard Differentiators

| Feature | Value | Complexity |
|---------|-------|------------|
| **Real-time updates** | Live data without refresh | High |
| **Bulk actions** | Multi-select with batch operations | Medium |
| **Export functionality** | CSV/PDF data export | Medium |
| **Keyboard shortcuts** | Power-user efficiency | Medium |
| **Saved views/filters** | Personalized data views | Medium |
| **Activity feed** | Recent actions timeline | Low |

### Dashboard Anti-Features

| Anti-Feature | Why Avoid |
|--------------|-----------|
| **Over-fetching data** | Performance, bandwidth |
| **Polling instead of events** | Server load, battery drain |
| **Complex nested navigation** | Confusing UX |
| **Unauthenticated access** | Security risk |

---

## Admin Panel Example

### Expected Pages/Features

Admin panels manage the system itself, not just user data.

| Page/Feature | Purpose | Complexity | Notes |
|--------------|---------|------------|-------|
| **Dashboard** | System health overview | Medium | Stats, activity, alerts |
| **User Management** | CRUD users | High | List, detail, edit, roles |
| **User Detail** | View/edit single user | Medium | Profile, activity, permissions |
| **Content Management** | CRUD any content type | High | Generic CRUD patterns |
| **Settings** | System configuration | Medium | Multiple setting categories |
| **Activity Logs** | Audit trail | Medium | Searchable, filterable log |
| **Roles/Permissions** | Access control | High | RBAC management |

### Admin-Specific Components

| Component | Purpose | Complexity |
|-----------|---------|------------|
| **User List Table** | Display users with roles, status | Medium |
| **User Detail Card** | Show user info + actions | Medium |
| **Role Badge** | Display user role visually | Low |
| **Status Indicator** | Active/inactive/banned | Low |
| **Activity Timeline** | Chronological event display | Medium |
| **Log Viewer** | Searchable log display | Medium |
| **Settings Form** | Key-value configuration | Medium |
| **Permission Matrix** | Role-permission grid | High |
| **Impersonation Warning** | Alert when impersonating | Low |

### Admin Differentiators

| Feature | Value | Complexity |
|---------|-------|------------|
| **User impersonation** | Debug user issues | High |
| **Bulk user operations** | Mass update/delete/export | Medium |
| **Advanced filtering** | Multi-field, saved filters | Medium |
| **Audit log export** | Compliance | Medium |
| **System health monitoring** | Uptime, errors, performance | High |

### Admin Anti-Features

| Anti-Feature | Why Avoid |
|--------------|-----------|
| **Over-complex RBAC UI** | Hard to use, error-prone |
| **No audit logging** | Compliance risk |
| **Permanent delete without confirmation** | Data loss |
| **All-or-nothing access** | Security risk |

---

## Auth Flows Example

### Expected Pages/Features

Authentication flows must be secure, accessible, and forgiving.

| Page/Feature | Purpose | Complexity | Notes |
|--------------|---------|------------|-------|
| **Login** | Primary authentication | Medium | Email/password, remember me |
| **Register** | New account creation | Medium | Form validation, terms acceptance |
| **Forgot Password** | Password recovery initiation | Low | Email input only |
| **Reset Password** | Set new password | Medium | Token validation, password rules |
| **Email Verification** | Confirm email ownership | Low | Token validation |
| **2FA Setup** (optional) | Add second factor | High | QR code, backup codes |
| **2FA Verify** (optional) | Second factor entry | Medium | TOTP input |

### Auth Form Components

| Component | Purpose | Complexity |
|-----------|---------|------------|
| **Email Input** | Email validation | Low |
| **Password Input** | Show/hide toggle, strength meter | Medium |
| **Password Strength Meter** | Visual feedback | Low |
| **Remember Me Checkbox** | Session persistence | Low |
| **Terms Checkbox** | Legal acceptance | Low |
| **OTP Input** | 6-digit code entry | Medium |
| **Social Login Buttons** | OAuth providers | Low (styling only) |

### Auth Page Layouts

| Layout | Purpose | Components |
|--------|---------|------------|
| **Centered Card** | Focused single action | Card, Logo, Form |
| **Split Screen** | Branding + form | Image/branding left, form right |
| **Full Page Form** | Mobile-first | Form takes full width |

### Auth UX Patterns (Table Stakes)

| Pattern | Why Expected |
|---------|--------------|
| **Clear error messages** | Users need to know what went wrong |
| **Password visibility toggle** | Reduce typos |
| **Password requirements shown** | Reduce frustration |
| **Real-time validation** | Immediate feedback |
| **Link to alternative** | "Don't have account? Register" |
| **Loading states** | Feedback during submission |
| **Success confirmation** | Clear completion signal |

### Auth Differentiators

| Feature | Value | Complexity |
|---------|-------|------------|
| **Magic link login** | Passwordless option | Medium |
| **Social login integration** | Convenience | Medium |
| **Passkey support** | Modern auth | High |
| **Login anomaly detection** | Security notification | High |
| **Session management** | View/revoke sessions | Medium |

### Auth Anti-Features

| Anti-Feature | Why Avoid |
|--------------|-----------|
| **Blocking password managers** | Frustrates users, reduces security |
| **Security questions** | Outdated, insecure |
| **Complex CAPTCHA on every attempt** | Friction |
| **Vague error messages** | "Invalid credentials" without hint |
| **Forced password expiration** | NIST recommends against |
| **Short token expiration** | Requiring re-login constantly |

---

## Shared Foundation

### Common Patterns Across Examples

These patterns appear in all four example types and should be standardized.

| Pattern | Description | Components Involved |
|---------|-------------|---------------------|
| **App Shell** | Consistent layout with navigation | Layout, Sidebar, Header |
| **Page Header** | Title + actions pattern | Heading, Button, Breadcrumb |
| **Form Handling** | Validation + submission + errors | Form, Input, Button, Alert |
| **Data Fetching** | Loading, error, success states | Skeleton, Spinner, Alert, Empty State |
| **Confirmation Flow** | Destructive action protection | Dialog, Button |
| **Toast Notifications** | Action feedback | Toast |
| **Responsive Navigation** | Mobile-friendly navigation | Sidebar, Header, Drawer |
| **Theme Support** | Light/dark mode | Theme provider |

### Shared Components (Used Across Examples)

| Component | Marketing | SaaS | Admin | Auth |
|-----------|-----------|------|-------|------|
| Button | Y | Y | Y | Y |
| Input | Y | Y | Y | Y |
| Form | Y | Y | Y | Y |
| Card | Y | Y | Y | Y |
| Alert | N | Y | Y | Y |
| Toast | N | Y | Y | Y |
| Modal | N | Y | Y | N |
| Sidebar | N | Y | Y | N |
| Header | Y | Y | Y | N |
| Table | N | Y | Y | N |
| Pagination | N | Y | Y | N |
| Dropdown | Y | Y | Y | N |
| Avatar | N | Y | Y | Y |
| Badge | N | Y | Y | N |
| Breadcrumb | N | Y | Y | N |
| Skeleton | N | Y | Y | N |
| Empty State | N | Y | Y | N |

### Styling Foundation

| Element | Approach | Notes |
|---------|----------|-------|
| **Colors** | Tailwind CSS color system | Primary, secondary, destructive, muted |
| **Typography** | Tailwind typography | Heading scale, body text, mono |
| **Spacing** | Tailwind spacing scale | Consistent padding/margin |
| **Shadows** | Tailwind shadows | Elevation levels |
| **Border Radius** | Tailwind border-radius | Consistent rounding |
| **Dark Mode** | Tailwind dark: prefix | System preference + manual toggle |

### Accessibility Foundation

| Requirement | Implementation | Priority |
|-------------|----------------|----------|
| **Keyboard Navigation** | All interactive elements focusable | Critical |
| **Focus Indicators** | Visible focus rings | Critical |
| **ARIA Labels** | Screen reader support | Critical |
| **Color Contrast** | WCAG AA minimum | Critical |
| **Focus Trapping** | Modals trap focus | High |
| **Skip Links** | Bypass navigation | High |
| **Error Announcements** | Live regions for errors | High |
| **Reduced Motion** | Respect prefers-reduced-motion | Medium |

---

## Feature Dependencies

```
Core Framework (existing)
    |
    +-- Element Helpers (Div, Button, Input, etc.)
    |
    +-- State Management (UseState, StateInt, etc.)
    |
    +-- Routing + Bundles
    |
    +-- CRUD Generation
    |
    +-- CSRF Protection
    |
Component Library (build on core)
    |
    +-- Primitives (Button, Input, Card, etc.)
    |       |
    |       +-- Layout Components (Sidebar, Header, Layout)
    |       |
    |       +-- Feedback Components (Toast, Alert, Modal)
    |       |
    |       +-- Data Components (Table, Pagination, Charts)
    |
    +-- Composite Components (Form, DataTable, CommandPalette)
    |
Example Apps (consume components)
    |
    +-- Marketing Site (SSR-focused, minimal JS)
    |
    +-- SaaS Dashboard (WASM + data fetching)
    |
    +-- Admin Panel (WASM + CRUD + RBAC patterns)
    |
    +-- Auth Flows (shared auth patterns)
```

---

## MVP Recommendation

### Phase 1: Shared Foundation + Auth
- Layout primitives (sidebar, header)
- Form components with validation
- Auth flows (login, register, password reset)
- **Rationale:** Auth is needed by SaaS and Admin examples

### Phase 2: Marketing Site
- Marketing-specific sections (hero, features, pricing)
- Responsive navigation
- Static content patterns
- **Rationale:** Simplest example, demonstrates SSR value

### Phase 3: SaaS Dashboard
- Dashboard layout
- Data table with CRUD
- Stats cards and charts
- **Rationale:** Core use case for framework

### Phase 4: Admin Panel
- User management CRUD
- Activity logs
- Settings patterns
- **Rationale:** Builds on SaaS patterns, adds admin-specific needs

---

## Sources

### Component Library Patterns
- [shadcn/ui Components](https://ui.shadcn.com/docs/components) - Component inventory and API patterns
- [Radix UI Primitives](https://www.radix-ui.com/primitives/docs/overview/introduction) - Accessibility patterns
- [Radix UI Accessibility](https://www.radix-ui.com/primitives/docs/overview/accessibility) - WAI-ARIA implementation

### SaaS/Dashboard Patterns
- [TailAdmin SaaS Dashboard Templates](https://tailadmin.com/blog/saas-dashboard-templates) - Dashboard feature expectations
- [SaaS Dashboard Examples - Userpilot](https://userpilot.com/blog/saas-dashboard-examples/) - Metrics and KPIs
- [SaaS Dashboards - NetSuite](https://www.netsuite.com/portal/resource/articles/erp/saas-dashboards.shtml) - Dashboard types

### Admin Panel Patterns
- [Admin Dashboard Guide - WeWeb](https://www.weweb.io/blog/admin-dashboard-ultimate-guide-templates-examples) - Admin features
- [AdminLTE Best Practices](https://adminlte.io/blog/admin-templates/) - Template patterns

### Auth Flow Patterns
- [Login & Signup UX Guide - Authgear](https://www.authgear.com/post/login-signup-ux-guide) - Auth UX patterns
- [Password Reset Best Practices - Authgear](https://www.authgear.com/post/authentication-security-password-reset-best-practices-and-more) - Security considerations
- [Authentication UX Design - Smashing Magazine](https://www.smashingmagazine.com/2022/08/authentication-ux-design-guidelines/) - UX guidelines

### Marketing Site Patterns
- [Hero Section Best Practices - Prismic](https://prismic.io/blog/website-hero-section) - Hero section design
- [B2B Homepage Examples - BlendB2B](https://www.blendb2b.com/blog/the-10-best-b2b-homepage-examples) - Marketing page structure

### Anti-Patterns
- [Stop Reinventing Component Libraries - Medium](https://medium.com/@rmehlinger/stop-reinventing-component-libraries-part-i-b191861982cc) - Build vs buy
- [Table Stakes Features - UpTech Studio](https://www.uptechstudio.com/blog/the-hidden-complexity-of-table-stakes-features) - Feature complexity
