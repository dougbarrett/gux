package pages

import (
	"strings"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// AdminLayout wraps admin pages with sidebar navigation.
func AdminLayout(page func(*core.Router) func() core.Node) func(*core.Router) func() core.Node {
	return func(r *core.Router) func() core.Node {
		// Get the inner page component
		pageComponent := page(r)

		return func() core.Node {
			// Sidebar state
			collapsed := r.StateBool("collapsed", false)
			mobileOpen := r.StateBool("mobileOpen", false)
			currentPath := r.Path()

			// Navigation items
			navItems := []ui.SidebarItemProps{
				{Label: "Dashboard", Href: "/admin", Icon: "home", Active: currentPath == "/admin"},
				{Label: "Users", Href: "/admin/users", Icon: "users", Active: strings.HasPrefix(currentPath, "/admin/users")},
				{Label: "Settings", Href: "/admin/settings", Icon: "settings", Active: currentPath == "/admin/settings"},
			}

			// Build sidebar items
			sidebarItems := make([]core.Node, len(navItems))
			for i, item := range navItems {
				item.Collapsed = collapsed.Get()
				sidebarItems[i] = ui.SidebarItem(item)
			}

			// Get user info
			userName := "Admin"
			userEmail := "admin@example.com"
			if r.User() != nil {
				userName = r.User().Name
				userEmail = r.User().Email
			}

			return ui.SidebarLayout(ui.SidebarLayoutProps{
				Sidebar: ui.Sidebar(ui.SidebarProps{
					Collapsed:     collapsed.Get(),
					MobileOpen:    mobileOpen.Get(),
					OnCloseMobile: func() { mobileOpen.Set(false) },
					Children: []core.Node{
						// Header with logo/title
						ui.SidebarHeader(ui.SidebarHeaderProps{
							Title:            "Test Admin",
							Collapsed:        collapsed.Get(),
							OnToggleCollapse: func() { collapsed.Set(!collapsed.Get()) },
							OnCloseMobile:    func() { mobileOpen.Set(false) },
						}),
						// Navigation
						ui.SidebarNav(ui.SidebarNavProps{
							Children: []core.Node{
								ui.SidebarSection(ui.SidebarSectionProps{
									Title:     "Main",
									Collapsed: collapsed.Get(),
									Children:  sidebarItems,
								}),
							},
						}),
						// Footer with user info
						ui.SidebarFooter(ui.SidebarFooterProps{
							Collapsed: collapsed.Get(),
							Children: []core.Node{
								ui.SidebarUser(ui.SidebarUserProps{
									Name:      userName,
									Email:     userEmail,
									Href:      "/admin/profile",
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
						HeaderTitle:  "Test Admin",
						Children: []core.Node{
							// Render the actual page content
							pageComponent(),
						},
					}),
				},
			})
		}
	}
}
