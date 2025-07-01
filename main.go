package main

import (
	"github.com/gdamore/tcell/v2"
	"image" // You need this for tcell.Point
	"log"
	"spaceship/vector"
	"time"
)

type GameState struct {
	vertices []vector.Vec3
	angle    float64
}

const fov = 90.0
const aspectRatio = 0.5

func (gs *GameState) Update() {
	gs.angle += 0.02
}

func drawLine(screen tcell.Screen, x1, y1, x2, y2 int, style tcell.Style) {
	dx := x2 - x1
	if dx < 0 {
		dx = -dx
	}
	dy := y2 - y1
	if dy < 0 {
		dy = -dy
	}

	sx, sy := -1, -1
	if x1 < x2 {
		sx = 1
	}
	if y1 < y2 {
		sy = 1
	}

	err := dx - dy
	for {
		screen.SetContent(x1, y1, '#', nil, style)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("failed to create screen: %v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("failed to initialize screen: %v", err)
	}
	defer screen.Fini()

	verts := []vector.Vec3{
		{X: -1, Y: -1, Z: -1}, // 0
		{X: 1, Y: -1, Z: -1},  // 1
		{X: 1, Y: 1, Z: -1},   // 2
		{X: -1, Y: 1, Z: -1},  // 3
		{X: -1, Y: -1, Z: 1},  // 4
		{X: 1, Y: -1, Z: 1},   // 5
		{X: 1, Y: 1, Z: 1},    // 6
		{X: -1, Y: 1, Z: 1},   // 7
	}

	edges := [][2]int{
		{0, 1}, {1, 2}, {2, 3}, {3, 0}, // Bottom face
		{4, 5}, {5, 6}, {6, 7}, {7, 4}, // Top face
		{0, 4}, {1, 5}, {2, 6}, {3, 7}, // Connecting sides
	}

	gs := &GameState{
		vertices: verts,
		angle:    0,
	}

	quit := make(chan struct{})
	go func() {
		for {
			ev := screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if ev.Rune() == 'q' {
					close(quit)
					return
				}
			}
		}
	}()

	for {
		select {
		case <-quit:
			return
		default:
			gs.Update()

			screen.Clear()

			rotX := vector.NewRotationX(gs.angle * 0.7) // Rotate on X
			rotY := vector.NewRotationY(gs.angle * 1.0) // Rotate on Y
			rotZ := vector.NewRotationZ(gs.angle * 1.3) // Rotate on Z
			width, height := screen.Size()

			projectedPoints := make([]image.Point, len(gs.vertices))
			rotationMatrix := rotZ.Multiply(rotY).Multiply(rotX)

			for i, vertex := range gs.vertices {
				rotatedVertex := rotationMatrix.MultiplyVec3(vertex)
				p3d := rotatedVertex.Add(vector.Vec3{Z: 5})
				screenX := (p3d.X / p3d.Z) * fov
				screenY := (p3d.Y / p3d.Z) * fov * aspectRatio
				drawX := int(screenX) + width/2
				drawY := int(screenY) + height/2
				projectedPoints[i] = image.Point{X: drawX, Y: drawY}
			}

			for _, edge := range edges {
				p1 := projectedPoints[edge[0]]
				p2 := projectedPoints[edge[1]]
				drawLine(screen, p1.X, p1.Y, p2.X, p2.Y, tcell.StyleDefault)
			}

			screen.Show()

			time.Sleep(time.Millisecond * 16)
		}
	}
}
