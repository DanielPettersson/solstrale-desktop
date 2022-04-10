package main

import (
	_ "embed"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/DanielPettersson/solstrale-desktop/controller"
	"github.com/DanielPettersson/solstrale/renderer"
	_ "github.com/robertkrimen/otto/underscore"
)

var (
	//go:embed default-scene.js
	defaultScene string
)

func main() {

	rasterW := 800
	rasterH := 600

	app := app.New()
	window := app.NewWindow("Solstråle")
	window.Resize(fyne.Size{
		Width:  float32(rasterW),
		Height: float32(rasterH),
	})

	var renderImage image.Image
	renderImage = image.NewRGBA(image.Rect(0, 0, 1, 1))

	raster := canvas.Raster{}

	progress := widget.NewProgressBar()

	runButton := widget.Button{
		Text: "Run",
	}
	stopButton := widget.Button{
		Text: "Stop",
	}
	stopButton.Disable()

	jsonInputEntry := widget.NewMultiLineEntry()
	jsonInputEntry.SetText(defaultScene)

	traceController := controller.NewTraceController(
		func() (string, int, int) {
			return jsonInputEntry.Text, rasterW, rasterH
		},
		func(rp renderer.RenderProgress) {
			renderImage = rp.RenderImage
			progress.SetValue(rp.Progress)
			raster.Refresh()
		},
		func() {
			runButton.Disable()
			stopButton.Enable()
		},
		func() {
			runButton.Enable()
			stopButton.Disable()
		},
		func(err error) {
			dialog.ShowError(err, window)
		},
	)

	raster.Generator =
		func(w, h int) image.Image {

			if w != rasterW || h != rasterH {
				rasterW = w
				rasterH = h
				traceController.Update()
			}

			return renderImage
		}

	runButton.OnTapped = func() {
		traceController.Update()
	}

	stopButton.OnTapped = func() {
		traceController.Stop()
	}

	topBar := container.New(layout.NewHBoxLayout(), &runButton, &stopButton)

	tabsContainer := container.NewAppTabs(
		container.NewTabItem("Input", jsonInputEntry),
		container.NewTabItem("Output", &raster),
	)
	tabsContainer.SelectIndex(1)

	container := container.New(layout.NewBorderLayout(topBar, progress, nil, nil),
		topBar, progress, tabsContainer)

	window.SetContent(container)
	window.ShowAndRun()

	traceController.Exit()
}
