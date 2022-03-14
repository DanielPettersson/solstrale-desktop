// Package controller provies controllers for handling the raytracing engine
package controller

import (
	"github.com/DanielPettersson/solstrale"
	"github.com/DanielPettersson/solstrale/renderer"
)

// TraceController is used to control the flow of raytracing
// Allowing multithreaded code to safely update and stopping the rendering
type TraceController struct {
	update        chan bool
	stop          chan bool
	exit          chan bool
	getScene      func() *renderer.Scene
	progress      func(renderer.RenderProgress)
	renderStarted func()
	renderStopped func()
}

// NewTraceController creates a new TraceController with supplied
// callback hooks for rendering events
func NewTraceController(
	getScene func() *renderer.Scene,
	progress func(renderer.RenderProgress),
	renderStarted func(),
	renderStopped func(),
) *TraceController {
	tc := TraceController{
		update:        make(chan bool, 100),
		stop:          make(chan bool),
		exit:          make(chan bool),
		getScene:      getScene,
		progress:      progress,
		renderStarted: renderStarted,
		renderStopped: renderStopped,
	}
	go tc.loop()
	return &tc
}

func (tc *TraceController) loop() {

	// Renderloop for controlling the render
	for {
		// Wait for either and update to go ahead and render
		// or a exit command to quit
		select {
		case <-tc.update:
		case <-tc.exit:
			return
		}

		// Here we consume all update messages as to not restart rendering
		// more times than neeeded.
	EatAllUpdates:
		for {
			select {
			case <-tc.update:
			default:
				break EatAllUpdates
			}
		}

		// Do the actual rendering
		renderProgress := make(chan renderer.RenderProgress, 1)
		abortRender := make(chan bool, 1)

		tc.renderStarted()

		go solstrale.RayTrace(tc.getScene(), renderProgress, abortRender)

		// Get the progress
		for p := range renderProgress {
			tc.progress(p)
			select {
			// When an update command, abort the current render
			// and add another update to restart rendering in next loop
			case <-tc.update:
				abortRender <- true
				tc.update <- true
			// Just abort the rendering.
			// Then we will wait for an update or exit in the loop
			case <-tc.stop:
				abortRender <- true
			// Exit, abort and quit the loop
			case <-tc.exit:
				abortRender <- true
				return
			default:
			}
		}
		tc.renderStopped()
	}

}

func (tc *TraceController) Update() {
	tc.update <- true
}

func (tc *TraceController) Stop() {
	tc.stop <- true
}

func (tc *TraceController) Exit() {
	tc.exit <- true
}
