package designs

import (
	"fmt"

	"github.com/ungerik/go3d/float64/mat4"
	"github.com/ungerik/go3d/float64/vec3"
	"github.com/ungerik/go3d/float64/vec4"
)

// DiscreteXYZ is a vector in an X-Y-Z coordinate system with integer components.
type DiscreteXYZ[Number ~int | ~uint | ~int8 | ~uint8 | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64] struct {
	X Number `yaml:"x,omitempty"`
	Y Number `yaml:"y,omitempty"`
	Z Number `yaml:"z,omitempty"`
}

// ContinuousXYZ is a vector in an X-Y-Z coordinate system with floating-point components.
type ContinuousXYZ[Number ~float32 | ~float64] struct {
	X Number `yaml:"x,omitempty"`
	Y Number `yaml:"y,omitempty"`
	Z Number `yaml:"z,omitempty"`
}

// UC2GridSpacings indicates the distance between the center of each UC2 grid space along each axis,
// in units of centimeters.
var UC2GridSpacings = ContinuousXYZ[float64]{
	X: 5,   //nolint:mnd
	Y: 5,   //nolint:mnd
	Z: 5.5, //nolint:mnd
}

const (
	DirXPos = "+x"
	DirXNeg = "-x"
	DirYPos = "+y"
	DirYNeg = "-y"
	DirZPos = "+z"
	DirZNeg = "-z"
)

// BasisVec3s holds the basis unit vectors for the various axis directions, as vec3.T's.
var BasisVec3s = map[string]vec3.T{
	DirXPos: {1, 0, 0},
	DirXNeg: {-1, 0, 0},
	DirYPos: {0, 1, 0},
	DirYNeg: {0, -1, 0},
	DirZPos: {0, 0, 1},
	DirZNeg: {0, 0, -1},
}

// BasisVec4s holds the basis unit vectors for the various axis directions, as vec4.T's.
var BasisVec4s = map[string]vec4.T{
	DirXPos: {1, 0, 0, 0},
	DirXNeg: {-1, 0, 0, 0},
	DirYPos: {0, 1, 0, 0},
	DirYNeg: {0, -1, 0, 0},
	DirZPos: {0, 0, 1, 0},
	DirZNeg: {0, 0, -1, 0},
}

// GridRotMats holds precomputed homogeneous transformation matrices for the axis-aligned rotations
// which can be specified by CompPoseRotGrid.
// The first key is the direction for Z, and the second key is the direction for X.
var GridRotMats = map[string]map[string]mat4.T{
	DirZPos: {
		DirXPos: mat4.T{
			BasisVec4s[DirXPos], // col 1
			BasisVec4s[DirYPos], // col 2
			BasisVec4s[DirZPos], // col 3
			vec4.UnitW,          // col 4
		},
		DirYPos: mat4.T{
			BasisVec4s[DirYPos],
			BasisVec4s[DirXNeg],
			BasisVec4s[DirZPos],
			vec4.UnitW,
		},
		DirXNeg: mat4.T{
			BasisVec4s[DirXNeg],
			BasisVec4s[DirYNeg],
			BasisVec4s[DirZPos],
			vec4.UnitW,
		},
		DirYNeg: mat4.T{
			BasisVec4s[DirYNeg],
			BasisVec4s[DirXPos],
			BasisVec4s[DirZPos],
			vec4.UnitW,
		},
	},
	DirZNeg: {
		DirXPos: mat4.T{
			BasisVec4s[DirXPos],
			BasisVec4s[DirYNeg],
			BasisVec4s[DirZNeg],
			vec4.UnitW,
		},
		DirYPos: mat4.T{
			BasisVec4s[DirYPos],
			BasisVec4s[DirXPos],
			BasisVec4s[DirZNeg],
			vec4.UnitW,
		},
		DirXNeg: mat4.T{
			BasisVec4s[DirXNeg],
			BasisVec4s[DirYPos],
			BasisVec4s[DirZNeg],
			vec4.UnitW,
		},
		DirYNeg: mat4.T{
			BasisVec4s[DirYNeg],
			BasisVec4s[DirXNeg],
			BasisVec4s[DirZNeg],
			vec4.UnitW,
		},
	},
	DirYPos: {
		DirXPos: mat4.T{
			BasisVec4s[DirXPos],
			BasisVec4s[DirZNeg],
			BasisVec4s[DirYPos],
			vec4.UnitW,
		},
		DirZNeg: mat4.T{
			BasisVec4s[DirZNeg],
			BasisVec4s[DirXNeg],
			BasisVec4s[DirYPos],
			vec4.UnitW,
		},
		DirXNeg: mat4.T{
			BasisVec4s[DirXNeg],
			BasisVec4s[DirZPos],
			BasisVec4s[DirYPos],
			vec4.UnitW,
		},
		DirZPos: mat4.T{
			BasisVec4s[DirZPos],
			BasisVec4s[DirXPos],
			BasisVec4s[DirYPos],
			vec4.UnitW,
		},
	},
	DirYNeg: {
		DirXPos: mat4.T{
			BasisVec4s[DirXPos],
			BasisVec4s[DirZPos],
			BasisVec4s[DirYNeg],
			vec4.UnitW,
		},
		DirZNeg: mat4.T{
			BasisVec4s[DirZNeg],
			BasisVec4s[DirXPos],
			BasisVec4s[DirYNeg],
			vec4.UnitW,
		},
		DirXNeg: mat4.T{
			BasisVec4s[DirXNeg],
			BasisVec4s[DirZNeg],
			BasisVec4s[DirYNeg],
			vec4.UnitW,
		},
		DirZPos: mat4.T{
			BasisVec4s[DirZPos],
			BasisVec4s[DirXNeg],
			BasisVec4s[DirYNeg],
			vec4.UnitW,
		},
	},
	DirXPos: {
		DirZNeg: mat4.T{
			BasisVec4s[DirZNeg],
			BasisVec4s[DirYPos],
			BasisVec4s[DirXPos],
			vec4.UnitW,
		},
		DirYPos: mat4.T{
			BasisVec4s[DirYPos],
			BasisVec4s[DirZPos],
			BasisVec4s[DirXPos],
			vec4.UnitW,
		},
		DirZPos: mat4.T{
			BasisVec4s[DirZPos],
			BasisVec4s[DirYNeg],
			BasisVec4s[DirXPos],
			vec4.UnitW,
		},
		DirYNeg: mat4.T{
			BasisVec4s[DirYNeg],
			BasisVec4s[DirZNeg],
			BasisVec4s[DirXPos],
			vec4.UnitW,
		},
	},
	DirXNeg: {
		DirZNeg: mat4.T{
			BasisVec4s[DirZNeg],
			BasisVec4s[DirYNeg],
			BasisVec4s[DirXNeg],
			vec4.UnitW,
		},
		DirYPos: mat4.T{
			BasisVec4s[DirYPos],
			BasisVec4s[DirZNeg],
			BasisVec4s[DirXNeg],
			vec4.UnitW,
		},
		DirZPos: mat4.T{
			BasisVec4s[DirZPos],
			BasisVec4s[DirYPos],
			BasisVec4s[DirXNeg],
			vec4.UnitW,
		},
		DirYNeg: mat4.T{
			BasisVec4s[DirYNeg],
			BasisVec4s[DirZPos],
			BasisVec4s[DirXNeg],
			vec4.UnitW,
		},
	},
}

// DiscreteXYZ

func (s DiscreteXYZ[Number]) String() string {
	switch {
	case s.Y == 0 && s.Z == 0:
		return fmt.Sprintf("[x%+d]", s.X)
	case s.X == 0 && s.Z == 0:
		return fmt.Sprintf("[y%+d]", s.Y)
	case s.X == 0 && s.Y == 0:
		return fmt.Sprintf("[z%+d]", s.Z)
	default:
		return fmt.Sprintf("+[%d %d %d]", s.X, s.Y, s.Z)
	}
}

func (s DiscreteXYZ[Number]) AsCM(gridSpacings ContinuousXYZ[float64]) ContinuousXYZ[float64] {
	return ContinuousXYZ[float64]{
		X: float64(s.X) * gridSpacings.X,
		Y: float64(s.Y) * gridSpacings.Y,
		Z: float64(s.Z) * gridSpacings.Z,
	}
}

func (s DiscreteXYZ[Number]) Added(t DiscreteXYZ[Number]) DiscreteXYZ[Number] {
	return DiscreteXYZ[Number]{
		X: s.X + t.X,
		Y: s.Y + t.Y,
		Z: s.Z + t.Z,
	}
}

// ContinuousXYZ

func (s ContinuousXYZ[Number]) String() string {
	switch {
	case s.Y == 0 && s.Z == 0:
		return fmt.Sprintf("[x%+.2f]", s.X)
	case s.X == 0 && s.Z == 0:
		return fmt.Sprintf("[y%+.2f]", s.Y)
	case s.X == 0 && s.Y == 0:
		return fmt.Sprintf("[z%+.2f]", s.Z)
	default:
		return fmt.Sprintf("+[%.2f %.2f %.2f]", s.X, s.Y, s.Z)
	}
}

func (s ContinuousXYZ[Number]) AsVec3() vec3.T {
	return vec3.T{float64(s.X), float64(s.Y), float64(s.Z)}
}

func (s ContinuousXYZ[Number]) Added(t ContinuousXYZ[Number]) ContinuousXYZ[Number] {
	return ContinuousXYZ[Number]{
		X: s.X + t.X,
		Y: s.Y + t.Y,
		Z: s.Z + t.Z,
	}
}
