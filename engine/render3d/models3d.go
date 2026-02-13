package render3d

import "math"

// FactionColors - bright and obvious
var (
	AlliedBlue  = Color3{0.20, 0.45, 0.95}
	SovietRed   = Color3{0.92, 0.18, 0.18}
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

// concrete foundation color
var concreteColor = Color3{0.60, 0.58, 0.55}
var concreteDark = Color3{0.48, 0.46, 0.43}

func addFoundation(m *Mesh3D, w, d, h float64) {
	slab := MakeBox(w, 0.08, d, concreteColor)
	m.Append(slab.Transform(Mat4Translate(0, h+0.04, 0)))
	// Edge trim
	trim := MakeBox(w+0.06, 0.04, d+0.06, concreteDark)
	m.Append(trim.Transform(Mat4Translate(0, h+0.02, 0)))
}

// --- Building Models ---

func MakeConstructionYard(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Foundation slab
	addFoundation(m, 3.0, 3.0, 0)

	// Main building
	body := MakeBox(2.2, 0.7, 2.2, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.45, 0)))

	// Smaller secondary building
	body2 := MakeBox(0.9, 0.5, 0.9, Color3{fc.R * 0.8, fc.G * 0.8, fc.B * 0.8})
	m.Append(body2.Transform(Mat4Translate(-1.0, 0.35, -0.9)))

	// Crane base pillar
	craneBase := MakeBox(0.2, 0.8, 0.2, Color3{0.75, 0.72, 0.15})
	m.Append(craneBase.Transform(Mat4Translate(0.9, 1.2, 0.9)))

	// Crane arm horizontal
	arm := MakeBox(0.08, 0.08, 1.6, Color3{0.75, 0.72, 0.15})
	m.Append(arm.Transform(Mat4Translate(0.9, 1.65, 0.1)))

	// Crane cable
	cable := MakeBox(0.02, 0.45, 0.02, Color3{0.3, 0.3, 0.3})
	m.Append(cable.Transform(Mat4Translate(0.9, 1.4, -0.6)))

	// Radar dish (cylinder on a post)
	radarPost := MakeBox(0.06, 0.5, 0.06, Color3{0.5, 0.5, 0.5})
	m.Append(radarPost.Transform(Mat4Translate(-0.8, 1.35, 0.8)))
	dish := MakeCylinder(0.25, 0.06, 8, Color3{0.7, 0.7, 0.75})
	m.Append(dish.Transform(Mat4Translate(-0.8, 1.65, 0.8)))

	// Antenna
	antenna := MakeBox(0.03, 0.7, 0.03, Color3{0.65, 0.65, 0.65})
	m.Append(antenna.Transform(Mat4Translate(0.5, 1.25, -0.6)))
	// Antenna tip
	tip := MakeBox(0.06, 0.04, 0.06, Color3{1.0, 0.2, 0.2})
	m.Append(tip.Transform(Mat4Translate(0.5, 1.62, -0.6)))

	return m
}

func MakePowerPlant(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	addFoundation(m, 2.0, 2.0, 0)

	// Base building
	base := MakeBox(1.6, 0.55, 1.6, fc)
	m.Append(base.Transform(Mat4Translate(0, 0.375, 0)))

	// Cooling tower (tall cylinder)
	tower := MakeCylinder(0.42, 1.1, 14, Color3{0.75, 0.75, 0.75})
	m.Append(tower.Transform(Mat4Translate(0.3, 1.2, 0.3)))

	// Tower rim at top
	rim := MakeCylinder(0.45, 0.06, 14, Color3{0.65, 0.65, 0.65})
	m.Append(rim.Transform(Mat4Translate(0.3, 1.78, 0.3)))

	// Chimney
	chimney := MakeCylinder(0.10, 0.7, 8, Color3{0.55, 0.55, 0.55})
	m.Append(chimney.Transform(Mat4Translate(-0.55, 1.0, -0.55)))
	// Chimney top ring
	chimRim := MakeCylinder(0.13, 0.04, 8, Color3{0.45, 0.45, 0.45})
	m.Append(chimRim.Transform(Mat4Translate(-0.55, 1.37, -0.55)))

	// Warning stripes on base (small colored boxes)
	for i := 0; i < 3; i++ {
		stripe := MakeBox(0.08, 0.12, 1.6, Color3{0.9, 0.8, 0.0})
		m.Append(stripe.Transform(Mat4Translate(-0.8, 0.16+float64(i)*0.18, 0)))
	}

	// Pipes connecting buildings
	pipe := MakeCylinder(0.04, 0.8, 6, Color3{0.45, 0.45, 0.50})
	pipeMat := Mat4RotateZ(math.Pi / 2).Mul(Mat4Identity())
	pipeMat = Mat4Translate(-0.1, 0.55, -0.6).Mul(pipeMat)
	m.Append(pipe.Transform(pipeMat))

	return m
}

func MakeBarracks(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	addFoundation(m, 2.0, 2.2, 0)

	// Building body
	body := MakeBox(1.6, 0.65, 1.8, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.425, 0)))

	// Peaked roof
	roof := MakeRoof(1.8, 0.65, 2.0, 0.45, Color3{0.45, 0.28, 0.15})
	m.Append(roof)

	// Door opening (dark recess)
	door := MakeBox(0.4, 0.5, 0.06, Color3{0.15, 0.10, 0.05})
	m.Append(door.Transform(Mat4Translate(0, 0.35, -0.93)))

	// Door frame
	doorFrameL := MakeBox(0.06, 0.55, 0.08, Color3{0.5, 0.35, 0.2})
	m.Append(doorFrameL.Transform(Mat4Translate(-0.23, 0.375, -0.93)))
	doorFrameR := MakeBox(0.06, 0.55, 0.08, Color3{0.5, 0.35, 0.2})
	m.Append(doorFrameR.Transform(Mat4Translate(0.23, 0.375, -0.93)))

	// Sandbag walls (small bumps along front)
	for i := 0; i < 4; i++ {
		bx := -0.6 + float64(i)*0.35
		sandbag := MakeBox(0.3, 0.12, 0.15, Color3{0.6, 0.55, 0.4})
		m.Append(sandbag.Transform(Mat4Translate(bx, 0.14, -1.05)))
	}

	return m
}

func MakeWarFactory(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	addFoundation(m, 2.8, 2.4, 0)

	// Main building (tall)
	body := MakeBox(2.4, 1.0, 2.0, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.6, 0)))

	// Flat roof
	roof := MakeBox(2.6, 0.08, 2.2, Color3{fc.R * 0.65, fc.G * 0.65, fc.B * 0.65})
	m.Append(roof.Transform(Mat4Translate(0, 1.14, 0)))

	// Garage opening (dark)
	garage := MakeBox(1.4, 0.85, 0.06, Color3{0.08, 0.08, 0.08})
	m.Append(garage.Transform(Mat4Translate(0, 0.525, -1.03)))

	// Roller door lines (horizontal stripes)
	for i := 0; i < 4; i++ {
		line := MakeBox(1.35, 0.015, 0.07, Color3{0.25, 0.25, 0.25})
		m.Append(line.Transform(Mat4Translate(0, 0.2+float64(i)*0.2, -1.04)))
	}

	// Vehicle ramp
	ramp := MakeBox(1.6, 0.04, 0.6, Color3{0.55, 0.55, 0.50})
	m.Append(ramp.Transform(Mat4Translate(0, 0.06, -1.3)))

	// Side rails
	for _, sx := range []float64{-1.05, 1.05} {
		rail := MakeBox(0.08, 0.12, 2.4, Color3{0.65, 0.60, 0.10})
		m.Append(rail.Transform(Mat4Translate(sx, 0.14, 0)))
	}

	// Overhead crane beam
	craneBeam := MakeBox(2.4, 0.08, 0.08, Color3{0.7, 0.7, 0.2})
	m.Append(craneBeam.Transform(Mat4Translate(0, 1.2, 0)))

	return m
}

func MakeRefinery(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	addFoundation(m, 2.4, 2.8, 0)

	// Main building
	body := MakeBox(2.0, 0.55, 2.4, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.375, 0)))

	// Silo 1 (tall)
	silo1 := MakeCylinder(0.32, 1.1, 12, Color3{0.72, 0.72, 0.72})
	m.Append(silo1.Transform(Mat4Translate(0.5, 1.18, 0.6)))
	silo1Top := MakeCone(0.32, 0.2, 12, Color3{0.65, 0.65, 0.65})
	m.Append(silo1Top.Transform(Mat4Translate(0.5, 1.83, 0.6)))

	// Silo 2 (shorter)
	silo2 := MakeCylinder(0.30, 0.85, 12, Color3{0.68, 0.68, 0.68})
	m.Append(silo2.Transform(Mat4Translate(-0.5, 1.08, 0.6)))
	silo2Top := MakeCone(0.30, 0.18, 12, Color3{0.60, 0.60, 0.60})
	m.Append(silo2Top.Transform(Mat4Translate(-0.5, 1.60, 0.6)))

	// Silo 3 (small)
	silo3 := MakeCylinder(0.22, 0.6, 10, Color3{0.66, 0.66, 0.66})
	m.Append(silo3.Transform(Mat4Translate(0.0, 0.95, 0.8)))

	// Ore deposit pad
	pad := MakeBox(1.4, 0.06, 1.0, Color3{0.65, 0.55, 0.12})
	m.Append(pad.Transform(Mat4Translate(0, 0.11, -1.0)))

	// Conveyor belt
	conveyor := MakeBox(0.3, 0.04, 1.4, Color3{0.35, 0.35, 0.35})
	m.Append(conveyor.Transform(Mat4Translate(0, 0.65, -0.2)))
	// Conveyor supports
	for _, sz := range []float64{-0.6, 0.0, 0.6} {
		support := MakeBox(0.06, 0.15, 0.06, Color3{0.4, 0.4, 0.4})
		m.Append(support.Transform(Mat4Translate(0, 0.56, sz-0.2)))
	}

	// Processing pipes
	pipe := MakeCylinder(0.05, 0.8, 6, Color3{0.50, 0.50, 0.55})
	pipeMat := Mat4RotateX(math.Pi / 2)
	pipeMat = Mat4Translate(0.8, 0.7, 0.2).Mul(pipeMat)
	m.Append(pipe.Transform(pipeMat))

	return m
}

// --- Unit Models ---

func MakeTankModel(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Hull (wedge-shaped — wider at back)
	hull := MakeBox(0.6, 0.18, 0.85, Color3{fc.R * 0.7, fc.G * 0.7, fc.B * 0.7})
	m.Append(hull.Transform(Mat4Translate(0, 0.09, 0)))

	// Hull top plate (angled slightly)
	hullTop := MakeBox(0.55, 0.04, 0.8, Color3{fc.R * 0.8, fc.G * 0.8, fc.B * 0.8})
	m.Append(hullTop.Transform(Mat4Translate(0, 0.20, 0)))

	// Turret (angular, not just cylinder)
	turret := MakeBox(0.32, 0.14, 0.35, fc)
	m.Append(turret.Transform(Mat4Translate(0, 0.29, -0.02)))
	// Turret rounded top
	turretTop := MakeCylinder(0.17, 0.06, 8, fc)
	m.Append(turretTop.Transform(Mat4Translate(0, 0.39, -0.02)))

	// Main barrel
	barrel := MakeBox(0.05, 0.05, 0.55, Color3{0.30, 0.30, 0.30})
	m.Append(barrel.Transform(Mat4Translate(0, 0.32, -0.5)))

	// Muzzle brake
	muzzle := MakeBox(0.08, 0.07, 0.06, Color3{0.25, 0.25, 0.25})
	m.Append(muzzle.Transform(Mat4Translate(0, 0.32, -0.78)))

	// Tracks (left and right)
	for _, sx := range []float64{-0.32, 0.32} {
		track := MakeBox(0.10, 0.14, 0.90, Color3{0.22, 0.22, 0.22})
		m.Append(track.Transform(Mat4Translate(sx, 0.07, 0)))
		// Track guard/fender
		guard := MakeBox(0.12, 0.02, 0.88, Color3{fc.R * 0.6, fc.G * 0.6, fc.B * 0.6})
		m.Append(guard.Transform(Mat4Translate(sx, 0.15, 0)))
		// Track segment lines
		for j := 0; j < 6; j++ {
			seg := MakeBox(0.11, 0.02, 0.02, Color3{0.18, 0.18, 0.18})
			m.Append(seg.Transform(Mat4Translate(sx, 0.0, -0.38+float64(j)*0.15)))
		}
	}

	return m
}

func MakeInfantryModel(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Torso
	body := MakeBox(0.14, 0.22, 0.10, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.24, 0)))

	// Head
	head := MakeBox(0.09, 0.09, 0.09, Color3{0.82, 0.68, 0.52})
	m.Append(head.Transform(Mat4Translate(0, 0.42, 0)))

	// Helmet
	helmet := MakeBox(0.11, 0.06, 0.11, Color3{fc.R * 0.45, fc.G * 0.45, fc.B * 0.45})
	m.Append(helmet.Transform(Mat4Translate(0, 0.50, 0)))

	// Legs
	for _, sx := range []float64{-0.04, 0.04} {
		leg := MakeBox(0.05, 0.14, 0.06, Color3{0.30, 0.28, 0.25})
		m.Append(leg.Transform(Mat4Translate(sx, 0.07, 0)))
	}

	// Arms
	armR := MakeBox(0.04, 0.18, 0.05, Color3{fc.R * 0.7, fc.G * 0.7, fc.B * 0.7})
	m.Append(armR.Transform(Mat4Translate(0.10, 0.26, 0)))
	armL := MakeBox(0.04, 0.18, 0.05, Color3{fc.R * 0.7, fc.G * 0.7, fc.B * 0.7})
	m.Append(armL.Transform(Mat4Translate(-0.10, 0.26, 0)))

	// Rifle
	rifle := MakeBox(0.025, 0.025, 0.28, Color3{0.20, 0.20, 0.20})
	m.Append(rifle.Transform(Mat4Translate(0.11, 0.30, -0.08)))

	return m
}

func MakeHarvesterModel(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Main body
	body := MakeBox(0.70, 0.32, 0.90, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.16, 0)))

	// Cabin (windowed)
	cabin := MakeBox(0.48, 0.22, 0.30, Color3{0.35, 0.55, 0.75})
	m.Append(cabin.Transform(Mat4Translate(0, 0.43, -0.22)))
	// Cabin roof
	cabinRoof := MakeBox(0.52, 0.03, 0.34, Color3{fc.R * 0.8, fc.G * 0.8, fc.B * 0.8})
	m.Append(cabinRoof.Transform(Mat4Translate(0, 0.555, -0.22)))

	// Cargo bed
	cargoBed := MakeBox(0.55, 0.15, 0.45, Color3{0.45, 0.42, 0.35})
	m.Append(cargoBed.Transform(Mat4Translate(0, 0.395, 0.20)))

	// Scoop arm base (articulated)
	scoopArm := MakeBox(0.08, 0.08, 0.35, Color3{0.65, 0.62, 0.20})
	m.Append(scoopArm.Transform(Mat4Translate(0, 0.20, 0.60)))
	// Scoop bucket
	scoopBucket := MakeBox(0.40, 0.10, 0.15, Color3{0.60, 0.58, 0.18})
	m.Append(scoopBucket.Transform(Mat4Translate(0, 0.13, 0.78)))

	// Big treads
	for _, sx := range []float64{-0.35, 0.35} {
		tread := MakeBox(0.12, 0.18, 0.95, Color3{0.22, 0.22, 0.22})
		m.Append(tread.Transform(Mat4Translate(sx, 0.09, 0)))
		// Track segments
		for j := 0; j < 6; j++ {
			seg := MakeBox(0.13, 0.02, 0.02, Color3{0.18, 0.18, 0.18})
			m.Append(seg.Transform(Mat4Translate(sx, 0.0, -0.40+float64(j)*0.16)))
		}
	}

	return m
}

func MakeMCVModel(faction string) *Mesh3D {
	fc := FactionColor(faction)
	m := NewMesh()

	// Large truck body
	body := MakeBox(0.80, 0.38, 1.1, fc)
	m.Append(body.Transform(Mat4Translate(0, 0.19, 0)))

	// Cabin (truck cab)
	cabin := MakeBox(0.65, 0.30, 0.38, Color3{0.32, 0.42, 0.62})
	m.Append(cabin.Transform(Mat4Translate(0, 0.53, -0.38)))
	// Windshield
	windshield := MakeBox(0.50, 0.18, 0.03, Color3{0.5, 0.6, 0.75})
	m.Append(windshield.Transform(Mat4Translate(0, 0.55, -0.58)))

	// Deployment rig on back (large equipment box)
	equip := MakeBox(0.55, 0.25, 0.55, Color3{0.50, 0.50, 0.50})
	m.Append(equip.Transform(Mat4Translate(0, 0.50, 0.22)))

	// Satellite dish on top
	dishPost := MakeBox(0.04, 0.30, 0.04, Color3{0.60, 0.60, 0.60})
	m.Append(dishPost.Transform(Mat4Translate(0.15, 0.78, 0.22)))
	dish := MakeCylinder(0.18, 0.04, 8, Color3{0.70, 0.70, 0.75})
	m.Append(dish.Transform(Mat4Translate(0.15, 0.95, 0.22)))

	// Antenna
	antenna := MakeBox(0.02, 0.45, 0.02, Color3{0.70, 0.70, 0.70})
	m.Append(antenna.Transform(Mat4Translate(-0.20, 0.85, 0.22)))
	// Red tip
	atip := MakeBox(0.04, 0.03, 0.04, Color3{1.0, 0.15, 0.15})
	m.Append(atip.Transform(Mat4Translate(-0.20, 1.09, 0.22)))

	// Wheels
	for _, sx := range []float64{-0.38, 0.38} {
		for _, sz := range []float64{-0.40, 0.0, 0.40} {
			wheel := MakeCylinder(0.10, 0.06, 8, Color3{0.20, 0.20, 0.20})
			wheelMat := Mat4RotateZ(math.Pi / 2)
			wheelMat = Mat4Translate(sx, 0.10, sz).Mul(wheelMat)
			m.Append(wheel.Transform(wheelMat))
		}
	}

	return m
}

// RotateModelY rotates a mesh around Y axis
func RotateModelY(mesh *Mesh3D, angle float64) *Mesh3D {
	return mesh.Transform(Mat4RotateY(angle))
}
