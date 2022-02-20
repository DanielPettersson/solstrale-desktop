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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/DanielPettersson/solstrale"
	"github.com/DanielPettersson/solstrale/camera"
	"github.com/DanielPettersson/solstrale/geo"
	"github.com/DanielPettersson/solstrale/hittable"
	"github.com/DanielPettersson/solstrale/material"
	"github.com/DanielPettersson/solstrale/spec"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	app := app.New()
	window := app.NewWindow("Solstråle")
	window.Resize(fyne.Size{
		Width:  800,
		Height: 600,
	})

	var renderImage image.Image
	renderImage = image.NewRGBA(image.Rect(0, 0, 1, 1))

	abortRender := make(chan bool, 1)

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

	runButton.OnTapped = func() {
		runButton.Disable()
		stopButton.Enable()

		renderProgress := make(chan spec.TraceProgress, 1)

		height := int(math.Round(float64(raster.Size().Height)))
		width := int(math.Round(float64(raster.Size().Width)))

		traceSpec := spec.TraceSpecification{
			ImageWidth:      width,
			ImageHeight:     height,
			SamplesPerPixel: 1000,
			MaxDepth:        50,
			RandomSeed:      0,
		}

		scene := TestScene(traceSpec)

		go solstrale.RayTrace(scene, renderProgress, abortRender)

		go func() {
			for p := range renderProgress {
				renderImage = p.RenderImage
				progress.SetValue(p.Progress)
				raster.Refresh()
			}
			runButton.Enable()
			stopButton.Disable()
		}()
	}

	stopButton.OnTapped = func() {
		runButton.Enable()
		stopButton.Disable()
		abortRender <- true
	}

	topBar := container.New(layout.NewHBoxLayout(), &runButton, &stopButton)

	container := container.New(layout.NewBorderLayout(topBar, progress, nil, nil),
		topBar, progress, raster)

	window.SetContent(container)
	window.ShowAndRun()

	abortRender <- true
}

func TestScene(traceSpec spec.TraceSpecification) *spec.Scene {
	camera := camera.New(
		traceSpec.ImageWidth,
		traceSpec.ImageHeight,
		20,
		0.8,
		8.3,
		geo.NewVec3(-5, 3, 6),
		geo.NewVec3(.25, 1, 0),
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

	f, _ := os.Open("grass.jpg")
	defer f.Close()
	image, _, _ := image.Decode(f)
	imageTex := material.ImageTexture{
		Image:  image,
		Mirror: false,
	}

	groundMaterial := material.Lambertian{Tex: imageTex}
	checkerMat := material.Lambertian{Tex: checkerTex}
	glassMat := material.Dielectric{Tex: material.SolidColor{ColorValue: geo.NewVec3(1, 1, 1)}, IndexOfRefraction: 1.5}
	goldMat := material.Metal{Tex: noiseTex, Fuzz: .2}
	lightMat := material.DiffuseLight{Emit: material.SolidColor{ColorValue: geo.NewVec3(5, 5, 5)}}

	world.Add(hittable.NewQuad(
		geo.NewVec3(-3, 0, -7), geo.NewVec3(10, 0, 0), geo.NewVec3(0, 0, 10),
		groundMaterial,
	))
	world.Add(hittable.NewSphere(geo.NewVec3(-1, 1, 0), 1, glassMat))
	world.Add(hittable.NewRotationY(
		hittable.NewBox(geo.NewVec3(0, 0, -.5), geo.NewVec3(1, 2, .5), checkerMat),
		15,
	))
	world.Add(hittable.NewSphere(geo.NewVec3(2.1, 1, 0), 1, goldMat))
	world.Add(hittable.NewSphere(geo.NewVec3(10, 5, 10), 10, lightMat))
	world.Add(hittable.NewSphere(geo.NewVec3(-10, 5, 10), 3, lightMat))

	return &spec.Scene{
		World:           &world,
		Cam:             camera,
		BackgroundColor: geo.NewVec3(.05, .1, .2),
		Spec:            traceSpec,
	}

}
