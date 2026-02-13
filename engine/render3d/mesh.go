package render3d

import "math"

// Vertex3D is a vertex with position, normal, and color
type Vertex3D struct {
	Pos    Vec3
	Normal Vec3
	Color  Color3
}

// Triangle3D is three vertices
type Triangle3D struct {
	V [3]Vertex3D
}

// Mesh3D is a collection of triangles
type Mesh3D struct {
	Triangles []Triangle3D
}

func NewMesh() *Mesh3D { return &Mesh3D{} }

func (m *Mesh3D) AddTriangle(v0, v1, v2 Vertex3D) {
	m.Triangles = append(m.Triangles, Triangle3D{V: [3]Vertex3D{v0, v1, v2}})
}

func (m *Mesh3D) AddQuad(v0, v1, v2, v3 Vertex3D) {
	m.AddTriangle(v0, v1, v2)
	m.AddTriangle(v0, v2, v3)
}

func (m *Mesh3D) Transform(mat Mat4) *Mesh3D {
	out := &Mesh3D{Triangles: make([]Triangle3D, len(m.Triangles))}
	for i, tri := range m.Triangles {
		for j := 0; j < 3; j++ {
			out.Triangles[i].V[j] = tri.V[j]
			out.Triangles[i].V[j].Pos = mat.TransformPoint(tri.V[j].Pos)
			out.Triangles[i].V[j].Normal = mat.TransformDir(tri.V[j].Normal).Normalize()
		}
	}
	return out
}

func (m *Mesh3D) Append(other *Mesh3D) {
	m.Triangles = append(m.Triangles, other.Triangles...)
}

func (m *Mesh3D) SetColor(c Color3) {
	for i := range m.Triangles {
		for j := 0; j < 3; j++ {
			m.Triangles[i].V[j].Color = c
		}
	}
}

// --- Primitive generators ---

func MakeBox(w, h, d float64, c Color3) *Mesh3D {
	m := NewMesh()
	hw, hh, hd := w/2, h/2, d/2

	v := [8]Vec3{
		{-hw, -hh, -hd}, {hw, -hh, -hd}, {hw, hh, -hd}, {-hw, hh, -hd},
		{-hw, -hh, hd}, {hw, -hh, hd}, {hw, hh, hd}, {-hw, hh, hd},
	}

	faces := [][4]int{
		{0, 1, 2, 3}, // front
		{5, 4, 7, 6}, // back
		{4, 0, 3, 7}, // left
		{1, 5, 6, 2}, // right
		{3, 2, 6, 7}, // top
		{4, 5, 1, 0}, // bottom
	}
	normals := []Vec3{
		{0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}, {0, 1, 0}, {0, -1, 0},
	}

	for fi, f := range faces {
		n := normals[fi]
		shade := 0.7 + 0.3*float64(fi)/5.0
		fc := Color3{c.R * shade, c.G * shade, c.B * shade}
		v0 := Vertex3D{Pos: v[f[0]], Normal: n, Color: fc}
		v1 := Vertex3D{Pos: v[f[1]], Normal: n, Color: fc}
		v2 := Vertex3D{Pos: v[f[2]], Normal: n, Color: fc}
		v3 := Vertex3D{Pos: v[f[3]], Normal: n, Color: fc}
		m.AddQuad(v0, v1, v2, v3)
	}
	return m
}

func MakeCylinder(radius, height float64, segments int, c Color3) *Mesh3D {
	m := NewMesh()
	if segments < 6 {
		segments = 6
	}
	hh := height / 2
	top := V3(0, hh, 0)
	bot := V3(0, -hh, 0)

	for i := 0; i < segments; i++ {
		a0 := float64(i) / float64(segments) * 2 * math.Pi
		a1 := float64(i+1) / float64(segments) * 2 * math.Pi
		x0, z0 := radius*math.Cos(a0), radius*math.Sin(a0)
		x1, z1 := radius*math.Cos(a1), radius*math.Sin(a1)

		p0t := V3(x0, hh, z0)
		p1t := V3(x1, hh, z1)
		p0b := V3(x0, -hh, z0)
		p1b := V3(x1, -hh, z1)

		n0 := V3(x0, 0, z0).Normalize()
		n1 := V3(x1, 0, z1).Normalize()

		sideShade := 0.8 + 0.2*float64(i%2)
		sc := Color3{c.R * sideShade, c.G * sideShade, c.B * sideShade}

		m.AddQuad(
			Vertex3D{Pos: p0b, Normal: n0, Color: sc},
			Vertex3D{Pos: p1b, Normal: n1, Color: sc},
			Vertex3D{Pos: p1t, Normal: n1, Color: sc},
			Vertex3D{Pos: p0t, Normal: n0, Color: sc},
		)

		topN := V3(0, 1, 0)
		tc := Color3{c.R, c.G, c.B}
		m.AddTriangle(
			Vertex3D{Pos: top, Normal: topN, Color: tc},
			Vertex3D{Pos: p0t, Normal: topN, Color: tc},
			Vertex3D{Pos: p1t, Normal: topN, Color: tc},
		)

		botN := V3(0, -1, 0)
		bc := Color3{c.R * 0.6, c.G * 0.6, c.B * 0.6}
		m.AddTriangle(
			Vertex3D{Pos: bot, Normal: botN, Color: bc},
			Vertex3D{Pos: p1b, Normal: botN, Color: bc},
			Vertex3D{Pos: p0b, Normal: botN, Color: bc},
		)
	}
	return m
}

func MakeRoof(w, h, d, peakH float64, c Color3) *Mesh3D {
	m := NewMesh()
	hw, hd := w/2, d/2

	r0 := V3(-hw, h+peakH, 0)
	r1 := V3(hw, h+peakH, 0)
	e0 := V3(-hw, h, -hd)
	e1 := V3(hw, h, -hd)
	e2 := V3(hw, h, hd)
	e3 := V3(-hw, h, hd)

	fn := V3(0, hd, -peakH).Normalize()
	m.AddQuad(
		Vertex3D{Pos: e0, Normal: fn, Color: c},
		Vertex3D{Pos: e1, Normal: fn, Color: c},
		Vertex3D{Pos: r1, Normal: fn, Color: c},
		Vertex3D{Pos: r0, Normal: fn, Color: c},
	)

	bn := V3(0, hd, peakH).Normalize()
	dark := Color3{c.R * 0.75, c.G * 0.75, c.B * 0.75}
	m.AddQuad(
		Vertex3D{Pos: r0, Normal: bn, Color: dark},
		Vertex3D{Pos: r1, Normal: bn, Color: dark},
		Vertex3D{Pos: e2, Normal: bn, Color: dark},
		Vertex3D{Pos: e3, Normal: bn, Color: dark},
	)

	gn1 := V3(-1, 0, 0)
	m.AddTriangle(
		Vertex3D{Pos: e0, Normal: gn1, Color: c},
		Vertex3D{Pos: r0, Normal: gn1, Color: c},
		Vertex3D{Pos: e3, Normal: gn1, Color: c},
	)
	gn2 := V3(1, 0, 0)
	m.AddTriangle(
		Vertex3D{Pos: e1, Normal: gn2, Color: c},
		Vertex3D{Pos: e2, Normal: gn2, Color: c},
		Vertex3D{Pos: r1, Normal: gn2, Color: c},
	)
	return m
}

func MakeCone(radius, height float64, segments int, c Color3) *Mesh3D {
	m := NewMesh()
	if segments < 4 {
		segments = 4
	}
	hh := height / 2
	tip := V3(0, hh, 0)
	bot := V3(0, -hh, 0)

	for i := 0; i < segments; i++ {
		a0 := float64(i) / float64(segments) * 2 * math.Pi
		a1 := float64(i+1) / float64(segments) * 2 * math.Pi
		x0, z0 := radius*math.Cos(a0), radius*math.Sin(a0)
		x1, z1 := radius*math.Cos(a1), radius*math.Sin(a1)

		p0b := V3(x0, -hh, z0)
		p1b := V3(x1, -hh, z1)

		// Side face normal (approximate)
		slopeY := radius / height
		n0 := V3(x0, slopeY, z0).Normalize()
		n1 := V3(x1, slopeY, z1).Normalize()
		nTip := n0.Add(n1).Scale(0.5).Normalize()

		shade := 0.8 + 0.2*float64(i%2)
		sc := Color3{c.R * shade, c.G * shade, c.B * shade}

		m.AddTriangle(
			Vertex3D{Pos: p0b, Normal: n0, Color: sc},
			Vertex3D{Pos: p1b, Normal: n1, Color: sc},
			Vertex3D{Pos: tip, Normal: nTip, Color: sc},
		)

		// Bottom cap
		botN := V3(0, -1, 0)
		bc := Color3{c.R * 0.5, c.G * 0.5, c.B * 0.5}
		m.AddTriangle(
			Vertex3D{Pos: bot, Normal: botN, Color: bc},
			Vertex3D{Pos: p1b, Normal: botN, Color: bc},
			Vertex3D{Pos: p0b, Normal: botN, Color: bc},
		)
	}
	return m
}

func MakeFlatDisc(innerR, outerR, y float64, segments int, c Color3) *Mesh3D {
	m := NewMesh()
	n := V3(0, 1, 0)
	for i := 0; i < segments; i++ {
		a0 := float64(i) / float64(segments) * 2 * math.Pi
		a1 := float64(i+1) / float64(segments) * 2 * math.Pi

		ix0, iz0 := innerR*math.Cos(a0), innerR*math.Sin(a0)
		ix1, iz1 := innerR*math.Cos(a1), innerR*math.Sin(a1)
		ox0, oz0 := outerR*math.Cos(a0), outerR*math.Sin(a0)
		ox1, oz1 := outerR*math.Cos(a1), outerR*math.Sin(a1)

		m.AddQuad(
			Vertex3D{Pos: V3(ix0, y, iz0), Normal: n, Color: c},
			Vertex3D{Pos: V3(ox0, y, oz0), Normal: n, Color: c},
			Vertex3D{Pos: V3(ox1, y, oz1), Normal: n, Color: c},
			Vertex3D{Pos: V3(ix1, y, iz1), Normal: n, Color: c},
		)
	}
	return m
}
