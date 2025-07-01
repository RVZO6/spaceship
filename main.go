package main

import (
	"github.com/gdamore/tcell/v2"
	"image"
	"log"
	"math"
	"spaceship/vector"
	"time"
)

type GameState struct {
	vertices []vector.Vec3
	angle    float64
	player   Player
}

type Player struct {
	Position vector.Vec3
	Velocity vector.Vec3
	Yaw      float64
	Pitch    float64
}

func (p *Player) Forward() vector.Vec3 {
	yaw := p.Yaw
	pitch := p.Pitch
	x := math.Sin(yaw) * math.Cos(pitch)
	y := -math.Sin(pitch)
	z := math.Cos(yaw) * math.Cos(pitch)
	return vector.Vec3{X: x, Y: y, Z: z}
}

const fov = 90.0
const aspectRatio = 0.5
const friction = 0.99
const thrust = 0.005
const brakeForce = 0.05

func (gs *GameState) Update() {
	gs.angle += 0.01
	gs.player.Velocity = gs.player.Velocity.Scale(friction)
	gs.player.Position = gs.player.Position.Add(gs.player.Velocity)
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
		{X: -1, Y: -1, Z: -1}, {X: 1, Y: -1, Z: -1}, {X: 1, Y: 1, Z: -1}, {X: -1, Y: 1, Z: -1},
		{X: -1, Y: -1, Z: 1}, {X: 1, Y: -1, Z: 1}, {X: 1, Y: 1, Z: 1}, {X: -1, Y: 1, Z: 1},
	}
	edges := [][2]int{
		{0, 1}, {1, 2}, {2, 3}, {3, 0}, {4, 5}, {5, 6}, {6, 7}, {7, 4},
		{0, 4}, {1, 5}, {2, 6}, {3, 7},
	}

	gs := &GameState{
		vertices: verts,
		angle:    0,
		player:   Player{},
	}

	quit := make(chan struct{})
	go func() {
		for {
			ev := screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				rune := ev.Rune()
				if rune == 'q' {
					close(quit)
					return
				}
				if rune == 'a' {
					gs.player.Yaw -= 0.03
				} else if rune == 'd' {
					gs.player.Yaw += 0.03
				}
				if rune == 'r' {
					gs.player.Pitch -= 0.03
				} else if rune == 'f' {
					gs.player.Pitch += 0.03
				}
				if rune == 'w' {
					forwardDir := gs.player.Forward()
					thrustVector := forwardDir.Scale(thrust)
					gs.player.Velocity = gs.player.Velocity.Add(thrustVector)
				} else if rune == 's' {
					brakeVector := gs.player.Velocity.Scale(-brakeForce)
					gs.player.Velocity = gs.player.Velocity.Add(brakeVector)
				}
			}
		}
	}()

	for {
		select {
		case <-quit:
			return
		default:
			// gs.Update()
			screen.Clear()

			modelMatrix := vector.NewRotationY(gs.angle).Multiply(vector.NewRotationX(gs.angle))

			playerRotMatrix := vector.NewRotationX(-gs.player.Pitch).Multiply(vector.NewRotationY(-gs.player.Yaw))
			playerTransMatrix := vector.NewTranslation(-gs.player.Position.X, -gs.player.Position.Y, -gs.player.Position.Z)
			viewMatrix := playerRotMatrix.Multiply(playerTransMatrix)

			modelViewMatrix := viewMatrix.Multiply(modelMatrix)

			width, height := screen.Size()
			projectedPoints := make([]image.Point, len(gs.vertices))

			for i, vertex := range gs.vertices {
				rotatedVertex := modelViewMatrix.MultiplyVec3(vertex)

				// Manual projection logic
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
