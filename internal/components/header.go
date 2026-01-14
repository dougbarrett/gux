//go:build js && wasm

package components

import "syscall/js"

type HeaderProps struct {
	Title   string
	Actions []HeaderAction
}

type HeaderAction struct {
	Label   string
	OnClick func()
}

const (
	headerClass        = "h-16 bg-white border-b border-gray-200 flex items-center justify-between px-6"
	headerTitleClass   = "text-xl font-semibold text-gray-800"
	headerActionsClass = "flex items-center gap-3"
	headerBtnClass     = "px-3 py-1.5 text-sm bg-gray-100 text-gray-700 rounded hover:bg-gray-200 cursor-pointer transition-colors"
)

func Header(props HeaderProps) js.Value {
	document := js.Global().Get("document")

	header := document.Call("createElement", "header")
	header.Set("className", headerClass)

	// Title
	title := document.Call("createElement", "h1")
	title.Set("className", headerTitleClass)
	title.Set("textContent", props.Title)
	header.Call("appendChild", title)

	// Actions
	if len(props.Actions) > 0 {
		actions := document.Call("createElement", "div")
		actions.Set("className", headerActionsClass)

		for _, action := range props.Actions {
			btn := document.Call("createElement", "button")
			btn.Set("className", headerBtnClass)
			btn.Set("textContent", action.Label)

			if action.OnClick != nil {
				fn := action.OnClick
				btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
					fn()
					return nil
				}))
			}

			actions.Call("appendChild", btn)
		}

		header.Call("appendChild", actions)
	}

	return header
}
