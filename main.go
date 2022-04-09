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
	solstralejson "github.com/DanielPettersson/solstrale-json"
	"github.com/DanielPettersson/solstrale/renderer"
)

var (
	//go:embed default-scene.json
	defaultScene string
)

func main() {

	app := app.New()
	window := app.NewWindow("Solstråle")
	window.Resize(fyne.Size{
		Width:  800,
		Height: 600,
	})

	var renderImage image.Image
	renderImage = image.NewRGBA(image.Rect(0, 0, 1, 1))

	rasterW := 0
	rasterH := 0
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
		func() *renderer.Scene {
			//height := int(math.Round(float64(raster.Size().Height)))
			//width := int(math.Round(float64(raster.Size().Width)))

			scene, err := solstralejson.ToScene([]byte(jsonInputEntry.Text))
			if err != nil {
				dialog.ShowError(err, window)
			}
			return scene
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

	container := container.New(layout.NewBorderLayout(topBar, progress, nil, nil),
		topBar, progress, tabsContainer)

	window.SetContent(container)
	window.ShowAndRun()

	traceController.Exit()
}
