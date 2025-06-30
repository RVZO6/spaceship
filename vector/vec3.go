package vector

// Vec3 represents a point or vector in 3D space.
type Vec3 struct {
	X, Y, Z float64
}

func (v1 Vec3) Add(v2 Vec3) Vec3 {
	return Vec3{X: v1.X + v2.X, Y: v1.Y + v2.Y, Z: v1.Z + v2.Z}
}
