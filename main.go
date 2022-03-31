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
	"github.com/DanielPettersson/solstrale/post"
	"github.com/DanielPettersson/solstrale/renderer"
)

var (
	grassImage *image.Image = nil

	samplesPerPixel binding.Float = nil
	maxRayBounces   binding.Float = nil
	cameraX         binding.Float = nil
	fieldOfView     binding.Float = nil
	apertureSize    binding.Float = nil
	lightSize       binding.Float = nil
	lightIntensity  binding.Float = nil
)

func main() {
	loadTextures()

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

	shaderSelect := widget.Select{
		Options: []string{"PathTracing", "Simple", "Albedo", "Normal"},
	}
	shaderSelect.SetSelectedIndex(0)

	postProcessCheck := widget.Check{
		Text:    "Postprocess",
		Checked: true,
	}

	traceController := controller.NewTraceController(
		func() *renderer.Scene {
			height := int(math.Round(float64(raster.Size().Height)))
			width := int(math.Round(float64(raster.Size().Width)))

			var shader renderer.Shader
			shaderIdx := shaderSelect.SelectedIndex()

			if shaderIdx == 0 {
				maxDepth, _ := maxRayBounces.Get()
				shader = renderer.PathTracingShader{MaxDepth: int(maxDepth)}
			} else if shaderIdx == 1 {
				shader = renderer.SimpleShader{}
			} else if shaderIdx == 2 {
				shader = renderer.AlbedoShader{}
			} else {
				shader = renderer.NormalShader{}
			}

			samplesPerPixelVal, _ := samplesPerPixel.Get()

			var postProcessor post.PostProcessor
			if postProcessCheck.Checked {
				postProcessor = post.OidnPostProcessor{
					OidnDenoiseExecutablePath: "/home/daniel/oidnDenoise",
				}
			}

			renderConfig := renderer.RenderConfig{
				ImageWidth:      width,
				ImageHeight:     height,
				SamplesPerPixel: int(samplesPerPixelVal),
				Shader:          shader,
				PostProcessor:   postProcessor,
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

	postProcessCheck.OnChanged = func(c bool) {
		traceController.Update()
	}

	samplesPerPixel = binding.NewFloat()
	samplesPerPixel.Set(50)
	samplesPerPixelLabel, samplesPerPixelSlider := sliderWithLabel(samplesPerPixel, traceController, "Samples Per Pixel: %0.0f", 1, 1000)

	maxRayBounces = binding.NewFloat()
	maxRayBounces.Set(50)
	maxRayBouncesLabel, maxRayBouncesSlider := sliderWithLabel(maxRayBounces, traceController, "Max Ray Bounces: %0.0f", 1, 50)

	cameraX = binding.NewFloat()
	cameraX.Set(278)
	cameraXLabel, cameraXSlider := sliderWithLabel(cameraX, traceController, "Camera X: %0.1f", 0, 555)

	fieldOfView = binding.NewFloat()
	fieldOfView.Set(40)
	fieldOfViewLabel, fieldOfViewSlider := sliderWithLabel(fieldOfView, traceController, "Field of View: %0.1f", 1, 70)

	apertureSize = binding.NewFloat()
	apertureSize.Set(0)
	apertureSizeLabel, apertureSizeSlider := sliderWithLabel(apertureSize, traceController, "Aperture Size: %0.1f", 0, 200)

	lightSize = binding.NewFloat()
	lightSize.Set(300)
	lightSizeLabel, lightSizeSlider := sliderWithLabel(lightSize, traceController, "Light Size: %0.1f", 0, 300)

	lightIntensity = binding.NewFloat()
	lightIntensity.Set(10)
	lightIntensityLabel, lightIntensitySlider := sliderWithLabel(lightIntensity, traceController, "Light Intensity: %0.1f", 0, 50)

	topBar := container.New(layout.NewHBoxLayout(), &shaderSelect, &postProcessCheck, &runButton, &stopButton)
	leftBar := container.New(
		layout.NewVBoxLayout(),
		samplesPerPixelLabel,
		samplesPerPixelSlider,
		maxRayBouncesLabel,
		maxRayBouncesSlider,
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
	lookAt := geo.NewVec3(278, 278, 278)
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

	white := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(.73, .73, .73)}}
	light := material.DiffuseLight{Emit: material.SolidColor{ColorValue: geo.NewVec3(lightIntensityValue, lightIntensityValue, lightIntensityValue)}}
	grass := material.Lambertian{Tex: material.ImageTexture{
		Image:  *grassImage,
		Mirror: false,
	}}

	world := hittable.NewHittableList()
	world.Add(hittable.NewQuad(geo.NewVec3(278-lightSizeValue/2, 800, 278-lightSizeValue/2), geo.NewVec3(lightSizeValue, 0, 0), geo.NewVec3(0, 0, lightSizeValue), light))
	world.Add(hittable.NewQuad(geo.NewVec3(-750, 0, -200), geo.NewVec3(2000, 0, 0), geo.NewVec3(0, 0, 2000), grass))
	world.Add(hittable.NewConstantMedium(
		hittable.NewBox(geo.NewVec3(-1000, -1000, -1000), geo.NewVec3(2000, 2000, 2000), white),
		.0007, white.Tex,
	))

	boxes := hittable.NewHittableList()

	step := 555 / 2.
	halfStep := step / 2.
	for x := halfStep; x < 555; x += step {
		for y := halfStep; y < 555; y += step {
			for z := halfStep; z < 555; z += step {
				boxes.Add(
					hittable.NewTranslation(
						hittable.NewRotationY(
							hittable.NewBox(geo.NewVec3(0, 0, 0), geo.NewVec3(halfStep, halfStep, halfStep), white),
							(x+y+z)*0.1,
						),
						geo.NewVec3(x, y-x*.1, z),
					),
				)
			}
		}
	}
	world.Add(hittable.NewBoundingVolumeHierarchy(boxes))

	return &renderer.Scene{
		World:           &world,
		Cam:             camera,
		BackgroundColor: geo.NewVec3(0, 0, 0),
		RenderConfig:    renderConfig,
	}

}
