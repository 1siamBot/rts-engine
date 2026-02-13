package render3d

import "math"

// FactionColors
var (
	AlliedBlue = Color3{0.15, 0.35, 0.85}
	SovietRed  = Color3{0.85, 0.15, 0.15}
	NeutralGray = Color3{0.6, 0.6, 0.6}
)

func FactionColor(faction string) Color3 {
	switch faction {
	case "Allied", "allied":
		return AlliedBlue
	case "Soviet", "soviet":
		return SovietRed
	default:
		return NeutralGray
	}
}

// --- Building Models ---

func MakeConstructionYard(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Main platform
	base := MakeBox(2.8, 0.3, 2.8, Color3{0.5, 0.5, 0.5})
	m.Append(base.Transform(Mat4Translate(0, 0.15, 0)))

	// Building body
	body := MakeBox(2.2, 0.8, 2.2, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.7, 0)))

	// Crane base
	craneBase := MakeBox(0.3, 0.6, 0.3, Color3{0.7, 0.7, 0.2})
	m.Append(craneBase.Transform(Mat4Translate(0.8, 1.4, 0.8)))

	// Crane arm (horizontal cylinder)
	arm := MakeBox(0.1, 0.1, 1.2, Color3{0.7, 0.7, 0.2})
	m.Append(arm.Transform(Mat4Translate(0.8, 1.8, 0.2)))

	// Crane cable
	cable := MakeBox(0.02, 0.5, 0.02, Color3{0.3, 0.3, 0.3})
	m.Append(cable.Transform(Mat4Translate(0.8, 1.5, -0.4)))

	return m
}

func MakePowerPlant(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Base building
	base := MakeBox(1.6, 0.6, 1.6, fc)
	m.Append(base.Transform(Mat4Translate(0, 0.3, 0)))

	// Cooling tower (cylinder)
	tower := MakeCylinder(0.45, 1.0, 12, Color3{0.7, 0.7, 0.7})
	m.Append(tower.Transform(Mat4Translate(0.3, 1.1, 0.3)))

	// Chimney
	chimney := MakeCylinder(0.1, 0.6, 8, Color3{0.5, 0.5, 0.5})
	m.Append(chimney.Transform(Mat4Translate(-0.5, 0.9, -0.5)))

	return m
}

func MakeBarracks(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Building body
	body := MakeBox(1.6, 0.7, 1.8, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.35, 0)))

	// Peaked roof
	roof := MakeRoof(1.8, 0.7, 2.0, 0.5, Color3{0.4, 0.25, 0.15})
	m.Append(roof)

	// Door
	door := MakeBox(0.4, 0.5, 0.05, Color3{0.3, 0.2, 0.1})
	m.Append(door.Transform(Mat4Translate(0, 0.25, -0.95)))

	return m
}

func MakeWarFactory(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Main building
	body := MakeBox(2.4, 1.0, 2.0, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.5, 0)))

	// Flat roof
	roof := MakeBox(2.6, 0.1, 2.2, Color3{fc.R * 0.7, fc.G * 0.7, fc.B * 0.7})
	m.Append(roof.Transform(Mat4Translate(0, 1.05, 0)))

	// Open front (darker area to simulate garage opening)
	garage := MakeBox(1.4, 0.8, 0.05, Color3{0.1, 0.1, 0.1})
	m.Append(garage.Transform(Mat4Translate(0, 0.4, -1.025)))

	// Side rails
	for _, sx := range []float64{-1.0, 1.0} {
		rail := MakeBox(0.1, 0.15, 2.0, Color3{0.5, 0.5, 0.1})
		m.Append(rail.Transform(Mat4Translate(sx, 0.05, 0)))
	}

	return m
}

func MakeRefinery(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Main building
	body := MakeBox(2.0, 0.6, 2.4, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.3, 0)))

	// Silo 1
	silo1 := MakeCylinder(0.35, 1.0, 10, Color3{0.7, 0.7, 0.7})
	m.Append(silo1.Transform(Mat4Translate(0.5, 1.1, 0.5)))

	// Silo 2
	silo2 := MakeCylinder(0.35, 0.8, 10, Color3{0.65, 0.65, 0.65})
	m.Append(silo2.Transform(Mat4Translate(-0.5, 1.0, 0.5)))

	// Ore deposit pad
	pad := MakeBox(1.2, 0.05, 0.8, Color3{0.6, 0.5, 0.1})
	m.Append(pad.Transform(Mat4Translate(0, 0.025, -1.0)))

	return m
}

// --- Unit Models ---

func MakeTankModel(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Hull
	hull := MakeBox(0.6, 0.2, 0.8, Color3{fc.R * 0.7, fc.G * 0.7, fc.B * 0.7})
	m.Append(hull.Transform(Mat4Translate(0, 0.1, 0)))

	// Turret
	turret := MakeCylinder(0.2, 0.15, 8, fc)
	m.Append(turret.Transform(Mat4Translate(0, 0.275, -0.05)))

	// Barrel
	barrel := MakeBox(0.06, 0.06, 0.5, Color3{0.3, 0.3, 0.3})
	m.Append(barrel.Transform(Mat4Translate(0, 0.3, -0.45)))

	// Tracks
	for _, sx := range []float64{-0.3, 0.3} {
		track := MakeBox(0.12, 0.1, 0.85, Color3{0.25, 0.25, 0.25})
		m.Append(track.Transform(Mat4Translate(sx, 0.05, 0)))
	}

	return m
}

func MakeInfantryModel(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Body (capsule approximated as box)
	body := MakeBox(0.15, 0.3, 0.12, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.25, 0)))

	// Head
	head := MakeBox(0.1, 0.1, 0.1, Color3{0.8, 0.65, 0.5})
	m.Append(head.Transform(Mat4Translate(0, 0.45, 0)))

	// Helmet
	helmet := MakeBox(0.12, 0.06, 0.12, Color3{fc.R * 0.5, fc.G * 0.5, fc.B * 0.5})
	m.Append(helmet.Transform(Mat4Translate(0, 0.52, 0)))

	// Rifle
	rifle := MakeBox(0.03, 0.03, 0.25, Color3{0.2, 0.2, 0.2})
	m.Append(rifle.Transform(Mat4Translate(0.1, 0.3, -0.05)))

	return m
}

func MakeHarvesterModel(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Body
	body := MakeBox(0.7, 0.35, 0.9, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.175, 0)))

	// Cabin
	cabin := MakeBox(0.5, 0.25, 0.3, Color3{0.3, 0.5, 0.7})
	m.Append(cabin.Transform(Mat4Translate(0, 0.475, -0.2)))

	// Scoop arm
	scoop := MakeBox(0.5, 0.08, 0.15, Color3{0.6, 0.6, 0.2})
	m.Append(scoop.Transform(Mat4Translate(0, 0.15, 0.55)))

	// Wheels
	for _, sx := range []float64{-0.3, 0.3} {
		for _, sz := range []float64{-0.35, 0.35} {
			wheel := MakeCylinder(0.1, 0.08, 6, Color3{0.2, 0.2, 0.2})
			// Rotate wheel to lie on X axis
			wheelMat := Mat4RotateZ(math.Pi / 2).Mul(Mat4Identity())
			wheelMat = Mat4Translate(sx, 0.1, sz).Mul(wheelMat)
			m.Append(wheel.Transform(wheelMat))
		}
	}

	return m
}

func MakeMCVModel(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Large truck body
	body := MakeBox(0.8, 0.4, 1.1, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.2, 0)))

	// Cabin
	cabin := MakeBox(0.65, 0.3, 0.35, Color3{0.3, 0.4, 0.6})
	m.Append(cabin.Transform(Mat4Translate(0, 0.55, -0.35)))

	// Equipment on top
	equip := MakeBox(0.5, 0.2, 0.5, Color3{0.5, 0.5, 0.5})
	m.Append(equip.Transform(Mat4Translate(0, 0.5, 0.2)))

	// Antenna
	antenna := MakeBox(0.02, 0.4, 0.02, Color3{0.7, 0.7, 0.7})
	m.Append(antenna.Transform(Mat4Translate(0.2, 0.8, 0.2)))

	return m
}

// RotateModelY rotates a mesh around Y axis
func RotateModelY(mesh *Mesh3D, angle float64) *Mesh3D {
	return mesh.Transform(Mat4RotateY(angle))
}
