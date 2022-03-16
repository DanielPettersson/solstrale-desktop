package main

import (
	"image"
	"math"
	"os"

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
	"github.com/DanielPettersson/solstrale/renderer"
)

var (
	grassImage     *image.Image  = nil
	cameraX        binding.Float = nil
	fieldOfView    binding.Float = nil
	apertureSize   binding.Float = nil
	lightSize      binding.Float = nil
	lightIntensity binding.Float = nil
)

func main() {
	loadTextures()

	app := app.New()
	window := app.NewWindow("Solstråle")
	window.Resize(fyne.Size{
		Width:  600,
		Height: 400,
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

	shaderSelect := widget.Select{
		Options: []string{"PathTracing", "Simple"},
	}
	shaderSelect.SetSelectedIndex(0)

	traceController := controller.NewTraceController(
		func() *renderer.Scene {
			height := int(math.Round(float64(raster.Size().Height)))
			width := int(math.Round(float64(raster.Size().Width)))

			var shader renderer.Shader
			if shaderSelect.SelectedIndex() == 0 {
				shader = renderer.PathTracingShader{MaxDepth: 50}
			} else {
				shader = renderer.SimpleShader{}
			}

			renderConfig := renderer.RenderConfig{
				ImageWidth:      width,
				ImageHeight:     height,
				SamplesPerPixel: 1000,
				Shader:          shader,
			}

			return TestScene(renderConfig)
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

	shaderSelect.OnChanged = func(s string) {
		traceController.Update()
	}

	cameraX = binding.NewFloat()
	cameraX.Set(278)
	cameraXLabel, cameraXSlider := sliderWithLabel(cameraX, traceController, "Camera X: %0.1f", 0, 555)

	fieldOfView = binding.NewFloat()
	fieldOfView.Set(40)
	fieldOfViewLabel, fieldOfViewSlider := sliderWithLabel(fieldOfView, traceController, "Field of View: %0.1f", 1, 70)

	apertureSize = binding.NewFloat()
	apertureSize.Set(0)
	apertureSizeLabel, apertureSizeSlider := sliderWithLabel(apertureSize, traceController, "Aperture Size: %0.1f", 0, 20)

	lightSize = binding.NewFloat()
	lightSize.Set(300)
	lightSizeLabel, lightSizeSlider := sliderWithLabel(lightSize, traceController, "Light Size: %0.1f", 0, 300)

	lightIntensity = binding.NewFloat()
	lightIntensity.Set(10)
	lightIntensityLabel, lightIntensitySlider := sliderWithLabel(lightIntensity, traceController, "Light Intensity: %0.1f", 0, 50)

	topBar := container.New(layout.NewHBoxLayout(), &shaderSelect, &runButton, &stopButton)
	leftBar := container.New(
		layout.NewVBoxLayout(),
		cameraXLabel,
		cameraXSlider,
		fieldOfViewLabel,
		fieldOfViewSlider,
		apertureSizeLabel,
		apertureSizeSlider,
		lightSizeLabel,
		lightSizeSlider,
		lightIntensityLabel,
		lightIntensitySlider,
	)

	container := container.New(layout.NewBorderLayout(topBar, progress, leftBar, nil),
		topBar, progress, leftBar, &raster)

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

func TestScene(renderConfig renderer.RenderConfig) *renderer.Scene {

	cameraXValue, _ := cameraX.Get()
	fieldOfViewValue, _ := fieldOfView.Get()
	apertureSizeValue, _ := apertureSize.Get()
	lightSizeValue, _ := lightSize.Get()
	lightIntensityValue, _ := lightIntensity.Get()

	lookFrom := geo.NewVec3(cameraXValue, 278, -800)
	lookAt := geo.NewVec3(278, 278, 0)
	lookLength := lookFrom.Sub(lookAt).Length()

	camera := camera.New(
		renderConfig.ImageWidth,
		renderConfig.ImageHeight,
		fieldOfViewValue,
		apertureSizeValue,
		lookLength,
		lookFrom,
		lookAt,
		geo.NewVec3(0, 1, 0),
	)

	red := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(.65, .05, .05)}}
	white := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(.73, .73, .73)}}
	green := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(.12, .45, .15)}}
	light := material.DiffuseLight{Emit: material.SolidColor{ColorValue: geo.NewVec3(lightIntensityValue, lightIntensityValue, lightIntensityValue)}}
	glass := material.Dielectric{
		Tex:               material.SolidColor{ColorValue: geo.NewVec3(1, 1, .8)},
		IndexOfRefraction: 1.33,
	}
	grass := material.Lambertian{Tex: material.ImageTexture{
		Image:  *grassImage,
		Mirror: false,
	}}

	world := hittable.NewHittableList()
	world.Add(hittable.NewQuad(geo.NewVec3(555, 0, 0), geo.NewVec3(0, 555, 0), geo.NewVec3(0, 0, 555), green))
	world.Add(hittable.NewQuad(geo.NewVec3(0, 0, 0), geo.NewVec3(0, 555, 0), geo.NewVec3(0, 0, 555), red))
	world.Add(hittable.NewQuad(geo.NewVec3(278-lightSizeValue/2, 554, 278-lightSizeValue/2), geo.NewVec3(lightSizeValue, 0, 0), geo.NewVec3(0, 0, lightSizeValue), light))
	world.Add(hittable.NewQuad(geo.NewVec3(0, 0, 0), geo.NewVec3(555, 0, 0), geo.NewVec3(0, 0, 555), grass))
	world.Add(hittable.NewQuad(geo.NewVec3(555, 555, 555), geo.NewVec3(-555, 0, 0), geo.NewVec3(0, 0, -555), white))
	world.Add(hittable.NewQuad(geo.NewVec3(0, 0, 555), geo.NewVec3(555, 0, 0), geo.NewVec3(0, 555, 0), white))
	world.Add(hittable.NewSphere(geo.NewVec3(200, 265, 200), 100, glass))
	world.Add(hittable.NewSphere(geo.NewVec3(200, 265, 200), -80, glass))

	world.Add(
		hittable.NewTranslation(
			hittable.NewRotationY(
				hittable.NewBox(geo.NewVec3(0, 0, 0), geo.NewVec3(165, 330, 165), white),
				15,
			),
			geo.NewVec3(265, 0, 295),
		),
	)

	world.Add(
		hittable.NewTranslation(
			hittable.NewRotationY(
				hittable.NewBox(geo.NewVec3(0, 0, 0), geo.NewVec3(165, 165, 165), white),
				-18,
			),
			geo.NewVec3(130, 0, 65),
		),
	)

	return &renderer.Scene{
		World:           &world,
		Cam:             camera,
		BackgroundColor: geo.NewVec3(0, 0, 0),
		RenderConfig:    renderConfig,
	}

}
