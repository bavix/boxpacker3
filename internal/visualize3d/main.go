// nolint:*
package main

import (
	"math/rand"
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/google/uuid"

	"github.com/bavix/boxpacker3"
)

const (
	BoxTypeF       = "8ec81501-11a4-4b3f-9a52-7cd2f9c8370c"
	BoxTypeE       = "9c69baf8-1ca3-46a0-9fc2-6f15ad9fef9a"
	BoxTypeG       = "2c5279d3-48ad-451b-b673-f6d9be7fc6f6"
	BoxTypeC       = "7f1cc68f-d554-4094-8734-c68df5c13154"
	BoxTypeB       = "76cede41-86bb-4487-bfb0-9513f032d53e"
	BoxTypeA       = "8e10cebf-cee6-4136-b060-1587b993d083"
	BoxTypeStd     = "ba973206-aa64-493b-b37a-c53192cde8fd"
	BoxTypeNotStd1 = "cb1ed5b8-7405-48c5-bfd0-d86f75c99261"
	BoxTypeNotStd2 = "d91e2661-aebb-4a55-bfb5-4ff9c6e3c008"
	BoxTypeNotStd3 = "a0ecd730-375a-4313-bbe8-820710606b3d"
	BoxTypeNotStd4 = "6dff37f0-4dd1-4143-abdc-c19ab94f2e68"
	BoxTypeNotStd5 = "abac6d59-b51f-4d62-a338-42aca7afe1cc"
	BoxTypeNotStd6 = "981ffb30-a7b9-4d9e-820e-04de2145763e"
)

func NewDefaultBoxList() []*boxpacker3.Box {
	return []*boxpacker3.Box{
		boxpacker3.NewBox(BoxTypeF, 220, 185, 50, 20000),           // 0
		boxpacker3.NewBox(BoxTypeE, 165, 215, 100, 20000),          // 1
		boxpacker3.NewBox(BoxTypeG, 265, 165, 190, 20000),          // 2
		boxpacker3.NewBox(BoxTypeC, 425, 165, 190, 20000),          // 3
		boxpacker3.NewBox(BoxTypeB, 425, 265, 190, 20000),          // 4
		boxpacker3.NewBox(BoxTypeA, 425, 265, 380, 20000),          // 5
		boxpacker3.NewBox(BoxTypeStd, 530, 380, 265, 20000),        // 6
		boxpacker3.NewBox(BoxTypeNotStd1, 1000, 500, 500, 20000),   // 7
		boxpacker3.NewBox(BoxTypeNotStd2, 1000, 1000, 1000, 20000), // 8
		boxpacker3.NewBox(BoxTypeNotStd3, 2000, 500, 500, 20000),   // 9
		boxpacker3.NewBox(BoxTypeNotStd4, 2000, 2000, 2000, 20000), // 10
		boxpacker3.NewBox(BoxTypeNotStd5, 2500, 2500, 2500, 20000), // 11
		boxpacker3.NewBox(BoxTypeNotStd6, 3000, 3000, 3000, 20000), // 12
	}
}

func main() {
	a := app.App()
	scene := core.NewNode()

	cam := camera.NewPerspective(1, 1, 3000, 60, camera.Vertical)
	cam.SetPosition(0, 0, 3)
	scene.Add(cam)
	camera.NewOrbitControl(cam)

	packer := boxpacker3.NewPacker()

	boxes := NewDefaultBoxList()
	items := []*boxpacker3.Item{
		// 5
		boxpacker3.NewItem(uuid.New().String(), 100, 100, 5, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 5, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2690),

		// 35
		boxpacker3.NewItem(uuid.New().String(), 35, 100, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 35, 100, 100, 2690),
	}

	packResult, _ := packer.Pack(boxes, items)

	for _, box := range packResult.Boxes {
		if len(box.GetItems()) == 0 {
			continue
		}

		boxGeo := geometry.NewBox(float32(box.GetWidth()), float32(box.GetHeight()), float32(box.GetDepth()))
		mat := material.NewStandard(math32.NewColor("DarkGreen"))
		mat.SetOpacity(0.25)
		boxMesh := graphic.NewMesh(boxGeo, mat)
		boxMesh.SetPosition(0, 0, 0)
		scene.Add(boxMesh)

		for _, item := range box.GetItems() {
			dimension := item.GetDimension()

			itemGeo := geometry.NewBox(float32(dimension[0]), float32(dimension[1]), float32(dimension[2]))
			mat := material.NewStandard(math32.NewColorHex(uint(rand.Uint32())))
			itemMesh := graphic.NewMesh(itemGeo, mat)
			mat.SetOpacity(0.7)

			itemMesh.SetPosition(
				float32(item.GetPosition()[boxpacker3.WidthAxis])-float32(box.GetWidth())/2+float32(dimension[0])/2,
				float32(item.GetPosition()[boxpacker3.HeightAxis])-float32(box.GetHeight())/2+float32(dimension[1])/2,
				float32(item.GetPosition()[boxpacker3.DepthAxis])-float32(box.GetDepth())/2+float32(dimension[2])/2)

			scene.Add(itemMesh)
		}

		break
	}

	scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8))
	pointLight := light.NewPoint(&math32.Color{1, 1, 0}, 5.0)
	pointLight.SetPosition(1, 0, 2)
	scene.Add(pointLight)

	a.Gls().ClearColor(0.5, 0.5, 0.5, 1.0)

	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, cam)
	})
}
