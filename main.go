package main

import (
	"image"
	"math"
	"math/rand"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/DanielPettersson/solstrale-desktop/controller"
	"github.com/DanielPettersson/solstrale/camera"
	"github.com/DanielPettersson/solstrale/geo"
	"github.com/DanielPettersson/solstrale/hittable"
	"github.com/DanielPettersson/solstrale/material"
	"github.com/DanielPettersson/solstrale/spec"
)

var (
	grassImage   *image.Image  = nil
	cameraX      binding.Float = nil
	fieldOfView  binding.Float = nil
	apertureSize binding.Float = nil
)

func main() {
	rand.Seed(time.Now().UnixNano())
	loadTextures()

	app := app.New()
	window := app.NewWindow("Solstråle")
	window.Resize(fyne.Size{
		Width:  800,
		Height: 450,
	})

	var renderImage image.Image
	renderImage = image.NewRGBA(image.Rect(0, 0, 1, 1))
	raster := canvas.NewRaster(
		func(w, h int) image.Image {
			return renderImage
		})

	progress := widget.NewProgressBar()

	runButton := widget.Button{
		Text: "Run",
	}
	stopButton := widget.Button{
		Text: "Stop",
	}
	stopButton.Disable()

	traceController := controller.NewTraceController(
		func() *spec.Scene {
			height := int(math.Round(float64(raster.Size().Height)))
			width := int(math.Round(float64(raster.Size().Width)))

			traceSpec := spec.TraceSpecification{
				ImageWidth:      width,
				ImageHeight:     height,
				SamplesPerPixel: 1000,
				MaxDepth:        50,
				RandomSeed:      0,
			}

			return TestScene(traceSpec)
		},
		func(tp spec.TraceProgress) {
			renderImage = tp.RenderImage
			progress.SetValue(tp.Progress)
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

	runButton.OnTapped = func() {
		traceController.Update()
	}

	stopButton.OnTapped = func() {
		traceController.Stop()
	}

	cameraX = binding.NewFloat()
	cameraX.Set(-5)
	cameraXLabel, cameraXSlider := sliderWithLabel(cameraX, traceController, "Camera X: %0.1f", -5, 5)

	fieldOfView = binding.NewFloat()
	fieldOfView.Set(20.)
	fieldOfViewLabel, fieldOfViewSlider := sliderWithLabel(fieldOfView, traceController, "Field of View: %0.1f", 1, 70)

	apertureSize = binding.NewFloat()
	apertureSize.Set(.8)
	apertureSizeLabel, apertureSizeSlider := sliderWithLabel(apertureSize, traceController, "Aperture Size: %0.1f", 0, 3)

	topBar := container.New(layout.NewHBoxLayout(), &runButton, &stopButton)
	leftBar := container.New(
		layout.NewVBoxLayout(),
		cameraXLabel,
		cameraXSlider,
		fieldOfViewLabel,
		fieldOfViewSlider,
		apertureSizeLabel,
		apertureSizeSlider,
	)

	container := container.New(layout.NewBorderLayout(topBar, progress, leftBar, nil),
		topBar, progress, leftBar, raster)

	window.SetContent(container)
	window.ShowAndRun()

	traceController.Exit()
}

func sliderWithLabel(
	b binding.Float,
	traceController *controller.TraceController,
	format string,
	min, max float64,
) (*widget.Label, *widget.Slider) {
	labelValue := binding.FloatToStringWithFormat(b, format)
	label := widget.NewLabelWithData(labelValue)

	slider := widget.NewSliderWithData(min, max, b)
	slider.Step = .1
	slider.OnChanged = func(value float64) {
		currVal, _ := b.Get()
		if math.Abs(currVal-value) > slider.Step/2 {
			b.Set(value)
			traceController.Update()
		}
	}
	return label, slider
}

func loadTextures() {
	f, _ := os.Open("grass.jpg")
	defer f.Close()
	gi, _, _ := image.Decode(f)
	grassImage = &gi
}

func TestScene(traceSpec spec.TraceSpecification) *spec.Scene {

	cameraXValue, _ := cameraX.Get()
	fieldOfViewValue, _ := fieldOfView.Get()
	apertureSizeValue, _ := apertureSize.Get()

	lookFrom := geo.NewVec3(cameraXValue, 3, 6)
	lookAt := geo.NewVec3(0, 1, 0)
	lookLength := lookFrom.Sub(lookAt).Length()

	camera := camera.New(
		traceSpec.ImageWidth,
		traceSpec.ImageHeight,
		fieldOfViewValue,
		apertureSizeValue,
		lookLength,
		lookFrom,
		lookAt,
		geo.NewVec3(0, 1, 0),
	)

	world := hittable.NewHittableList()

	checkerTex := material.CheckerTexture{
		Scale: 0.1,
		Even:  material.SolidColor{ColorValue: geo.NewVec3(0.4, 0.2, 0.1)},
		Odd:   material.SolidColor{ColorValue: geo.NewVec3(0.8, 0.4, 0.2)},
	}
	noiseTex := material.NoiseTexture{
		ColorValue: geo.NewVec3(1, 1, 0),
		Scale:      .01,
	}

	imageTex := material.ImageTexture{
		Image:  *grassImage,
		Mirror: false,
	}

	groundMaterial := material.Lambertian{Tex: imageTex}
	checkerMat := material.Lambertian{Tex: checkerTex}
	glassMat := material.Dielectric{Tex: material.SolidColor{ColorValue: geo.NewVec3(1, 1, 1)}, IndexOfRefraction: 1.5}
	goldMat := material.Metal{Tex: noiseTex, Fuzz: .2}
	lightMat := material.DiffuseLight{Emit: material.SolidColor{ColorValue: geo.NewVec3(5, 5, 5)}}

	world.Add(hittable.NewQuad(
		geo.NewVec3(-5, 0, -5), geo.NewVec3(10, 0, 0), geo.NewVec3(0, 0, 10),
		groundMaterial,
	))
	world.Add(hittable.NewSphere(geo.NewVec3(-1.5, 1, 0), 1, glassMat))
	world.Add(hittable.NewBox(geo.NewVec3(-.5, 0, -.5), geo.NewVec3(.5, 2, .5), checkerMat))
	world.Add(hittable.NewSphere(geo.NewVec3(1.5, 1, 0), 1, goldMat))

	world.Add(hittable.NewSphere(geo.NewVec3(10, 5, 16), 10, lightMat))
	world.Add(hittable.NewSphere(geo.NewVec3(-10, 7, 16), 3, lightMat))

	return &spec.Scene{
		World:           &world,
		Cam:             camera,
		BackgroundColor: geo.NewVec3(.05, .1, .2),
		Spec:            traceSpec,
	}

}
