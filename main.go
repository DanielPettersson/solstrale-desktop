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

	traceController := controller.NewTraceController(
		func() (int, int) {
			return rasterW, rasterH
		},
		func(rp renderer.RenderProgress) {

			renderImage = rp.RenderImage
			progress.SetValue(rp.Progress)
			raster.Refresh()
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

	container := container.New(layout.NewBorderLayout(nil, progress, nil, nil),
		progress, &raster)

	window.SetContent(container)
	window.ShowAndRun()

	traceController.Exit()
}
