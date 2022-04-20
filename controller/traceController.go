// Package controller provies controllers for handling the raytracing engine
package controller

import (
	_ "embed"

	"github.com/DanielPettersson/solstrale"
	"github.com/DanielPettersson/solstrale-desktop/scene"
	"github.com/DanielPettersson/solstrale/renderer"
)

// TraceController is used to control the flow of raytracing
// Allowing multithreaded code to safely update and stopping the rendering
type TraceController struct {
	update        chan bool
	stop          chan bool
	exit          chan bool
	getImageSize  func() (int, int)
	progress      func(renderer.RenderProgress)
	buildingScene func()
	renderStarted func()
	renderStopped func()
	renderError   func(error)
}

// NewTraceController creates a new TraceController with supplied
// callback hooks for rendering events
func NewTraceController(
	getImageSize func() (int, int),
	progress func(renderer.RenderProgress),
	buildingScene func(),
	renderStarted func(),
	renderStopped func(),
	renderError func(error),
) *TraceController {
	tc := TraceController{
		update:        make(chan bool, 1000),
		stop:          make(chan bool, 1),
		exit:          make(chan bool, 1),
		getImageSize:  getImageSize,
		progress:      progress,
		buildingScene: buildingScene,
		renderStarted: renderStarted,
		renderStopped: renderStopped,
		renderError:   renderError,
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

		tc.buildingScene()

		scene, err := scene.Scene(tc.getImageSize())
		if err != nil {
			tc.renderError(err)
		}

		if scene != nil {

			tc.renderStarted()
			go solstrale.RayTrace(scene, renderProgress, abortRender)

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
		}
		tc.renderStopped()
	}
}

// Update aborts current rendering and starts new
func (tc *TraceController) Update() {
	tc.update <- true
}

// Stop rendering
func (tc *TraceController) Stop() {
	tc.stop <- true
}

// Exit stops rendering and quits render loop
func (tc *TraceController) Exit() {
	tc.exit <- true
}
