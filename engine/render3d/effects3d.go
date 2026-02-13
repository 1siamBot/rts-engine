package render3d

import "math"

// Particle represents a single particle in 3D space
type Particle struct {
	Pos      Vec3
	Vel      Vec3
	Color    Color3
	Alpha    float64
	Size     float64
	Life     float64
	MaxLife  float64
}

// ParticleSystem manages particles
type ParticleSystem struct {
	Particles []Particle
}

func NewParticleSystem() *ParticleSystem {
	return &ParticleSystem{}
}

// AddExplosion spawns explosion particles at a world position
func (ps *ParticleSystem) AddExplosion(wx, wz float64) {
	for i := 0; i < 20; i++ {
		angle := float64(i) / 20.0 * 2 * math.Pi
		speed := 0.5 + float64(i%5)*0.3
		ps.Particles = append(ps.Particles, Particle{
			Pos:     V3(wx, 0.3, wz),
			Vel:     V3(math.Cos(angle)*speed, 1.0+float64(i%3)*0.5, math.Sin(angle)*speed),
			Color:   Color3{1.0, 0.6 + float64(i%5)*0.08, 0.1},
			Alpha:   1.0,
			Size:    0.15 + float64(i%3)*0.05,
			Life:    0,
			MaxLife: 0.5 + float64(i%4)*0.15,
		})
	}
	// Smoke
	for i := 0; i < 8; i++ {
		angle := float64(i) / 8.0 * 2 * math.Pi
		ps.Particles = append(ps.Particles, Particle{
			Pos:     V3(wx, 0.2, wz),
			Vel:     V3(math.Cos(angle)*0.3, 0.8, math.Sin(angle)*0.3),
			Color:   Color3{0.3, 0.3, 0.3},
			Alpha:   0.7,
			Size:    0.2,
			Life:    0,
			MaxLife: 1.0 + float64(i%3)*0.3,
		})
	}
}

// AddMuzzleFlash spawns a brief muzzle flash
func (ps *ParticleSystem) AddMuzzleFlash(wx, wy, wz float64) {
	ps.Particles = append(ps.Particles, Particle{
		Pos:     V3(wx, wy, wz),
		Vel:     V3(0, 0.1, 0),
		Color:   Color3{1.0, 0.9, 0.3},
		Alpha:   1.0,
		Size:    0.2,
		Life:    0,
		MaxLife: 0.1,
	})
}

// Update advances particles
func (ps *ParticleSystem) Update(dt float64) {
	alive := ps.Particles[:0]
	for i := range ps.Particles {
		p := &ps.Particles[i]
		p.Life += dt
		if p.Life >= p.MaxLife {
			continue
		}
		// Physics
		p.Pos = p.Pos.Add(p.Vel.Scale(dt))
		p.Vel.Y -= 2.0 * dt // gravity
		// Fade
		p.Alpha = 1.0 - p.Life/p.MaxLife
		alive = append(alive, *p)
	}
	ps.Particles = alive
}

// GenerateParticleMeshes creates renderable quads for all particles (billboard approximation)
func (ps *ParticleSystem) GenerateParticleMeshes() *Mesh3D {
	mesh := NewMesh()
	up := V3(0, 1, 0)

	for _, p := range ps.Particles {
		if p.Alpha < 0.01 {
			continue
		}
		// Simple flat quad on XZ plane at particle height
		hs := p.Size / 2
		c := Color3{p.Color.R * p.Alpha, p.Color.G * p.Alpha, p.Color.B * p.Alpha}
		v0 := Vertex3D{Pos: V3(p.Pos.X-hs, p.Pos.Y, p.Pos.Z-hs), Normal: up, Color: c}
		v1 := Vertex3D{Pos: V3(p.Pos.X+hs, p.Pos.Y, p.Pos.Z-hs), Normal: up, Color: c}
		v2 := Vertex3D{Pos: V3(p.Pos.X+hs, p.Pos.Y, p.Pos.Z+hs), Normal: up, Color: c}
		v3 := Vertex3D{Pos: V3(p.Pos.X-hs, p.Pos.Y, p.Pos.Z+hs), Normal: up, Color: c}
		mesh.AddQuad(v0, v1, v2, v3)
	}
	return mesh
}
