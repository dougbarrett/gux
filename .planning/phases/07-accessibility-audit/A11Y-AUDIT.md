# GoQuery Accessibility Audit Report

**Audit Date:** 2026-01-15
**Standard:** WCAG 2.1 Level AA
**Components Audited:** 25
**Auditor:** Claude Code

---

## Executive Summary

| Priority | Count | Description |
|----------|-------|-------------|
| **P0 - Critical** | 8 | Blockers preventing access entirely |
| **P1 - High** | 52 | Major barriers impeding usage |
| **P2 - Medium** | 37 | Minor barriers with workarounds |
| **P3 - Low** | 17 | Enhancements beyond minimum compliance |
| **Total** | 114 | |

**Key Findings:**
- **7 components** lack proper label-input association (1.3.1)
- **5 components** missing complete ARIA widget patterns (4.1.2)
- **6 components** need keyboard navigation improvements (2.1.1)
- **4 feedback components** missing live regions (4.1.3)
- **2 components** nearly compliant (Toggle, Breadcrumbs)
- **1 component** fully accessible (Link)

---

## WCAG 2.1 AA Compliance Matrix

### Principle 1: Perceivable

| Criterion | Status | Components Affected | Phase |
|-----------|--------|---------------------|-------|
| 1.1.1 Non-text Content | Partial | Alert, Toast, Spinner | 8 |
| 1.3.1 Info and Relationships | **Missing** | Input, Select, Checkbox, TextArea, DatePicker, FormBuilder, Table | 8 |
| 1.3.5 Identify Input Purpose | Missing | Input (no autocomplete) | 8 |

### Principle 2: Operable

| Criterion | Status | Components Affected | Phase |
|-----------|--------|---------------------|-------|
| 2.1.1 Keyboard | Partial | Tabs, DatePicker, Modal | 9 |
| 2.1.2 No Keyboard Trap | Partial | Modal (no focus trap) | 9 |
| 2.4.1 Bypass Blocks | Missing | Sidebar (no skip link) | 9 |
| 2.4.3 Focus Order | Partial | Modal, Dropdown, Sidebar | 9 |
| 2.4.7 Focus Visible | Partial | Dropdown, Toggle | 10 |
| 2.4.8 Location | Missing | Sidebar (no aria-current) | 8 |

### Principle 3: Understandable

| Criterion | Status | Components Affected | Phase |
|-----------|--------|---------------------|-------|
| 3.3.1 Error Identification | **Missing** | Input, TextArea, Select, Form, FormBuilder | 8 |
| 3.3.2 Labels or Instructions | Partial | Input, Select, TextArea, FormBuilder | 8 |

### Principle 4: Robust

| Criterion | Status | Components Affected | Phase |
|-----------|--------|---------------------|-------|
| 4.1.2 Name, Role, Value | **Missing** | Modal, CommandPalette, Combobox, Tabs, DatePicker, Progress, Spinner | 8 |
| 4.1.3 Status Messages | Missing | Alert, Toast, Progress, Spinner, Skeleton, Table, Pagination | 8 |

---

## By Component Category

### Interactive Components (8 components)

| Component | High | Medium | Low | Most Critical Gap |
|-----------|------|--------|-----|-------------------|
| Modal | 4 | 2 | 1 | Missing role="dialog", focus trap |
| Dropdown | 2 | 2 | 1 | Trigger lacks aria-expanded |
| CommandPalette | 5 | 3 | 1 | Missing dialog/listbox roles |
| ConfirmDialog | 1 | 2 | 1 | Should use role="alertdialog" |
| Combobox | 6 | 3 | 1 | Missing complete ARIA combobox pattern |
| Tabs | 7 | 2 | 0 | Missing all ARIA tab roles + keyboard |
| Accordion | 1 | 3 | 1 | Missing aria-expanded |
| Table | 2 | 5 | 2 | Missing scope="col" and aria-sort |

**Subtotal:** 28 High, 22 Medium, 8 Low = **58 gaps**

### Form Control Components (8 components)

| Component | High | Medium | Low | Most Critical Gap |
|-----------|------|--------|-----|-------------------|
| Input | 3 | 2 | 0 | Label not associated, no aria-invalid |
| Select | 1 | 2 | 0 | Label not associated |
| Checkbox | 1 | 1 | 1 | Label not associated via htmlFor |
| Toggle | 0 | 1 | 2 | Minor - needs aria-labelledby |
| TextArea | 1 | 2 | 0 | Label not associated |
| DatePicker | 6 | 2 | 1 | Missing complete combobox/grid pattern |
| Form | 2 | 1 | 1 | Error messages not linked |
| FormBuilder | 2 | 2 | 1 | Missing aria-describedby for errors |

**Subtotal:** 16 High, 13 Medium, 6 Low = **35 gaps**

### Navigation Components (4 components)

| Component | High | Medium | Low | Most Critical Gap |
|-----------|------|--------|-----|-------------------|
| Sidebar | 1 | 2 | 2 | Missing aria-current="page" |
| Breadcrumbs | 0 | 1 | 0 | Separator should be aria-hidden |
| Pagination | 0 | 0 | 2 | Page buttons could use "Page X" labels |
| Link | 0 | 0 | 0 | **Fully accessible** |

**Subtotal:** 1 High, 3 Medium, 4 Low = **8 gaps**

### Feedback Components (5 components)

| Component | High | Medium | Low | Most Critical Gap |
|-----------|------|--------|-----|-------------------|
| Alert | 2 | 1 | 0 | Missing role="alert" |
| Toast | 3 | 2 | 0 | No ARIA live region |
| Progress | 3 | 2 | 0 | Missing role="progressbar" |
| Spinner | 3 | 1 | 0 | Missing role="status" |
| Skeleton | 0 | 2 | 1 | Needs aria-hidden + parent aria-busy |

**Subtotal:** 11 High, 8 Medium, 1 Low = **20 gaps** (includes P0 critical)

---

## Remediation Priorities

### P0 - Critical (Blockers)
Issues that prevent access entirely for users with disabilities.

| # | Component | Gap | WCAG | Phase |
|---|-----------|-----|------|-------|
| 1 | Modal | No focus trap - focus escapes dialog | 2.1.2 | 9 |
| 2 | Modal | No focus restoration on close | 2.4.3 | 9 |
| 3 | Tabs | No arrow key navigation | 2.1.1 | 9 |
| 4 | DatePicker | No keyboard navigation in calendar | 2.1.1 | 9 |
| 5 | Alert/Toast | Messages not announced to screen readers | 4.1.3 | 8 |
| 6 | Progress/Spinner | Loading states invisible to AT | 4.1.2, 4.1.3 | 8 |
| 7 | Input/Select/TextArea | Labels not programmatically associated | 1.3.1 | 8 |
| 8 | Form errors | Error messages not linked to inputs | 3.3.1 | 8 |

### P1 - High (Major Barriers)
Significant barriers that impede usage but workarounds may exist.

| Category | Count | Components | Phase |
|----------|-------|------------|-------|
| Missing ARIA roles | 18 | Modal, CommandPalette, Combobox, Tabs, DatePicker, Progress | 8 |
| Missing aria-expanded/selected | 10 | Dropdown, Combobox, Tabs, Accordion, DatePicker | 8 |
| Missing aria-invalid for errors | 8 | Input, TextArea, Select, Form, FormBuilder | 8 |
| Missing live regions | 8 | Alert, Toast, Progress, Spinner | 8 |
| Missing label associations | 8 | Input, Select, Checkbox, TextArea, DatePicker | 8 |

### P2 - Medium (Minor Barriers)
Issues that cause difficulty but workarounds exist.

| Category | Count | Components | Phase |
|----------|-------|------------|-------|
| Missing aria-controls links | 9 | Modal, Dropdown, CommandPalette, Combobox, Tabs, Accordion | 8 |
| Missing aria-describedby | 6 | Toggle, ConfirmDialog, Form, FormBuilder, Checkbox | 8 |
| Icon accessibility | 5 | Alert, Toast, Table | 8 |
| Focus indicator improvements | 4 | Toggle, Dropdown | 10 |
| Navigation labels | 3 | Sidebar, Breadcrumbs | 8 |

### P3 - Low (Enhancements)
Best practices and enhancements beyond minimum compliance.

| Category | Count | Components | Phase |
|----------|-------|------------|-------|
| Live region for status changes | 5 | CommandPalette, Combobox, Table, Pagination | 8 |
| Skip link support | 2 | Sidebar | 9 |
| Keyboard shortcuts documentation | 2 | Sidebar, CommandPalette | 10 |
| Enhanced ARIA descriptions | 8 | Various | 8 |

---

## Phase Mapping

### Phase 8: ARIA & Semantic Markup
**Focus:** Screen reader support, labels, roles, live regions

**Components to remediate:**
1. **Modal** - Add role="dialog", aria-modal, aria-labelledby
2. **CommandPalette** - Add role="dialog", listbox/option pattern
3. **ConfirmDialog** - Override to role="alertdialog", aria-describedby
4. **Combobox** - Complete ARIA combobox pattern
5. **Tabs** - Complete ARIA tablist/tab/tabpanel pattern
6. **Accordion** - Add aria-expanded, aria-controls
7. **Dropdown** - Add trigger ARIA (aria-expanded, aria-haspopup)
8. **Table** - Add scope="col", aria-sort
9. **Form Controls** (Input, Select, Checkbox, TextArea) - Label associations, aria-invalid
10. **DatePicker** - ARIA combobox/dialog pattern
11. **Alert/Toast** - role="alert"/role="status"
12. **Progress/Spinner** - role="progressbar"/role="status"
13. **Sidebar** - aria-current="page"
14. **Breadcrumbs** - aria-hidden on separators
15. **Form/FormBuilder** - aria-describedby for errors

**Estimated issues addressed:** 60 (High + Medium ARIA issues)

### Phase 9: Keyboard Navigation
**Focus:** Comprehensive keyboard support for all components

**Components to remediate:**
1. **Modal** - Focus trap implementation, focus restoration
2. **Tabs** - Arrow key navigation, Home/End, tabindex management
3. **DatePicker** - Arrow key calendar navigation
4. **Sidebar** - Focus management on mobile open
5. **Dropdown** - Focus restoration to trigger on close
6. **Skip link** - Add skip navigation support

**Estimated issues addressed:** 15 (Keyboard-related gaps)

### Phase 10: Visual Accessibility
**Focus:** Focus indicators, contrast, reduced motion

**Components to remediate:**
1. **Toggle** - Visible focus indicator
2. **Dropdown** - Focus visible verification
3. **All components** - Focus-visible styling audit
4. **Animations** - prefers-reduced-motion support
5. **Color contrast** - Audit against WCAG AA ratios

**Estimated issues addressed:** 10 (Visual gaps)

### Phase 11: Testing Infrastructure
**Focus:** axe-core integration, testing patterns

**Testing patterns to implement:**
1. axe-core integration for automated scanning
2. Unit tests for ARIA attribute presence
3. Integration tests for keyboard navigation
4. Screen reader testing guidelines
5. Accessibility regression prevention

**Estimated issues addressed:** Verification + prevention

---

## Baseline Metrics

| Metric | Current | Target (Post-Phase 11) |
|--------|---------|------------------------|
| Components with keyboard support | 12/25 (48%) | 25/25 (100%) |
| Components with ARIA labels | 8/25 (32%) | 25/25 (100%) |
| Components with visible focus | 20/25 (80%) | 25/25 (100%) |
| Components with proper roles | 3/25 (12%) | 25/25 (100%) |
| Form controls with label association | 2/8 (25%) | 8/8 (100%) |
| Feedback components with live regions | 0/5 (0%) | 5/5 (100%) |
| **Overall WCAG 2.1 AA compliance** | ~40% | 100% |

### Current State by Category

| Category | Compliant | Partial | Non-compliant |
|----------|-----------|---------|---------------|
| Interactive (8) | 0 | 4 | 4 |
| Form Controls (8) | 1 (Toggle) | 1 (FormBuilder) | 6 |
| Navigation (4) | 1 (Link) | 3 | 0 |
| Feedback (5) | 0 | 0 | 5 |

### Components Requiring Most Work

1. **DatePicker** - 9 gaps, needs complete ARIA pattern + keyboard
2. **Tabs** - 9 gaps, needs all ARIA roles + keyboard navigation
3. **Combobox** - 10 gaps, needs complete ARIA combobox pattern
4. **CommandPalette** - 9 gaps, needs dialog + listbox pattern
5. **Modal** - 7 gaps, needs role + focus management

### Components Closest to Compliance

1. **Link** - 0 gaps, fully accessible
2. **Toggle** - 3 minor gaps, has role="switch" + aria-checked
3. **Breadcrumbs** - 1 gap, just needs separator fix
4. **Pagination** - 2 low gaps, has good ARIA foundation
5. **Dropdown** - 5 gaps, has menu pattern, needs trigger ARIA

---

## Appendix: WCAG 2.1 AA Success Criteria Coverage

| Criterion | Name | Status | Notes |
|-----------|------|--------|-------|
| 1.1.1 | Non-text Content | Partial | Icon accessibility needed |
| 1.3.1 | Info and Relationships | **Failing** | Label associations missing |
| 1.3.2 | Meaningful Sequence | Pass | DOM order is logical |
| 1.3.3 | Sensory Characteristics | Pass | No sensory-only instructions |
| 1.3.4 | Orientation | Pass | Works in both orientations |
| 1.3.5 | Identify Input Purpose | Partial | Autocomplete not supported |
| 1.4.1 | Use of Color | Pass | Not color-only |
| 1.4.3 | Contrast (Minimum) | Untested | Needs visual audit |
| 1.4.4 | Resize Text | Pass | Uses relative units |
| 1.4.5 | Images of Text | Pass | No images of text |
| 1.4.10 | Reflow | Pass | Responsive design |
| 1.4.11 | Non-text Contrast | Untested | Needs visual audit |
| 1.4.12 | Text Spacing | Pass | CSS handles spacing |
| 1.4.13 | Content on Hover | Partial | Tooltips need review |
| 2.1.1 | Keyboard | **Failing** | Several components lack keyboard |
| 2.1.2 | No Keyboard Trap | **Failing** | Modal lacks focus trap |
| 2.1.4 | Character Key Shortcuts | Pass | No single-key shortcuts |
| 2.4.1 | Bypass Blocks | Partial | No skip links |
| 2.4.2 | Page Titled | N/A | Application handles |
| 2.4.3 | Focus Order | Partial | Focus management issues |
| 2.4.4 | Link Purpose | Pass | Links have context |
| 2.4.5 | Multiple Ways | N/A | Application handles |
| 2.4.6 | Headings and Labels | Pass | Proper heading structure |
| 2.4.7 | Focus Visible | Partial | Some components need review |
| 3.1.1 | Language of Page | N/A | Application handles |
| 3.1.2 | Language of Parts | N/A | Single language |
| 3.2.1 | On Focus | Pass | No focus-triggered changes |
| 3.2.2 | On Input | Pass | No unexpected changes |
| 3.2.3 | Consistent Navigation | Pass | Sidebar consistent |
| 3.2.4 | Consistent Identification | Pass | Components consistent |
| 3.3.1 | Error Identification | **Failing** | Errors not accessible |
| 3.3.2 | Labels or Instructions | Partial | Some labels missing |
| 3.3.3 | Error Suggestion | Partial | Validation exists |
| 3.3.4 | Error Prevention | Pass | Confirm dialogs exist |
| 4.1.1 | Parsing | Pass | Valid HTML |
| 4.1.2 | Name, Role, Value | **Failing** | Many ARIA gaps |
| 4.1.3 | Status Messages | **Failing** | No live regions |

**Legend:** Pass = Compliant | Partial = Some issues | **Failing** = Major gaps | Untested = Needs manual verification | N/A = Not applicable at component level

---

*Generated by Phase 07 Accessibility Audit*
*GoQuery v1.1 Accessibility Milestone*
