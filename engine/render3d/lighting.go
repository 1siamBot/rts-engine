package render3d

import "math"

// DirectionalLight represents a sun-like light
type DirectionalLight struct {
	Direction Vec3   // normalized direction TO the light (from surface)
	Color     Color3 // light color
	Intensity float64
}

// AmbientLight provides fill lighting
type AmbientLight struct {
	Color     Color3
	Intensity float64
}

// LightingSetup contains the scene lighting
type LightingSetup struct {
	Sun     DirectionalLight
	Ambient AmbientLight
}

// DefaultLighting returns a nice RTS lighting setup
func DefaultLighting() LightingSetup {
	return LightingSetup{
		Sun: DirectionalLight{
			Direction: V3(-0.5, 0.8, -0.3).Normalize(), // top-left
			Color:     Color3{1.0, 0.95, 0.85},         // warm sunlight
			Intensity: 0.75,
		},
		Ambient: AmbientLight{
			Color:     Color3{0.4, 0.45, 0.6}, // cool sky ambient
			Intensity: 0.35,
		},
	}
}

// ComputeLighting calculates the lit color for a surface
func (ls *LightingSetup) ComputeLighting(normal Vec3, baseColor Color3) Color3 {
	// Ambient
	ambient := baseColor.Mul(ls.Ambient.Color).Scale(ls.Ambient.Intensity)

	// Diffuse (Lambert)
	ndotl := math.Max(0, normal.Dot(ls.Sun.Direction))
	diffuse := baseColor.Mul(ls.Sun.Color).Scale(ndotl * ls.Sun.Intensity)

	return ambient.Add(diffuse)
}

// ComputeLightingWithShadow includes a shadow factor (0=shadow, 1=lit)
func (ls *LightingSetup) ComputeLightingWithShadow(normal Vec3, baseColor Color3, shadow float64) Color3 {
	ambient := baseColor.Mul(ls.Ambient.Color).Scale(ls.Ambient.Intensity)
	ndotl := math.Max(0, normal.Dot(ls.Sun.Direction))
	diffuse := baseColor.Mul(ls.Sun.Color).Scale(ndotl * ls.Sun.Intensity * shadow)
	return ambient.Add(diffuse)
}
