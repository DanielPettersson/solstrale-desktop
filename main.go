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
			SamplesPerPixel: 100,
			MaxDepth:        50,
			RandomSeed:      rand.Int(),
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
		0.1,
		10,
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
		Scale:      .05,
	}

	f, _ := os.Open("tex.jpg")
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
	fogMat := material.Isotropic{Albedo: material.SolidColor{ColorValue: geo.NewVec3(1, 1, 1)}}
	redMat := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(1, 0, 0)}}

	world.Add(hittable.NewQuad(
		geo.NewVec3(-5, 0, -15), geo.NewVec3(20, 0, 0), geo.NewVec3(0, 0, 20),
		groundMaterial,
	))
	world.Add(hittable.NewSphere(geo.NewVec3(-1, 1, 0), 1, glassMat))
	world.Add(hittable.NewRotationY(
		hittable.NewBox(geo.NewVec3(0, 0, -.5), geo.NewVec3(1, 2, .5), checkerMat),
		15,
	))
	world.Add(hittable.NewConstantMedium(
		hittable.NewTranslation(
			hittable.NewBox(geo.NewVec3(0, 0, -.5), geo.NewVec3(1, 2, .5), fogMat),
			geo.NewVec3(0, 0, 1),
		),
		0.1,
		material.SolidColor{ColorValue: geo.NewVec3(1, 1, 1)},
	))
	world.Add(hittable.NewSphere(geo.NewVec3(2, 1, 0), 1, goldMat))
	world.Add(hittable.NewSphere(geo.NewVec3(10, 5, 10), 10, lightMat))

	world.Add(hittable.NewMotionBlur(
		hittable.NewBox(geo.NewVec3(-1, 2, 0), geo.NewVec3(-.5, 2.5, .5), redMat),
		geo.NewVec3(0, 1, 0),
	))

	balls := hittable.NewHittableList()
	for i := 0.; i < 1; i += .2 {
		for j := 0.; j < 1; j += .2 {
			for k := 0.; k < 1; k += .2 {
				balls.Add(hittable.NewSphere(geo.NewVec3(i, j+.05, k+.8), .05, redMat))
			}
		}
	}

	world.Add(hittable.NewBoundingVolumeHierarchy(balls))

	return &spec.Scene{
		World:           &world,
		Cam:             camera,
		BackgroundColor: geo.NewVec3(.2, .3, .5),
		Spec:            traceSpec,
	}

}

// FinalScene sets up a scene to ray trace
func FinalScene(traceSpec spec.TraceSpecification) *spec.Scene {
	camera := camera.New(
		traceSpec.ImageWidth,
		traceSpec.ImageHeight,
		40,
		0,
		100,
		geo.NewVec3(478, 278, -600),
		geo.NewVec3(278, 278, 0),
		geo.NewVec3(0, 1, 0),
	)

	boxes1 := hittable.NewHittableList()
	ground := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(0.48, 0.83, 0.53)}}

	boxesPerSide := 20.
	for i := .0; i < boxesPerSide; i++ {
		for j := .0; j < boxesPerSide; j++ {
			w := 100.0
			x0 := -1000 + i*w
			z0 := -1000 + j*w
			y0 := .0
			x1 := x0 + w
			y1 := rand.Float64()*100 + 1
			z1 := z0 + w

			boxes1.Add(hittable.NewBox(geo.NewVec3(x0, y0, z0), geo.NewVec3(x1, y1, z1), ground))
		}
	}

	world := hittable.NewHittableList()
	world.Add(hittable.NewBoundingVolumeHierarchy(boxes1))

	light := material.DiffuseLight{Emit: material.SolidColor{ColorValue: geo.NewVec3(7, 7, 7)}}
	world.Add(hittable.NewQuad(geo.NewVec3(123, 554, 147), geo.NewVec3(300, 0, 0), geo.NewVec3(0, 0, 265), light))

	movingSphereMaterial := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(0.7, 0.3, 0.1)}}
	world.Add(hittable.NewMotionBlur(hittable.NewSphere(geo.NewVec3(400, 400, 200), 50, movingSphereMaterial), geo.NewVec3(30, 0, 0)))

	glass := material.Dielectric{Tex: material.SolidColor{ColorValue: geo.NewVec3(1, 1, 1)}, IndexOfRefraction: 1.5}

	world.Add(hittable.NewSphere(geo.NewVec3(260, 150, 45), 50, glass))
	world.Add(hittable.NewSphere(geo.NewVec3(0, 150, 145), 50, material.Metal{Tex: material.SolidColor{ColorValue: geo.NewVec3(0.8, 0.8, 0.9)}, Fuzz: 1}))

	boundary := hittable.NewSphere(geo.NewVec3(360, 150, 145), 70, glass)
	world.Add(boundary)
	world.Add(hittable.NewConstantMedium(boundary, 0.2, material.SolidColor{ColorValue: geo.NewVec3(0.2, 0.4, 0.9)}))
	boundary = hittable.NewSphere(geo.NewVec3(0, 0, 0), 5000, glass)
	world.Add(hittable.NewConstantMedium(boundary, 0.00013, material.SolidColor{ColorValue: geo.NewVec3(1, 1, 1)}))

	// world.Add(hittable.NewSphere(geo.NewVec3(400, 200, 400), 100, material.Lambertian{imageTexture{}}))
	noiseTexture := material.NoiseTexture{ColorValue: geo.NewVec3(1, 1, 1), Scale: .1}
	world.Add(hittable.NewSphere(geo.NewVec3(220, 280, 300), 80, material.Lambertian{Tex: noiseTexture}))

	boxes2 := hittable.NewHittableList()
	white := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(0.73, 0.73, 0.73)}}
	for j := 0; j < 1000; j++ {
		boxes2.Add(hittable.NewSphere(geo.RandomVec3(0, 165), 10, white))
	}

	world.Add(hittable.NewTranslation(hittable.NewRotationY(hittable.NewBoundingVolumeHierarchy(boxes2), 15), geo.NewVec3(-100, 270, 395)))

	return &spec.Scene{
		World:           &world,
		Cam:             camera,
		BackgroundColor: geo.NewVec3(0, 0, 0),
		Spec:            traceSpec,
	}
}

// CornellBox sets up a scene to ray trace
func CornellBox(traceSpec spec.TraceSpecification) *spec.Scene {
	camera := camera.New(
		traceSpec.ImageWidth,
		traceSpec.ImageHeight,
		40,
		20,
		1070,
		geo.NewVec3(278, 278, -800),
		geo.NewVec3(278, 278, 0),
		geo.NewVec3(0, 1, 0),
	)

	red := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(.65, .05, .05)}}
	white := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(.73, .73, .73)}}
	green := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(.12, .45, .15)}}
	light := material.DiffuseLight{Emit: material.SolidColor{ColorValue: geo.NewVec3(15, 15, 15)}}

	world := hittable.NewHittableList()

	world.Add(hittable.NewQuad(geo.NewVec3(555, 0, 0), geo.NewVec3(0, 555, 0), geo.NewVec3(0, 0, 555), green))
	world.Add(hittable.NewQuad(geo.NewVec3(0, 0, 0), geo.NewVec3(0, 555, 0), geo.NewVec3(0, 0, 555), red))
	world.Add(hittable.NewQuad(geo.NewVec3(408, 554, 383), geo.NewVec3(-260, 0, 0), geo.NewVec3(0, 0, -210), light))
	world.Add(hittable.NewQuad(geo.NewVec3(0, 0, 0), geo.NewVec3(555, 0, 0), geo.NewVec3(0, 0, 555), white))
	world.Add(hittable.NewQuad(geo.NewVec3(555, 555, 555), geo.NewVec3(-555, 0, 0), geo.NewVec3(0, 0, -555), white))
	world.Add(hittable.NewQuad(geo.NewVec3(0, 0, 555), geo.NewVec3(555, 0, 0), geo.NewVec3(0, 555, 0), white))

	box1 := hittable.NewBox(geo.NewVec3(0, 0, 0), geo.NewVec3(165, 330, 165), white)
	box1 = hittable.NewRotationY(box1, 15)
	box1 = hittable.NewTranslation(box1, geo.NewVec3(265, 0, 295))
	world.Add(box1)

	box2 := hittable.NewBox(geo.NewVec3(0, 0, 0), geo.NewVec3(165, 165, 165), white)
	box2 = hittable.NewRotationY(box2, -18)
	box2 = hittable.NewTranslation(box2, geo.NewVec3(130, 0, 65))
	world.Add(box2)

	return &spec.Scene{
		World:           &world,
		Cam:             camera,
		BackgroundColor: geo.NewVec3(0, 0, 0),
		Spec:            traceSpec,
	}
}

// RandomSpheres sets up a scene to ray trace
func RandomSpheres(traceSpec spec.TraceSpecification) *spec.Scene {

	camera := camera.New(
		traceSpec.ImageWidth,
		traceSpec.ImageHeight,
		20,
		0.1,
		10,
		geo.NewVec3(13, 2, 3),
		geo.NewVec3(0, 0, 0),
		geo.NewVec3(0, 1, 0),
	)

	world := hittable.NewHittableList()

	groundMaterial := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(0.2, 0.3, 0.1)}}
	world.Add(hittable.NewSphere(geo.NewVec3(0, -1000, 0), 1000, groundMaterial))

	spheres := hittable.NewHittableList()
	for a := -7.0; a < 7; a++ {
		for b := -5.0; b < 5; b++ {
			chooseMat := rand.Float64()
			center := geo.NewVec3(a+0.9*rand.Float64(), 0.2, b+0.9*rand.Float64())

			if center.Sub(geo.NewVec3(4, 0.2, 0)).Length() > 0.9 {

				if chooseMat < 0.8 {
					material := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.RandomVec3(0, 1).Mul(geo.RandomVec3(0, 1))}}
					sphere := hittable.NewSphere(center, 0.2, material)
					blur := hittable.NewMotionBlur(sphere, geo.NewVec3(0, rand.Float64()*.5, 0))
					spheres.Add(blur)
				} else if chooseMat < 0.95 {
					material := material.Metal{Tex: material.SolidColor{ColorValue: geo.RandomVec3(0.5, 1)}, Fuzz: rand.Float64() * .5}
					spheres.Add(hittable.NewSphere(center, 0.2, material))
				} else {
					material := material.Dielectric{Tex: material.SolidColor{ColorValue: geo.NewVec3(1, 1, 1)}, IndexOfRefraction: 1.5}
					spheres.Add(hittable.NewSphere(center, 0.2, material))
				}

			}
		}
	}

	spheres.Add(hittable.NewSphere(geo.NewVec3(0, 1, 0), 1.0, material.Dielectric{Tex: material.SolidColor{ColorValue: geo.NewVec3(1, 1, 1)}, IndexOfRefraction: 1.5}))
	spheres.Add(hittable.NewSphere(geo.NewVec3(-4, 1, 0), 1, material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(0.4, 0.2, 0.1)}}))
	spheres.Add(hittable.NewSphere(geo.NewVec3(4, 1, 0), 1.0, material.Metal{Tex: material.SolidColor{ColorValue: geo.NewVec3(0.7, 0.6, 0.5)}, Fuzz: 0}))
	world.Add(hittable.NewBoundingVolumeHierarchy(spheres))

	return &spec.Scene{
		World:           &world,
		Cam:             camera,
		BackgroundColor: geo.NewVec3(0.70, 0.80, 1.00),
		Spec:            traceSpec,
	}
}
