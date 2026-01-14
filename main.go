//go:build js && wasm

package main

import (
	"context"
	"syscall/js"

	"goquery/internal/api"
	"goquery/internal/components"
)

var (
	layout  *components.Layout
	display *components.DataDisplay
	router  *components.Router
	posts   *api.PostsClient
)

func main() {
	// Load Tailwind CSS
	components.LoadTailwind()

	document := js.Global().Get("document")

	// Reset body styles for full-page layout
	body := document.Get("body")
	body.Set("className", "m-0 p-0")

	// Clear app container
	app := document.Call("getElementById", "app")
	app.Set("innerHTML", "")

	// Initialize API clients (uses same origin by default)
	posts = api.NewPostsClient()

	// Create router
	router = components.NewRouter()
	router.Register("/", showDashboard)
	router.Register("/api-test", showAPITest)
	router.Register("/settings", showSettings)

	// Update sidebar active state on navigation
	router.OnNavigate(func(path string) {
		if layout != nil {
			layout.Sidebar().SetActive(path)
		}
	})

	// Create data display (used by API Test page)
	display = components.NewDataDisplay()

	// Create layout
	layout = components.NewLayout(components.LayoutProps{
		Sidebar: components.SidebarProps{
			Title: "Admin Panel",
			Items: []components.NavItem{
				{Label: "Dashboard", Icon: "📊", Path: "/"},
				{Label: "API Test", Icon: "🔌", Path: "/api-test"},
				{Label: "Settings", Icon: "⚙️", Path: "/settings"},
			},
		},
		Header: components.HeaderProps{
			Title: "Dashboard",
			Actions: []components.HeaderAction{
				{Label: "Refresh", OnClick: func() { js.Global().Get("location").Call("reload") }},
			},
		},
	})

	app.Call("appendChild", layout.Element())

	// Start router (handles initial URL)
	router.Start()

	// Keep the Go program running
	select {}
}

func showDashboard() {
	document := js.Global().Get("document")

	welcome := document.Call("createElement", "div")
	welcome.Set("className", "bg-white rounded-lg shadow p-6")

	h2 := document.Call("createElement", "h2")
	h2.Set("className", "text-lg font-semibold mb-4")
	h2.Set("textContent", "Welcome to the Admin Dashboard")
	welcome.Call("appendChild", h2)

	p := document.Call("createElement", "p")
	p.Set("className", "text-gray-600")
	p.Set("textContent", "This is a proof of concept admin dashboard built entirely with Go WASM. Select an option from the sidebar to explore.")
	welcome.Call("appendChild", p)

	layout.SetContent(welcome)
}

func showAPITest() {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "bg-white rounded-lg shadow p-6")

	h2 := document.Call("createElement", "h2")
	h2.Set("className", "text-lg font-semibold mb-4")
	h2.Set("textContent", "API Test")
	container.Call("appendChild", h2)

	// Fetch single post button
	fetchOneBtn := components.Button(components.ButtonProps{
		Text: "Fetch Post #1",
		OnClick: func() {
			go fetchSinglePost()
		},
	})
	container.Call("appendChild", fetchOneBtn)

	// Fetch all posts button
	fetchAllBtn := components.Button(components.ButtonProps{
		Text:      "Fetch All Posts",
		ClassName: "px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 cursor-pointer ml-2 transition-colors",
		OnClick: func() {
			go fetchAllPosts()
		},
	})
	container.Call("appendChild", fetchAllBtn)

	container.Call("appendChild", display.Element())

	layout.SetContent(container)
}

func showSettings() {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "bg-white rounded-lg shadow p-6")

	h2 := document.Call("createElement", "h2")
	h2.Set("className", "text-lg font-semibold mb-4")
	h2.Set("textContent", "Settings")
	container.Call("appendChild", h2)

	p := document.Call("createElement", "p")
	p.Set("className", "text-gray-600")
	p.Set("textContent", "Settings page - add form components here.")
	container.Call("appendChild", p)

	layout.SetContent(container)
}

func fetchSinglePost() {
	display.ShowLoading("Fetching post #1...")

	// Clean API call!
	post, err := posts.GetByID(context.Background(), 1)
	if err != nil {
		display.ShowError("Error: " + err.Error())
		return
	}

	display.ShowJSON(post)
}

func fetchAllPosts() {
	display.ShowLoading("Fetching all posts...")

	// Clean API call!
	allPosts, err := posts.GetAll(context.Background())
	if err != nil {
		display.ShowError("Error: " + err.Error())
		return
	}

	// Show first 5 posts
	if len(allPosts) > 5 {
		allPosts = allPosts[:5]
	}
	display.ShowJSON(allPosts)
}
