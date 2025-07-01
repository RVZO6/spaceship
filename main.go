package main

import (
	"github.com/gdamore/tcell/v2"
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
	Position   vector.Vec3
	Velocity   vector.Vec3
	Yaw        float64
	Pitch      float64
	prevMouseX int
	prevMouseY int
	firstMouse bool
}

func (p *Player) Forward() vector.Vec3 {
	yaw := p.Yaw
	pitch := p.Pitch
	x := -math.Sin(yaw) * math.Cos(pitch)
	y := math.Sin(pitch)
	z := -math.Cos(yaw) * math.Cos(pitch)
	return vector.Vec3{X: x, Y: y, Z: z}
}

const fov = 90.0
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
	screen.EnableMouse()

	defer screen.Fini()

	verts := []vector.Vec3{
		{X: -1, Y: -1, Z: -1}, {X: 1, Y: -1, Z: -1}, {X: 1, Y: 1, Z: -1}, {X: -1, Y: 1, Z: -1},
		{X: -1, Y: -1, Z: 1}, {X: 1, Y: -1, Z: 1}, {X: 1, Y: 1, Z: 1}, {X: -1, Y: 1, Z: 1},
	}
	triangles := [][3]int{
		{0, 2, 1}, {0, 3, 2}, // Front face
		{1, 2, 6}, {1, 6, 5}, // Right face
		{0, 1, 5}, {0, 5, 4}, // Top face
		{3, 7, 6}, {3, 6, 2}, // Bottom face
		{0, 4, 7}, {0, 7, 3}, // Left face
		{4, 5, 6}, {4, 6, 7}, // Back face
	}

	gs := &GameState{
		vertices: verts,
		angle:    0,
		player:   Player{Position: vector.Vec3{X: 0, Y: 0, Z: 5}, firstMouse: true},
	}

	quit := make(chan struct{})
	go func() {
		for {
			ev := screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Rune() {
				case 'q':
					close(quit)
					return
				case 'a':
					gs.player.Yaw += 0.03
				case 'd':
					gs.player.Yaw -= 0.03
				case 'r':
					gs.player.Pitch += 0.03
				case 'f':
					gs.player.Pitch -= 0.03
				case 'w':
					forwardDir := gs.player.Forward()
					thrustVector := forwardDir.Scale(thrust)
					gs.player.Velocity = gs.player.Velocity.Add(thrustVector)
				case 's':
					brakeVector := gs.player.Velocity.Scale(-brakeForce)
					gs.player.Velocity = gs.player.Velocity.Add(brakeVector)
				}

				// Clamp pitch to avoid gimbal lock
				const maxPitch = math.Pi/2 - 0.01
				if gs.player.Pitch > maxPitch {
					gs.player.Pitch = maxPitch
				} else if gs.player.Pitch < -maxPitch {
					gs.player.Pitch = -maxPitch
				}

			case *tcell.EventMouse:
				newX, newY := ev.Position()

				if gs.player.firstMouse {
					gs.player.prevMouseX = newX
					gs.player.prevMouseY = newY
					gs.player.firstMouse = false
				} else {
					dx := float64(newX - gs.player.prevMouseX)
					dy := float64(newY - gs.player.prevMouseY)

					gs.player.Yaw -= dx * 0.05
					gs.player.Pitch -= dy * 0.05

					gs.player.prevMouseX = newX
					gs.player.prevMouseY = newY
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

			modelMatrix := vector.NewRotationY(gs.angle).Multiply(vector.NewRotationX(gs.angle))

			playerRotMatrix := vector.NewRotationX(-gs.player.Pitch).Multiply(vector.NewRotationY(-gs.player.Yaw))
			playerTransMatrix := vector.NewTranslation(-gs.player.Position.X, -gs.player.Position.Y, -gs.player.Position.Z)
			viewMatrix := playerRotMatrix.Multiply(playerTransMatrix)

			width, height := screen.Size()
			aspectRatio := float64(width) / float64(height) * 0.5
			projectionMatrix := vector.NewPerspective(fov, aspectRatio, 0.1, 100.0)
			mvpMatrix := projectionMatrix.Multiply(viewMatrix.Multiply(modelMatrix))
			modelViewMatrix := viewMatrix.Multiply(modelMatrix)

			for _, triangle := range triangles {
				v1 := gs.vertices[triangle[0]]
				v2 := gs.vertices[triangle[1]]
				v3 := gs.vertices[triangle[2]]

				// Transform vertices by the model-view matrix
				tv1, _ := modelViewMatrix.MultiplyVec3(v1)
				tv2, _ := modelViewMatrix.MultiplyVec3(v2)
				tv3, _ := modelViewMatrix.MultiplyVec3(v3)

				// Back-face culling
				normal := (tv2.Sub(tv1)).Cross(tv3.Sub(tv1))
				if normal.Dot(tv1) >= 0 {
					continue
				}

				// Project vertices
				pv1, w1 := mvpMatrix.MultiplyVec3(v1)
				pv2, w2 := mvpMatrix.MultiplyVec3(v2)
				pv3, w3 := mvpMatrix.MultiplyVec3(v3)

				// Clipping
				if w1 < 0.1 || w2 < 0.1 || w3 < 0.1 {
					continue
				}

				// Perspective divide
				pv1.X /= w1
				pv1.Y /= w1
				pv2.X /= w2
				pv2.Y /= w2
				pv3.X /= w3
				pv3.Y /= w3

				// Convert to screen coordinates
				sx1 := int((pv1.X + 1) / 2 * float64(width))
				sy1 := int((1 - pv1.Y) / 2 * float64(height))
				sx2 := int((pv2.X + 1) / 2 * float64(width))
				sy2 := int((1 - pv2.Y) / 2 * float64(height))
				sx3 := int((pv3.X + 1) / 2 * float64(width))
				sy3 := int((1 - pv3.Y) / 2 * float64(height))

				drawLine(screen, sx1, sy1, sx2, sy2, tcell.StyleDefault)
				drawLine(screen, sx2, sy2, sx3, sy3, tcell.StyleDefault)
				drawLine(screen, sx3, sy3, sx1, sy1, tcell.StyleDefault)
			}

			screen.Show()
			time.Sleep(time.Millisecond * 16)
		}
	}
}
