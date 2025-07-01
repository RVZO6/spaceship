package main

import (
	"github.com/gdamore/tcell/v2"
	"image" // You need this for tcell.Point
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

// Forward calculates and returns a unit vector representing the direction the player is facing.
func (p *Player) Forward() vector.Vec3 {
	yaw := p.Yaw
	pitch := p.Pitch
	x := math.Sin(yaw) * math.Cos(pitch)
	y := -math.Sin(pitch)
	z := math.Cos(yaw) * math.Cos(pitch)
	return vector.Vec3{X: x, Y: y, Z: z}
}

// --- World and Physics Constants ---
const fov = 90.0                 // Field of View in degrees
const characterAspectRatio = 0.5 // Corrects for non-square terminal characters (tune to your font)
const friction = 0.99
const thrust = 0.005
const brakeForce = 0.05

// Update handles the game's physics simulation each frame.
func (gs *GameState) Update() {
	gs.angle += 0.01 // Keep the cube spinning
	gs.player.Velocity = gs.player.Velocity.Scale(friction)
	gs.player.Position = gs.player.Position.Add(gs.player.Velocity)
}

// drawLine is a utility function for drawing on the terminal screen.
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

	// --- Geometry Definition ---
	verts := []vector.Vec3{
		{X: -1, Y: -1, Z: -1}, {X: 1, Y: -1, Z: -1}, {X: 1, Y: 1, Z: -1}, {X: -1, Y: 1, Z: -1},
		{X: -1, Y: -1, Z: 1}, {X: 1, Y: -1, Z: 1}, {X: 1, Y: 1, Z: 1}, {X: -1, Y: 1, Z: 1},
	}
	edges := [][2]int{
		{0, 1}, {1, 2}, {2, 3}, {3, 0}, {4, 5}, {5, 6}, {6, 7}, {7, 4},
		{0, 4}, {1, 5}, {2, 6}, {3, 7},
	}

	// --- Game State Initialization ---
	gs := &GameState{
		vertices: verts,
		angle:    0,
		player: Player{
			// CRITICAL: Start the player outside the cube, looking towards the origin.
			Position: vector.Vec3{X: 0, Y: 0, Z: -5},
		},
	}

	// --- Input Handling Goroutine ---
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

	// --- Main Game Loop ---
	for {
		select {
		case <-quit:
			return
		default:
			gs.Update() // Run the physics simulation
			screen.Clear()

			// --- 1. Model Matrix: The cube's personal rotation ---
			modelMatrix := vector.NewRotationY(gs.angle).Multiply(vector.NewRotationX(gs.angle))

			// --- 2. View Matrix: The camera's point of view ---
			playerRotMatrix := vector.NewRotationX(-gs.player.Pitch).Multiply(vector.NewRotationY(-gs.player.Yaw))
			playerTransMatrix := vector.NewTranslation(-gs.player.Position.X, -gs.player.Position.Y, -gs.player.Position.Z)
			viewMatrix := playerRotMatrix.Multiply(playerTransMatrix) // Translate first, then rotate.

			// --- 3. Projection Matrix: Simulates the camera lens and perspective ---
			width, height := screen.Size()
			// Calculate the final aspect ratio dynamically to prevent squishing.
			screenAspectRatio := float64(width) / float64(height)
			finalAspectRatio := screenAspectRatio * characterAspectRatio
			fovRad := fov * (math.Pi / 180.0)
			projectionMatrix := vector.NewPerspective(fovRad, finalAspectRatio, 0.1, 100.0)

			// --- 4. Final MVP Matrix: Combine all transformations ---
			modelViewMatrix := viewMatrix.Multiply(modelMatrix)
			mvpMatrix := projectionMatrix.Multiply(modelViewMatrix)

			// --- 5. Process and Project all vertices ---
			projectedPoints := make([]image.Point, len(gs.vertices))
			for i, vertex := range gs.vertices {
				// A single multiplication transforms a vertex all the way to the screen's "clip space".
				clipSpaceVertex := mvpMatrix.MultiplyVec3(vertex)

				// Convert from Clip Space [-1, 1] to screen coordinates [0, width] / [0, height].
				drawX := int((clipSpaceVertex.X + 1) * 0.5 * float64(width))
				drawY := int((-clipSpaceVertex.Y + 1) * 0.5 * float64(height)) // Y is flipped for terminals

				projectedPoints[i] = image.Point{X: drawX, Y: drawY}
			}

			// --- 6. Draw the edges ---
			for _, edge := range edges {
				p1 := projectedPoints[edge[0]]
				p2 := projectedPoints[edge[1]]
				drawLine(screen, p1.X, p1.Y, p2.X, p2.Y, tcell.StyleDefault)
			}

			screen.Show()
			time.Sleep(time.Millisecond * 16) // Aim for ~60 FPS
		}
	}
}
