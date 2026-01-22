# v2.0 Core Components — Requirements

## Overview

Build a fresh component library on core's Node system with four example applications demonstrating real-world use cases.

## Functional Requirements

### Component Library

**Layout Components (P0):**
- [ ] Card with header/content/footer composition
- [ ] Container with max-width constraints
- [ ] VStack/HStack for flex layouts
- [ ] Grid with responsive columns
- [ ] Divider (horizontal/vertical)

**Form Components (P0):**
- [ ] Input with label, placeholder, error states
- [ ] Textarea with resize control
- [ ] Select with options and placeholder
- [ ] Checkbox with label
- [ ] Radio group
- [ ] Switch toggle
- [ ] Form wrapper with validation display

**Data Display Components (P0):**
- [ ] Table with compound pattern (Table/Thead/Tbody/Tr/Th/Td)
- [ ] DataTable[T] with generic typing
- [ ] Badge with variants (default, success, warning, error)
- [ ] Avatar with fallback
- [ ] List with items

**Feedback Components (P0):**
- [ ] Alert with variants (info, success, warning, error)
- [ ] Toast notification system
- [ ] Spinner/loading indicator
- [ ] Progress bar
- [ ] Skeleton loader

**Navigation Components (P0):**
- [ ] Breadcrumb with separator
- [ ] Pagination with page numbers
- [ ] Tabs with panels
- [ ] Menu/nav links

**Overlay Components (P0):**
- [ ] Modal/Dialog with open/close
- [ ] Dropdown menu
- [ ] Tooltip on hover

**Action Components (P0):**
- [ ] Button with variants (primary, secondary, outline, ghost, destructive)
- [ ] Button with sizes (sm, md, lg)
- [ ] ButtonGroup
- [ ] IconButton

### Marketing Example

**Pages (P0):**
- [ ] Home page with hero, feature grid, testimonials, CTA
- [ ] Features page with detailed feature sections
- [ ] Pricing page with pricing cards
- [ ] About page with team/story
- [ ] Contact page with form

**Patterns (P0):**
- [ ] Responsive navigation with mobile menu
- [ ] Footer with links
- [ ] SEO-optimized SSR

### SaaS Dashboard Example

**Pages (P0):**
- [ ] Dashboard overview with stats cards and charts
- [ ] Resource list with DataTable
- [ ] Resource detail/edit form
- [ ] Settings page
- [ ] Profile page

**Patterns (P0):**
- [ ] Authenticated routes
- [ ] CRUD operations via generated API
- [ ] Dashboard layout with sidebar

### Admin Panel Example

**Pages (P0):**
- [ ] Dashboard with activity metrics
- [ ] User list with search/filter
- [ ] User detail/edit
- [ ] Activity logs
- [ ] Settings page

**Patterns (P0):**
- [ ] Admin layout with navigation
- [ ] Role-based UI (display only, defer actual RBAC)
- [ ] Bulk actions on lists

### Auth Example

**Pages (P0):**
- [ ] Login with email/password
- [ ] Registration with validation
- [ ] Forgot password request
- [ ] Password reset with token
- [ ] Email verification landing

**Patterns (P0):**
- [ ] Form validation feedback
- [ ] Session management patterns
- [ ] Redirect after auth

## Non-Functional Requirements

**Rendering (P0):**
- All components must render correctly in SSR and WASM
- No hydration mismatches on initial render
- Components use core.Node interface, not direct js.Value

**Styling (P0):**
- Tailwind CSS utility classes only
- No additional CSS frameworks
- Consistent design tokens (colors, spacing, typography)

**Performance (P1):**
- WASM bundles < 5MB per route bundle
- Event handlers properly cleaned up to prevent memory leaks
- No blocking operations in render path

**Developer Experience (P1):**
- Props structs for all components (no interface{})
- Clear prop naming conventions
- Components work with hot reload

## Constraints

- **No backwards compatibility** — fresh implementation on core.Node
- **Tailwind CSS only** — no DaisyUI or other Tailwind plugins
- **Core auth flows only** — no OAuth, 2FA, magic links
- **Focused examples** — 5-10 pages each, not complete products

## Acceptance Criteria

1. Component library provides all P0 components listed above
2. Each example app runs standalone with `gux dev`
3. All examples demonstrate SSR + WASM hydration
4. No TypeScript/JavaScript required to build examples
5. Components render identically in SSR and WASM

---
*Created: 2026-01-22*
