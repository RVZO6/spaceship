# Understanding the Spaceship Code: From Vectors to 3D Graphics

This document breaks down the Go codebase for the spaceship program, explaining everything from the ground up. We'll start with the basic building blocks of 3D math and work our way up to the full 3D rendering engine.

## Part 1: The Building Blocks - `vector/vec3.go`

Imagine you're in a giant, empty 3D space. How do you describe where something is or in which direction it's pointing? You use a **vector**. In this project, we use a `Vec3`, which is a vector with three components: X, Y, and Z.

**Analogy**: Think of a `Vec3` as a set of instructions on a treasure map. "From the starting point, go 5 steps East (X), 3 steps North (Y), and 2 steps Up (Z)." This set of instructions is a vector. It can represent a **position** (a point in space) or a **direction** (a way to travel).

```go
package vector

type Vec3 struct {
	X, Y, Z float64
}
```

This simple `struct` is the foundation of our entire 3D world. Now, let's look at what we can do with these vectors.

### Vector Operations

#### 1. `Add`: Combining Directions

If you have two sets of instructions, adding them gives you a new, combined set of instructions.

**Analogy**: If one map says "go 5 steps East and 2 steps North" and another says "go 3 steps East and 1 step North", adding them gives you a final path of "8 steps East and 3 steps North".

```go
func (v1 Vec3) Add(v2 Vec3) Vec3 {
	return Vec3{X: v1.X + v2.X, Y: v1.Y + v2.Y, Z: v1.Z + v2.Z}
}
```

**Math**:
$$
\vec{v_1} + \vec{v_2} = \begin{pmatrix} x_1 \\ y_1 \\ z_1 \end{pmatrix} + \begin{pmatrix} x_2 \\ y_2 \\ z_2 \end{pmatrix} = \begin{pmatrix} x_1 + x_2 \\ y_1 + y_2 \\ z_1 + z_2 \end{pmatrix}
$$

#### 2. `Sub`: Finding the Path Between Two Points

Subtraction tells you the direction and distance from one point to another.

**Analogy**: If you know the location of your friend (Point A) and the location of a treasure (Point B), subtracting your friend's position from the treasure's position (`B - A`) gives you the exact instructions to get from your friend to the treasure.

```go
func (v1 Vec3) Sub(v2 Vec3) Vec3 {
	return Vec3{v1.X - v2.X, v1.Y - v2.Y, v1.Z - v2.Z}
}
```

**Math**:
$$
\vec{v_1} - \vec{v_2} = \begin{pmatrix} x_1 \\ y_1 \\ z_1 \end{pmatrix} - \begin{pmatrix} x_2 \\ y_2 \\ z_2 \end{pmatrix} = \begin{pmatrix} x_1 - x_2 \\ y_1 - y_2 \\ z_1 - z_2 \end{pmatrix}
$$

#### 3. `Scale`: Making a Path Longer or Shorter

Scaling a vector means multiplying its components by a single number (a "scalar"). This changes the vector's length but not its direction.

**Analogy**: If your instructions say "go 1 step East", scaling by 10 means "go 10 steps East". You're going in the same direction, just farther.

```go
func (v Vec3) Scale(s float64) Vec3 {
	return Vec3{X: v.X * s, Y: v.Y * s, Z: v.Z * s}
}
```

**Math**:
$$
s \cdot \vec{v} = s \begin{pmatrix} x \\ y \\ z \end{pmatrix} = \begin{pmatrix} s \cdot x \\ s \cdot y \\ s \cdot z \end{pmatrix}
$$

#### 4. `Dot` Product: Measuring Alignment

The dot product is a bit more abstract. It takes two vectors and gives you a single number that tells you how much they point in the same direction.

*   If the dot product is **positive**, the vectors point in generally the same direction.
*   If it's **zero**, they are perpendicular (at a 90-degree angle).
*   If it's **negative**, they point in generally opposite directions.

**Analogy**: Imagine you're pushing a heavy box. If you push it straight forward (your force vector) and it moves straight forward (the displacement vector), you're being very effective. The dot product is high. If you push on the side of the box, it won't move forward as much. The dot product is lower. If you push on the front of the box while your friend pushes from the other side, you're working against each other. The dot product is negative.

In our code, we use this for **back-face culling**. We calculate the "normal" vector of a triangle (a vector pointing straight out from its surface) and dot it with the vector from the camera to the triangle. If the dot product is positive, it means the triangle is facing away from us, so we don't need to draw it.

```go
func (v1 Vec3) Dot(v2 Vec3) float64 {
	return v1.X*v2.X + v1.Y*v2.Y + v1.Z*v2.Z
}
```

**Math**:
$$
\vec{v_1} \cdot \vec{v_2} = x_1x_2 + y_1y_2 + z_1z_2
$$

#### 5. `Cross` Product: Finding the Perpendicular

The cross product takes two vectors and gives you a *new vector* that is perpendicular to both of the original vectors.

**Analogy**: Imagine two vectors lying flat on a table. The cross product gives you a new vector that points straight up, away from the table (or straight down, into the table). The direction follows the "right-hand rule": if you curl the fingers of your right hand from the first vector to the second, your thumb points in the direction of the cross product.

We use this to calculate the **normal vector** for back-face culling. If we have two vectors that represent two edges of a triangle, their cross product gives us a vector that points directly out of the triangle's face.

```go
func (v1 Vec3) Cross(v2 Vec3) Vec3 {
	return Vec3{
		X: v1.Y*v2.Z - v1.Z*v2.Y,
		Y: v1.Z*v2.X - v1.X*v2.Z,
		Z: v1.X*v2.Y - v1.Y*v2.X,
	}
}
```

**Math**:
$$
\vec{v_1} \times \vec{v_2} = \begin{pmatrix} y_1z_2 - z_1y_2 \\ z_1x_2 - x_1z_2 \\ x_1y_2 - y_1x_2 \end{pmatrix}
$$

## Part 2: The Transformation Machine - `vector/mat4x4.go`

A `Vec3` can represent a point in our 3D world. But how do we move it, rotate it, or make it look like it's far away? We use a **matrix**. In our case, a `Mat4x4` (a 4x4 matrix).

**Analogy**: Think of a matrix as a "transformation machine". You put a vector in one end, and a new, transformed vector comes out the other. A 4x4 matrix is special because it can perform rotation, scaling, and—crucially—translation (moving) all in one operation.

```go
type Mat4x4 struct {
	M [16]float64
}
```
We store the 16 numbers of the matrix in a simple array.

### Matrix Operations

#### 1. `NewTranslation`, `NewRotationX`, `NewRotationY`

These functions create specific "transformation machines".

*   `NewTranslation(x, y, z)`: Creates a matrix that moves (translates) a vector by a certain amount.
*   `NewRotationX(angle)`: Creates a matrix that rotates a vector around the X-axis.
*   `NewRotationY(angle)`: Creates a matrix that rotates a vector around the Y-axis.

**Analogy**: These are like blueprints for our machines. One blueprint is for a "sliding machine", another for a "spinning machine".

#### 2. `Multiply`: Combining Transformations

What if you want to rotate an object *and then* move it? You can combine two matrices by multiplying them. The result is a single new matrix that does both transformations at once.

**Analogy**: This is like building an assembly line. You connect the "spinning machine" to the "sliding machine". Now you have one big machine that spins and then slides. The order matters! Rotating then translating is different from translating then rotating.

```go
func (m1 Mat4x4) Multiply(m2 Mat4x4) Mat4x4 {
    // ... complex math ...
}
```

#### 3. `MultiplyVec3`: The Actual Transformation

This is where we feed a vector into our machine. The function takes a `Vec3` and multiplies it by the matrix, returning the new, transformed `Vec3`.

```go
func (mat Mat4x4) MultiplyVec3(v Vec3) (Vec3, float64) {
    // ... complex math ...
}
```

You might notice it also returns a `float64` called `w`. This is the fourth component of our 4D math, and it's essential for perspective. We'll get to that.

#### 4. `NewPerspective`: The Camera Lens

This is the most magical matrix. It transforms the 3D world in a way that mimics how our eyes or a camera work: objects farther away appear smaller.

**Analogy**: This matrix is the "camera lens" of our 3D engine. It takes the 3D scene and squishes and stretches it to create the illusion of depth on your flat 2D screen.

It's defined by four things:
*   `fov`: Field of View. A wider FOV is like a wide-angle lens.
*   `aspectRatio`: The ratio of the screen's width to its height.
*   `near`: The closest distance you can see.
*   `far`: The farthest distance you can see.

## Part 3: Putting It All Together - `main.go`

This is the main program where we create our 3D world, handle user input, and draw everything on the screen.

### The Game Loop

The core of any game is the **game loop**. It's a loop that runs continuously:
1.  **Process Input**: Check if the user has pressed any keys or moved the mouse.
2.  **Update State**: Update the game's state (e.g., move the player, rotate the cube).
3.  **Render**: Draw the current state of the world to the screen.
4.  **Repeat**: Do it all over again, very fast (in our case, about 60 times per second).

```go
for {
    select {
    case <-quit:
        return
    default:
        gs.Update() // Update State
        // ... Render ...
    }
}
```

### The Rendering Pipeline

This is the most complex part of `main.go`. It's the sequence of steps to take a 3D object (our cube) and figure out how to draw it on your 2D terminal screen.

Here's the journey of a single corner of our cube (a vertex) from its 3D model to a 2D pixel:

#### Step 1: Model Matrix

First, we apply the **Model Matrix**. This matrix transforms the cube from its "local" or "model" space (where its center is at `(0,0,0)`) into the main "world" space. In our code, this matrix rotates the cube.

```go
modelMatrix := vector.NewRotationY(gs.angle).Multiply(vector.NewRotationX(gs.angle))
```

#### Step 2: View Matrix

Next, we apply the **View Matrix**. This matrix does the opposite of the player's transformation. It positions the entire world so that the "camera" is at the origin `(0,0,0)` and looking down the Z-axis.

**Analogy**: Imagine you're filming a movie. Instead of moving the camera around the set, you move the entire set around the camera. The View Matrix does this. It makes the math much easier.

```go
playerRotMatrix := vector.NewRotationX(-gs.player.Pitch).Multiply(vector.NewRotationY(-gs.player.Yaw))
playerTransMatrix := vector.NewTranslation(-gs.player.Position.X, -gs.player.Position.Y, -gs.player.Position.Z)
viewMatrix := playerRotMatrix.Multiply(playerTransMatrix)
```

#### Step 3: Projection Matrix

Now we apply the **Projection Matrix**, our "camera lens". This transforms the 3D world space into a special 3D space where the perspective effect is encoded.

```go
projectionMatrix := vector.NewPerspective(fov, aspectRatio, 0.1, 100.0)
```

We combine all three matrices into one super-matrix, the **MVP Matrix** (`Model-View-Projection`).

```go
mvpMatrix := projectionMatrix.Multiply(viewMatrix.Multiply(modelMatrix))
```

This one matrix can now take a vertex from the cube's original model and transform it all the way to its final projected position!

#### Step 4: Back-Face Culling

Before we draw, we do an optimization. We calculate the triangle's normal vector (using the `Cross` product) and check which way it's facing (using the `Dot` product). If it's facing away from the camera, we don't bother drawing it.

#### Step 5: Clipping

We check if any part of the triangle is behind the camera (or too close). If it is, we "clip" it, meaning we don't draw it. This prevents weird visual glitches.

#### Step 6: Perspective Divide

This is the crucial step that creates the 3D illusion. We take the X and Y coordinates of our transformed vertex and divide them by that `w` value we got from the matrix multiplication.

```go
pv1.X /= w1
pv1.Y /= w1
```
This is what makes objects that are farther away (which have a larger `w` value) appear smaller and closer to the center of the screen.

#### Step 7: Convert to Screen Coordinates

Finally, we take the resulting 2D coordinates (which are in a range from -1 to 1) and scale them up to match the pixel coordinates of our terminal window.

```go
sx1 := int((pv1.X + 1) / 2 * float64(width))
sy1 := int((1 - pv1.Y) / 2 * float64(height))
```

And with that, we have the 2D screen position for a corner of our 3D cube. We do this for all three corners of a triangle and then use `drawLine` to connect the dots. Repeat for all visible triangles of the cube, and you have a rendered 3D object!
