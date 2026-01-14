//go:build js && wasm

package components

import (
	"fmt"
	"syscall/js"
)

// StepStatus defines the state of a step
type StepStatus string

const (
	StepPending  StepStatus = "pending"
	StepCurrent  StepStatus = "current"
	StepComplete StepStatus = "complete"
	StepError    StepStatus = "error"
)

// Step represents a single step in a stepper
type Step struct {
	Title       string
	Description string
	Content     js.Value // content to show when step is active
}

// StepperProps configures a Stepper
type StepperProps struct {
	Steps       []Step
	CurrentStep int
	Vertical    bool
	OnStepClick func(step int) // optional: make steps clickable
}

// Stepper is a multi-step wizard component
type Stepper struct {
	element     js.Value
	contentArea js.Value
	steps       []Step
	current     int
	stepEls     []js.Value
	onComplete  func()
}

// NewStepper creates a new Stepper component
func NewStepper(props StepperProps) *Stepper {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "w-full")

	s := &Stepper{
		element: container,
		steps:   props.Steps,
		current: props.CurrentStep,
		stepEls: make([]js.Value, len(props.Steps)),
	}

	// Step indicators
	var stepsClass string
	if props.Vertical {
		stepsClass = "flex flex-col space-y-4"
	} else {
		stepsClass = "flex items-center justify-between mb-8"
	}

	stepsContainer := document.Call("createElement", "div")
	stepsContainer.Set("className", stepsClass)

	for i, step := range props.Steps {
		stepEl := s.createStepIndicator(i, step, props.Vertical, props.OnStepClick)
		stepsContainer.Call("appendChild", stepEl)
		s.stepEls[i] = stepEl

		// Connector line (except after last step)
		if i < len(props.Steps)-1 && !props.Vertical {
			connector := document.Call("createElement", "div")
			connector.Set("className", "flex-1 h-0.5 bg-gray-200 mx-4")
			stepsContainer.Call("appendChild", connector)
		}
	}

	container.Call("appendChild", stepsContainer)

	// Content area
	contentArea := document.Call("createElement", "div")
	contentArea.Set("className", "mt-4")
	if len(props.Steps) > 0 && props.Steps[props.CurrentStep].Content.Truthy() {
		contentArea.Call("appendChild", props.Steps[props.CurrentStep].Content)
	}
	container.Call("appendChild", contentArea)
	s.contentArea = contentArea

	return s
}

func (s *Stepper) createStepIndicator(index int, step Step, vertical bool, onClick func(int)) js.Value {
	document := js.Global().Get("document")

	var status StepStatus
	if index < s.current {
		status = StepComplete
	} else if index == s.current {
		status = StepCurrent
	} else {
		status = StepPending
	}

	container := document.Call("createElement", "div")
	containerClass := "flex items-center"
	if vertical {
		containerClass += " flex-row"
	} else {
		containerClass += " flex-col"
	}
	if onClick != nil {
		containerClass += " cursor-pointer"
	}
	container.Set("className", containerClass)

	// Circle indicator
	circle := document.Call("createElement", "div")
	var circleClass string
	switch status {
	case StepComplete:
		circleClass = "w-8 h-8 rounded-full bg-green-500 text-white flex items-center justify-center"
	case StepCurrent:
		circleClass = "w-8 h-8 rounded-full bg-blue-500 text-white flex items-center justify-center ring-4 ring-blue-100"
	case StepError:
		circleClass = "w-8 h-8 rounded-full bg-red-500 text-white flex items-center justify-center"
	default:
		circleClass = "w-8 h-8 rounded-full bg-gray-200 text-gray-500 flex items-center justify-center"
	}
	circle.Set("className", circleClass)

	if status == StepComplete {
		circle.Set("innerHTML", `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>`)
	} else {
		circle.Set("textContent", fmt.Sprintf("%d", index+1))
	}

	container.Call("appendChild", circle)

	// Labels
	labels := document.Call("createElement", "div")
	if vertical {
		labels.Set("className", "ml-3")
	} else {
		labels.Set("className", "mt-2 text-center")
	}

	title := document.Call("createElement", "div")
	titleClass := "text-sm font-medium"
	if status == StepCurrent {
		titleClass += " text-blue-600"
	} else if status == StepComplete {
		titleClass += " text-green-600"
	} else {
		titleClass += " text-gray-500"
	}
	title.Set("className", titleClass)
	title.Set("textContent", step.Title)
	labels.Call("appendChild", title)

	if step.Description != "" {
		desc := document.Call("createElement", "div")
		desc.Set("className", "text-xs text-gray-400")
		desc.Set("textContent", step.Description)
		labels.Call("appendChild", desc)
	}

	container.Call("appendChild", labels)

	// Click handler
	if onClick != nil {
		container.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			onClick(index)
			return nil
		}))
	}

	return container
}

// Element returns the DOM element
func (s *Stepper) Element() js.Value {
	return s.element
}

// GoTo navigates to a specific step
func (s *Stepper) GoTo(step int) {
	if step < 0 || step >= len(s.steps) {
		return
	}

	s.current = step
	s.updateIndicators()
	s.updateContent()
}

// Next advances to the next step
func (s *Stepper) Next() bool {
	if s.current >= len(s.steps)-1 {
		if s.onComplete != nil {
			s.onComplete()
		}
		return false
	}

	s.current++
	s.updateIndicators()
	s.updateContent()
	return true
}

// Prev goes back to the previous step
func (s *Stepper) Prev() bool {
	if s.current <= 0 {
		return false
	}

	s.current--
	s.updateIndicators()
	s.updateContent()
	return true
}

// Current returns the current step index
func (s *Stepper) Current() int {
	return s.current
}

// OnComplete sets a callback for when the last step is completed
func (s *Stepper) OnComplete(fn func()) {
	s.onComplete = fn
}

func (s *Stepper) updateIndicators() {
	for i := range s.stepEls {
		var status StepStatus
		if i < s.current {
			status = StepComplete
		} else if i == s.current {
			status = StepCurrent
		} else {
			status = StepPending
		}

		circle := s.stepEls[i].Get("firstChild")
		title := s.stepEls[i].Get("lastChild").Get("firstChild")

		// Update circle
		switch status {
		case StepComplete:
			circle.Set("className", "w-8 h-8 rounded-full bg-green-500 text-white flex items-center justify-center")
			circle.Set("innerHTML", `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>`)
			title.Set("className", "text-sm font-medium text-green-600")
		case StepCurrent:
			circle.Set("className", "w-8 h-8 rounded-full bg-blue-500 text-white flex items-center justify-center ring-4 ring-blue-100")
			circle.Set("textContent", fmt.Sprintf("%d", i+1))
			title.Set("className", "text-sm font-medium text-blue-600")
		default:
			circle.Set("className", "w-8 h-8 rounded-full bg-gray-200 text-gray-500 flex items-center justify-center")
			circle.Set("textContent", fmt.Sprintf("%d", i+1))
			title.Set("className", "text-sm font-medium text-gray-500")
		}
	}
}

func (s *Stepper) updateContent() {
	// Clear content area
	s.contentArea.Set("innerHTML", "")

	// Add new content
	if s.steps[s.current].Content.Truthy() {
		s.contentArea.Call("appendChild", s.steps[s.current].Content)
	}
}
