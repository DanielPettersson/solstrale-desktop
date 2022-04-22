package scene

import (
	"github.com/DanielPettersson/solstrale/camera"
	"github.com/DanielPettersson/solstrale/geo"
	"github.com/DanielPettersson/solstrale/hittable"
	"github.com/DanielPettersson/solstrale/material"
	"github.com/DanielPettersson/solstrale/renderer"
)

func Scene() (*renderer.Scene, error) {

	world := hittable.NewHittableList()

	red := material.NewLambertian(material.NewSolidColor(1, 0, 0))
	light := material.NewLight(15, 15, 15)

	world.Add(hittable.NewSphere(geo.NewVec3(0, 0, 0), 1, red))
	world.Add(hittable.NewSphere(geo.NewVec3(3, 5, 2), 1, light))

	return &renderer.Scene{
		World: &world,
		Camera: camera.CameraConfig{
			VerticalFovDegrees: 50,
			ApertureSize:       0,
			FocusDistance:      1,
			LookFrom:           geo.NewVec3(0, 0, 4),
			LookAt:             geo.NewVec3(0, 0, 0),
		},
		BackgroundColor: geo.NewVec3(0, 0, 0),
		RenderConfig: renderer.RenderConfig{
			SamplesPerPixel: 50,
			Shader: renderer.PathTracingShader{
				MaxDepth: 50,
			},
			PostProcessor: nil,
		},
	}, nil
}
