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
	Sun      DirectionalLight
	Fill     DirectionalLight // secondary fill light
	Ambient  AmbientLight
	HasFill  bool
}

// DefaultLighting returns a balanced RTS lighting setup (not over-bright)
func DefaultLighting() LightingSetup {
	return LightingSetup{
		Sun: DirectionalLight{
			Direction: V3(-0.4, 0.85, -0.35).Normalize(),
			Color:     Color3{1.0, 0.97, 0.90},
			Intensity: 0.65,
		},
		Fill: DirectionalLight{
			Direction: V3(0.5, 0.4, 0.6).Normalize(),
			Color:     Color3{0.6, 0.7, 0.9},
			Intensity: 0.20,
		},
		Ambient: AmbientLight{
			Color:     Color3{0.65, 0.68, 0.75},
			Intensity: 0.35,
		},
		HasFill: true,
	}
}

// ComputeLighting calculates the lit color for a surface
func (ls *LightingSetup) ComputeLighting(normal Vec3, baseColor Color3) Color3 {
	// Ambient
	ambient := baseColor.Mul(ls.Ambient.Color).Scale(ls.Ambient.Intensity)

	// Diffuse (Lambert) - sun
	ndotl := math.Max(0, normal.Dot(ls.Sun.Direction))
	diffuse := baseColor.Mul(ls.Sun.Color).Scale(ndotl * ls.Sun.Intensity)

	result := ambient.Add(diffuse)

	// Fill light
	if ls.HasFill {
		ndotf := math.Max(0, normal.Dot(ls.Fill.Direction))
		fill := baseColor.Mul(ls.Fill.Color).Scale(ndotf * ls.Fill.Intensity)
		result = result.Add(fill)
	}

	// Clamp
	result.R = math.Min(result.R, 1.0)
	result.G = math.Min(result.G, 1.0)
	result.B = math.Min(result.B, 1.0)

	return result
}

// ComputeLightingWithShadow includes a shadow factor (0=shadow, 1=lit)
func (ls *LightingSetup) ComputeLightingWithShadow(normal Vec3, baseColor Color3, shadow float64) Color3 {
	ambient := baseColor.Mul(ls.Ambient.Color).Scale(ls.Ambient.Intensity)
	ndotl := math.Max(0, normal.Dot(ls.Sun.Direction))
	diffuse := baseColor.Mul(ls.Sun.Color).Scale(ndotl * ls.Sun.Intensity * shadow)
	result := ambient.Add(diffuse)

	if ls.HasFill {
		ndotf := math.Max(0, normal.Dot(ls.Fill.Direction))
		fill := baseColor.Mul(ls.Fill.Color).Scale(ndotf * ls.Fill.Intensity * (0.5 + 0.5*shadow))
		result = result.Add(fill)
	}

	result.R = math.Min(result.R, 1.0)
	result.G = math.Min(result.G, 1.0)
	result.B = math.Min(result.B, 1.0)
	return result
}
