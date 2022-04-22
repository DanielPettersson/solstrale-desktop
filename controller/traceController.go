// Package controller provies controllers for handling the raytracing engine
package controller

import (
	"github.com/DanielPettersson/solstrale"
	"github.com/DanielPettersson/solstrale-desktop/scene"
	"github.com/DanielPettersson/solstrale/renderer"
)

// TraceController is used to control the flow of raytracing
// Allowing multithreaded code to safely update and stopping the rendering
type TraceController struct {
	update       chan bool
	exit         chan bool
	getImageSize func() (int, int)
	progress     func(renderer.RenderProgress)
	renderError  func(error)
}

// NewTraceController creates a new TraceController with supplied
// callback hooks for rendering events
func NewTraceController(
	getImageSize func() (int, int),
	progress func(renderer.RenderProgress),
	renderError func(error),
) *TraceController {
	tc := TraceController{
		update:       make(chan bool, 1000),
		exit:         make(chan bool, 1),
		getImageSize: getImageSize,
		progress:     progress,
		renderError:  renderError,
	}
	go tc.loop()
	return &tc
}

func (tc *TraceController) loop() {

	scene, err := scene.Scene()
	if err != nil {
		tc.renderError(err)
		return
	}

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

		imageWidth, imageHeight := tc.getImageSize()
		go solstrale.RayTrace(imageWidth, imageHeight, scene, renderProgress, abortRender)

		// Get the progress
		for p := range renderProgress {

			if p.Error != nil {
				tc.renderError(p.Error)
				return
			}

			tc.progress(p)
			select {
			// When an update command, abort the current render
			// and add another update to restart rendering in next loop
			case <-tc.update:
				abortRender <- true
				tc.update <- true
			// Exit, abort and quit the loop
			case <-tc.exit:
				abortRender <- true
				return
			default:
			}
		}
	}
}

// Update aborts current rendering and starts new
func (tc *TraceController) Update() {
	tc.update <- true
}

// Exit stops rendering and quits render loop
func (tc *TraceController) Exit() {
	tc.exit <- true
}
