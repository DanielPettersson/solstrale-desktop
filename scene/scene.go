package scene

import (
	"github.com/DanielPettersson/solstrale/camera"
	"github.com/DanielPettersson/solstrale/geo"
	"github.com/DanielPettersson/solstrale/hittable"
	"github.com/DanielPettersson/solstrale/material"
	"github.com/DanielPettersson/solstrale/renderer"
)

func Scene(width, height int) (*renderer.Scene, error) {

	world := hittable.NewHittableList()

	red := material.Lambertian{Tex: material.SolidColor{ColorValue: geo.NewVec3(1, 0, 0)}}
	light := material.DiffuseLight{
		Emit: material.SolidColor{ColorValue: geo.NewVec3(15, 15, 15)},
	}

	world.Add(hittable.NewSphere(geo.NewVec3(0, 0, 0), 1, red))
	world.Add(hittable.NewSphere(geo.NewVec3(3, 5, 2), 1, light))

	return &renderer.Scene{
		World:           &world,
		Cam:             camera.New(width, height, 50, 0, 1, geo.NewVec3(0, 0, 4), geo.NewVec3(0, 0, 0), geo.NewVec3(0, 1, 0)),
		BackgroundColor: geo.NewVec3(0, 0, 0),
		RenderConfig: renderer.RenderConfig{
			ImageWidth:      width,
			ImageHeight:     height,
			SamplesPerPixel: 50,
			Shader: renderer.PathTracingShader{
				MaxDepth: 50,
			},
			PostProcessor: nil,
		},
	}, nil
}
