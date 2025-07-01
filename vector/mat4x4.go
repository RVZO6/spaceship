package vector

import "math"

type Mat4x4 struct {
	M [16]float64
}

func (mat Mat4x4) MultiplyVec3(v Vec3) (Vec3, float64) {
	var result Vec3
	result.X = v.X*mat.M[0] + v.Y*mat.M[4] + v.Z*mat.M[8] + mat.M[12]
	result.Y = v.X*mat.M[1] + v.Y*mat.M[5] + v.Z*mat.M[9] + mat.M[13]
	result.Z = v.X*mat.M[2] + v.Y*mat.M[6] + v.Z*mat.M[10] + mat.M[14]
	w := v.X*mat.M[3] + v.Y*mat.M[7] + v.Z*mat.M[11] + mat.M[15]

	return result, w
}

func NewRotationZ(angle float64) Mat4x4 {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Mat4x4{
		M: [16]float64{
			cos, sin, 0, 0,
			-sin, cos, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		},
	}
}

func NewRotationY(angle float64) Mat4x4 {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Mat4x4{
		M: [16]float64{
			cos, 0, -sin, 0,
			0, 1, 0, 0,
			sin, 0, cos, 0,
			0, 0, 0, 1,
		},
	}
}

func NewRotationX(angle float64) Mat4x4 {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Mat4x4{
		M: [16]float64{
			1, 0, 0, 0,
			0, cos, sin, 0,
			0, -sin, cos, 0,
			0, 0, 0, 1,
		},
	}
}

func NewTranslation(x, y, z float64) Mat4x4 {
	return Mat4x4{
		M: [16]float64{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			x, y, z, 1,
		},
	}
}

func (m1 Mat4x4) Multiply(m2 Mat4x4) Mat4x4 {
	var result Mat4x4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			sum := 0.0
			for k := 0; k < 4; k++ {
				sum += m1.M[k*4+j] * m2.M[i*4+k]
			}
			result.M[i*4+j] = sum
		}
	}
	return result
}

func NewPerspective(fov, aspectRatio, near, far float64) Mat4x4 {
	f := 1.0 / math.Tan(fov*math.Pi/360.0)
	return Mat4x4{
		M: [16]float64{
			f / aspectRatio, 0, 0, 0,
			0, f, 0, 0,
			0, 0, (far + near) / (near - far), -1,
			0, 0, (2 * far * near) / (near - far), 0,
		},
	}
}

