package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// ==================== LIGHTING ENGINE ====================

var lightDir = vec3{-0.5, -0.7, 0.5}

type vec3 struct{ x, y, z float64 }

func (v vec3) normalize() vec3 {
	l := math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
	if l == 0 { return vec3{0, 0, 1} }
	return vec3{v.x / l, v.y / l, v.z / l}
}
func (v vec3) dot(o vec3) float64 { return v.x*o.x + v.y*o.y + v.z*o.z }

type Material struct {
	BaseColor  color.RGBA
	Roughness  float64
	Metallic   float64
	AO         float64
}

func shade(mat Material, normal vec3, ao float64) color.RGBA {
	light := lightDir.normalize()
	n := normal.normalize()
	diffuse := math.Max(0, n.dot(light))
	viewDir := vec3{0, 0, 1}
	halfDir := vec3{light.x + viewDir.x, light.y + viewDir.y, light.z + viewDir.z}.normalize()
	specPower := 2.0 + (1.0-mat.Roughness)*60.0
	specular := math.Pow(math.Max(0, n.dot(halfDir)), specPower)
	specIntensity := (1.0 - mat.Roughness) * 0.6
	if mat.Metallic > 0.5 { specIntensity *= 1.5 }
	ambient := 0.25 + 0.1*n.z
	aoFactor := ao*mat.AO + (1.0 - mat.AO)
	totalLight := (ambient + diffuse*0.65) * aoFactor
	r := clampF64(float64(mat.BaseColor.R)/255*totalLight+specular*specIntensity, 0, 1)
	g := clampF64(float64(mat.BaseColor.G)/255*totalLight+specular*specIntensity*0.95, 0, 1)
	b := clampF64(float64(mat.BaseColor.B)/255*totalLight+specular*specIntensity*0.9, 0, 1)
	return color.RGBA{cu8(r * 255), cu8(g * 255), cu8(b * 255), mat.BaseColor.A}
}

// ==================== NOISE ====================

func fbm(x, y float64, octaves int, persistence, seed float64) float64 {
	total, amp, freq, maxV := 0.0, 1.0, 1.0, 0.0
	for i := 0; i < octaves; i++ {
		total += valueNoise(x*freq+seed*17.3, y*freq+seed*31.7) * amp
		maxV += amp; amp *= persistence; freq *= 2.0
	}
	return total / maxV
}

func valueNoise(x, y float64) float64 {
	ix, iy := int(math.Floor(x)), int(math.Floor(y))
	fx, fy := x-math.Floor(x), y-math.Floor(y)
	fx = fx * fx * (3.0 - 2.0*fx); fy = fy * fy * (3.0 - 2.0*fy)
	return lerp64(lerp64(hashF(ix, iy), hashF(ix+1, iy), fx), lerp64(hashF(ix, iy+1), hashF(ix+1, iy+1), fx), fy)
}

func hashF(x, y int) float64 {
	h := x*374761393 + y*668265263
	h = (h ^ (h >> 13)) * 1274126177
	h = h ^ (h >> 16)
	return float64(h&0x7FFFFFFF) / float64(0x7FFFFFFF)
}

func lerp64(a, b, t float64) float64 { return a*(1-t) + b*t }

func worley(x, y, scale float64, seed int) float64 {
	px, py := x/scale, y/scale
	ix, iy := int(math.Floor(px)), int(math.Floor(py))
	minDist := 999.0
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			cx := float64(ix+dx) + hashF(ix+dx+seed, iy+dy+seed*3)
			cy := float64(iy+dy) + hashF(ix+dx+seed*7, iy+dy+seed*11)
			ddx, ddy := px-cx, py-cy
			d := math.Sqrt(ddx*ddx + ddy*ddy)
			if d < minDist { minDist = d }
		}
	}
	return clampF64(minDist, 0, 1)
}

// ==================== UTILITIES ====================

func cu8(v float64) uint8 {
	if v < 0 { return 0 }; if v > 255 { return 255 }; return uint8(v)
}
func clampF64(v, lo, hi float64) float64 {
	if v < lo { return lo }; if v > hi { return hi }; return v
}
func lerpColor64(a, b color.RGBA, t float64) color.RGBA {
	t = clampF64(t, 0, 1)
	return color.RGBA{
		cu8(float64(a.R)*(1-t) + float64(b.R)*t),
		cu8(float64(a.G)*(1-t) + float64(b.G)*t),
		cu8(float64(a.B)*(1-t) + float64(b.B)*t),
		cu8(float64(a.A)*(1-t) + float64(b.A)*t),
	}
}

func spxBlend(img *image.RGBA, x, y int, c color.RGBA) {
	b := img.Bounds()
	if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y || c.A == 0 { return }
	ex := img.RGBAAt(x, y)
	if ex.A == 0 { img.SetRGBA(x, y, c); return }
	a := float64(c.A) / 255.0
	img.SetRGBA(x, y, color.RGBA{
		cu8(float64(ex.R)*(1-a) + float64(c.R)*a),
		cu8(float64(ex.G)*(1-a) + float64(c.G)*a),
		cu8(float64(ex.B)*(1-a) + float64(c.B)*a),
		cu8(math.Max(float64(ex.A), float64(c.A))),
	})
}

func absI(x int) int { if x < 0 { return -x }; return x }

func drawSoftShadow(img *image.RGBA, cx, cy, rx, ry int, intensity float64) {
	for py := cy - ry*2; py <= cy+ry*2; py++ {
		for px := cx - rx*2; px <= cx+rx*2; px++ {
			dx, dy := float64(px-cx)/float64(rx), float64(py-cy)/float64(ry)
			d2 := dx*dx + dy*dy
			if d2 < 4.0 {
				spxBlend(img, px, py, color.RGBA{0, 0, 0, cu8(intensity * 255 * math.Exp(-d2*1.2))})
			}
		}
	}
}

func darken(c color.RGBA, amt float64) color.RGBA {
	f := 1.0 - clampF64(amt, 0, 0.9)
	return color.RGBA{cu8(float64(c.R) * f), cu8(float64(c.G) * f), cu8(float64(c.B) * f), c.A}
}
func brighten(c color.RGBA, amt float64) color.RGBA {
	return color.RGBA{cu8(float64(c.R) + amt*255), cu8(float64(c.G) + amt*255), cu8(float64(c.B) + amt*255), c.A}
}

// ==================== DRAWING PRIMITIVES ====================

func inDiamond(px, py, w, h int) bool {
	cx, cy := float64(w)/2, float64(h)/2
	return math.Abs(float64(px)-cx)/cx+math.Abs(float64(py)-cy)/cy <= 1.0
}

func diamondDist(px, py, w, h int) float64 {
	cx, cy := float64(w)/2, float64(h)/2
	d := math.Abs(float64(px)-cx)/cx + math.Abs(float64(py)-cy)/cy
	if d > 1 { return 0 }; return 1.0 - d
}

func lineAA(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx, dy := math.Abs(float64(x1-x0)), math.Abs(float64(y1-y0))
	steps := int(math.Max(dx, dy))
	if steps == 0 { spxBlend(img, x0, y0, c); return }
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := float64(x0) + t*float64(x1-x0); y := float64(y0) + t*float64(y1-y0)
		ix, iy := int(x), int(y); fx, fy := x-float64(ix), y-float64(iy)
		spxBlend(img, ix, iy, color.RGBA{c.R, c.G, c.B, cu8(float64(c.A) * (1 - fx) * (1 - fy))})
		spxBlend(img, ix+1, iy, color.RGBA{c.R, c.G, c.B, cu8(float64(c.A) * fx * (1 - fy))})
		spxBlend(img, ix, iy+1, color.RGBA{c.R, c.G, c.B, cu8(float64(c.A) * (1 - fx) * fy)})
		spxBlend(img, ix+1, iy+1, color.RGBA{c.R, c.G, c.B, cu8(float64(c.A) * fx * fy)})
	}
}

func thickLine(img *image.RGBA, x0, y0, x1, y1 int, thick float64, c color.RGBA) {
	dx, dy := float64(x1-x0), float64(y1-y0)
	l := math.Sqrt(dx*dx + dy*dy); if l == 0 { return }
	nx, ny := -dy/l, dx/l
	for t := -thick / 2; t <= thick/2; t += 0.5 {
		lineAA(img, x0+int(nx*t), y0+int(ny*t), x1+int(nx*t), y1+int(ny*t), c)
	}
}

func fRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for py := y; py < y+h; py++ { for px := x; px < x+w; px++ { spxBlend(img, px, py, c) } }
}

func fRectGrad(img *image.RGBA, x, y, w, h int, top, bot color.RGBA) {
	for py := y; py < y+h; py++ {
		t := float64(py-y) / float64(h)
		c := lerpColor64(top, bot, t)
		for px := x; px < x+w; px++ { spxBlend(img, px, py, c) }
	}
}

func fCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for py := cy - r - 1; py <= cy+r+1; py++ {
		for px := cx - r - 1; px <= cx+r+1; px++ {
			d := math.Sqrt(float64((px-cx)*(px-cx) + (py-cy)*(py-cy)))
			if d <= float64(r)+0.5 {
				a := c.A; if d > float64(r)-0.5 { a = cu8(float64(c.A) * (float64(r) + 0.5 - d)) }
				spxBlend(img, px, py, color.RGBA{c.R, c.G, c.B, a})
			}
		}
	}
}

func fCircleGrad(img *image.RGBA, cx, cy, r int, center, edge color.RGBA) {
	for py := cy - r - 1; py <= cy+r+1; py++ {
		for px := cx - r - 1; px <= cx+r+1; px++ {
			d := math.Sqrt(float64((px-cx)*(px-cx) + (py-cy)*(py-cy)))
			if d <= float64(r)+0.5 {
				c := lerpColor64(center, edge, d/float64(r))
				if d > float64(r)-0.5 { c.A = cu8(float64(c.A) * (float64(r) + 0.5 - d)) }
				spxBlend(img, px, py, c)
			}
		}
	}
}

func fEllipse(img *image.RGBA, cx, cy, rx, ry int, c color.RGBA) {
	for py := cy - ry - 1; py <= cy+ry+1; py++ {
		for px := cx - rx - 1; px <= cx+rx+1; px++ {
			dx, dy := float64(px-cx)/float64(rx), float64(py-cy)/float64(ry)
			d := dx*dx + dy*dy
			if d <= 1.0 {
				a := c.A; if d > 0.85 { a = cu8(float64(a) * (1.0 - d) / 0.15) }
				spxBlend(img, px, py, color.RGBA{c.R, c.G, c.B, a})
			}
		}
	}
}

func fTriangle(img *image.RGBA, x0, y0, x1, y1, x2, y2 int, c color.RGBA) {
	minX, maxX := x0, x0; minY, maxY := y0, y0
	if x1 < minX { minX = x1 }; if x2 < minX { minX = x2 }
	if x1 > maxX { maxX = x1 }; if x2 > maxX { maxX = x2 }
	if y1 < minY { minY = y1 }; if y2 < minY { minY = y2 }
	if y1 > maxY { maxY = y1 }; if y2 > maxY { maxY = y2 }
	for py := minY; py <= maxY; py++ {
		for px := minX; px <= maxX; px++ {
			d1 := float64((px-x1)*(y0-y1) - (x0-x1)*(py-y1))
			d2 := float64((px-x2)*(y1-y2) - (x1-x2)*(py-y2))
			d3 := float64((px-x0)*(y2-y0) - (x2-x0)*(py-y0))
			if !((d1 < 0 || d2 < 0 || d3 < 0) && (d1 > 0 || d2 > 0 || d3 > 0)) {
				spxBlend(img, px, py, c)
			}
		}
	}
}

func diamondOutline(img *image.RGBA, w, h int, c color.RGBA) {
	hw, hh := w/2, h/2
	lineAA(img, hw, 0, w-1, hh, c); lineAA(img, w-1, hh, hw, h-1, c)
	lineAA(img, hw, h-1, 0, hh, c); lineAA(img, 0, hh, hw, 0, c)
}

// ==================== REALISTIC ISO BOX ====================

func isoBoxReal(img *image.RGBA, cx, cy, halfW, halfH, height int, mat Material, weatherAmt float64) {
	topY := cy - height
	// LEFT FACE
	for py := topY; py <= cy; py++ {
		prog := float64(py-topY) / float64(cy-topY+1)
		edgeX := cx - int(float64(halfW)*(1.0-prog*0.5))
		for px := edgeX; px <= cx; px++ {
			t := float64(px-edgeX) / float64(cx-edgeX+1)
			normal := vec3{-0.7, 0.2, 0.5}
			ao := 1.0 - 0.3*math.Pow(1.0-prog, 3) - 0.2*math.Pow(prog, 3)
			ao *= 1.0 - 0.15*math.Pow(1.0-t, 2)
			wn := fbm(float64(px)*0.15, float64(py)*0.15, 3, 0.5, 1.0) * weatherAmt * 0.2
			lm := mat; lm.BaseColor = darken(lm.BaseColor, 0.15+wn)
			spxBlend(img, px, py, shade(lm, normal, ao))
		}
	}
	// RIGHT FACE
	for py := topY; py <= cy; py++ {
		prog := float64(py-topY) / float64(cy-topY+1)
		edgeX := cx + int(float64(halfW)*(1.0-prog*0.5))
		for px := cx; px <= edgeX; px++ {
			t := float64(px-cx) / float64(edgeX-cx+1)
			normal := vec3{0.7, 0.2, 0.5}
			ao := 1.0 - 0.25*math.Pow(1.0-prog, 3) - 0.15*math.Pow(prog, 3)
			ao *= 1.0 - 0.15*math.Pow(t, 2)
			wn := fbm(float64(px)*0.15, float64(py)*0.15, 3, 0.5, 2.0) * weatherAmt * 0.15
			lm := mat; lm.BaseColor = darken(lm.BaseColor, 0.08+wn)
			spxBlend(img, px, py, shade(lm, normal, ao))
		}
	}
	// TOP FACE
	for py := topY - halfH; py <= topY+halfH; py++ {
		for px := cx - halfW; px <= cx+halfW; px++ {
			dx := math.Abs(float64(px-cx)) / float64(halfW)
			dy := math.Abs(float64(py-topY)) / float64(halfH)
			if dx+dy <= 1.0 {
				edgeDist := 1.0 - (dx + dy)
				ao := 1.0 - 0.4*math.Pow(1.0-edgeDist, 2)
				wn := fbm(float64(px)*0.12, float64(py)*0.12, 3, 0.5, 3.0) * weatherAmt * 0.1
				lm := mat; lm.BaseColor = darken(lm.BaseColor, wn)
				spxBlend(img, px, py, shade(lm, vec3{0, -0.3, 1.0}, ao))
			}
		}
	}
}

// Weathering: rust stains, dirt
func applyWeather(img *image.RGBA, x, y, w, h int, intensity, seed float64) {
	rng := rand.New(rand.NewSource(int64(seed * 1000)))
	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			bnd := img.Bounds()
			if px < bnd.Min.X || py < bnd.Min.Y || px >= bnd.Max.X || py >= bnd.Max.Y { continue }
			t := float64(py-y) / float64(h)
			dn := fbm(float64(px)*0.1, float64(py)*0.1, 3, 0.5, seed)
			da := t * t * intensity * 0.4 * (0.5 + dn)
			if da > 0.02 {
				ex := img.RGBAAt(px, py)
				if ex.A > 0 { spxBlend(img, px, py, color.RGBA{60, 50, 35, cu8(da * 255)}) }
			}
		}
	}
	for i := 0; i < int(intensity*5); i++ {
		rx := x + rng.Intn(w); ry := y + h/3 + rng.Intn(h/3)
		for j := 0; j < 5+rng.Intn(15); j++ {
			if ry+j < y+h {
				spxBlend(img, rx+rng.Intn(3)-1, ry+j, color.RGBA{120, 60, 20, uint8(40 + rng.Intn(40))})
			}
		}
	}
}

// Realistic diamond fill with per-pixel lighting
func diamondReal(img *image.RGBA, w, h int, base color.RGBA, roughness, seed float64) {
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !inDiamond(px, py, w, h) { continue }
			dist := diamondDist(px, py, w, h)
			fx, fy := float64(px), float64(py)
			n := fbm(fx*0.08, fy*0.08, 4, 0.5, seed)
			detail := fbm(fx*0.3, fy*0.3, 2, 0.4, seed+10) * 0.3
			lx := 1.0 - float64(px)/float64(w)*0.12
			ly := 1.0 - float64(py)/float64(h)*0.1
			nx := (fbm((fx+1)*0.08, fy*0.08, 4, 0.5, seed) - n) * 5
			ny := (fbm(fx*0.08, (fy+1)*0.08, 4, 0.5, seed) - n) * 5
			normal := vec3{nx, ny, 1}.normalize()
			diffuse := math.Max(0.2, normal.dot(lightDir.normalize()))
			bright := lx * ly * diffuse * (1.0 + n*roughness + detail)
			c := color.RGBA{cu8(float64(base.R) * bright), cu8(float64(base.G) * bright), cu8(float64(base.B) * bright), 255}
			if dist < 0.04 { c.A = cu8(dist / 0.04 * 255) }
			if dist < 0.15 {
				ao := dist / 0.15
				c.R = cu8(float64(c.R) * (0.7 + 0.3*ao))
				c.G = cu8(float64(c.G) * (0.7 + 0.3*ao))
				c.B = cu8(float64(c.B) * (0.7 + 0.3*ao))
			}
			img.SetRGBA(px, py, c)
		}
	}
	diamondOutline(img, w, h, color.RGBA{0, 0, 0, 30})
}

func dirAngle(dir int) float64 { return float64(dir) * math.Pi / 4.0 }
