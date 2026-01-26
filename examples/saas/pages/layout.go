package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// DashboardLayout wraps content with sidebar navigation and dark theme.
// Takes the router for state management and current path for active nav highlighting.
func DashboardLayout(r *core.Router, currentPath string, children ...core.Node) core.Node {
	// Sidebar state
	collapsed := r.StateBool("sidebarCollapsed", false)
	mobileOpen := r.StateBool("sidebarMobileOpen", false)

	// Build nav items
	navItems := buildNavItems(currentPath, collapsed.Get())

	return ui.SidebarLayout(ui.SidebarLayoutProps{
		Sidebar: ui.Sidebar(ui.SidebarProps{
			Collapsed:  collapsed.Get(),
			MobileOpen: mobileOpen.Get(),
			OnCloseMobile: func() {
				mobileOpen.Set(false)
			},
			Children: []core.Node{
				// Header with brand
				ui.SidebarHeader(ui.SidebarHeaderProps{
					Title:     "SaaS App",
					Collapsed: collapsed.Get(),
					OnToggleCollapse: func() {
						collapsed.Set(!collapsed.Get())
					},
					OnCloseMobile: func() {
						mobileOpen.Set(false)
					},
				}),
				// Navigation
				ui.SidebarNav(ui.SidebarNavProps{
					Children: navItems,
				}),
				// Footer with user profile
				ui.SidebarFooter(ui.SidebarFooterProps{
					Collapsed: collapsed.Get(),
					Children: []core.Node{
						ui.SidebarUser(ui.SidebarUserProps{
							Name:      "John Doe",
							Email:     "john@example.com",
							Href:      "/profile",
							Collapsed: collapsed.Get(),
						}),
					},
				}),
			},
		}),
		Children: []core.Node{
			ui.SidebarMain(ui.SidebarMainProps{
				Collapsed:    collapsed.Get(),
				MobileOpen:   mobileOpen.Get(),
				OnOpenMobile: func() { mobileOpen.Set(true) },
				HeaderTitle:  "SaaS App",
				Children:     children,
			}),
		},
	})
}

// buildNavItems creates the sidebar navigation items.
func buildNavItems(currentPath string, collapsed bool) []core.Node {
	items := []struct {
		label string
		href  string
		icon  string
	}{
		{"Dashboard", "/", "home"},
		{"Projects", "/projects", "folder"},
		{"Settings", "/settings", "settings"},
	}

	var nodes []core.Node
	for _, item := range items {
		// Check if this is the active route
		active := currentPath == item.href
		// Also check for sub-paths (e.g., /projects/new should highlight Projects)
		if !active && item.href != "/" {
			active = len(currentPath) > len(item.href) && currentPath[:len(item.href)] == item.href
		}

		nodes = append(nodes, ui.SidebarItem(ui.SidebarItemProps{
			Label:     item.label,
			Href:      item.href,
			Icon:      item.icon,
			Active:    active,
			Collapsed: collapsed,
		}))
	}
	return nodes
}
