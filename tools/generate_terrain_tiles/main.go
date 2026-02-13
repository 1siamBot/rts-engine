package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"path/filepath"
)

func main() {
	dir := filepath.Join("assets", "ra2", "terrain")
	os.MkdirAll(dir, 0755)

	// Generate missing terrain types
	generateTiles(dir, "dirt", 6, func(x, y int, rng *rand.Rand) color.RGBA {
		base := 120.0
		noise := rng.Float64()*20 - 10
		grain := 5.0 * math.Sin(float64(x)*0.8+float64(y)*0.3)
		v := base + noise + grain
		return color.RGBA{uint8(v * 1.05), uint8(v * 0.82), uint8(v * 0.55), 255}
	})

	generateTiles(dir, "road", 4, func(x, y int, rng *rand.Rand) color.RGBA {
		base := 95.0
		noise := rng.Float64()*8 - 4
		// Asphalt texture - horizontal streaks
		streak := 3.0 * math.Sin(float64(y)*0.5+float64(x)*0.02)
		v := base + noise + streak
		return color.RGBA{uint8(v), uint8(v * 0.98), uint8(v * 0.93), 255}
	})

	generateTiles(dir, "ore", 6, func(x, y int, rng *rand.Rand) color.RGBA {
		// Golden ore on dirt
		base := 110.0
		noise := rng.Float64()*15 - 7
		v := base + noise
		// Ore nugget clusters
		cx, cy := x%16-8, y%16-8
		dist := math.Sqrt(float64(cx*cx + cy*cy))
		if dist < 5+rng.Float64()*3 {
			return color.RGBA{uint8(v * 1.5), uint8(v * 1.1), uint8(v * 0.3), 255}
		}
		return color.RGBA{uint8(v * 0.95), uint8(v * 0.75), uint8(v * 0.5), 255}
	})

	generateTiles(dir, "sand", 6, func(x, y int, rng *rand.Rand) color.RGBA {
		base := 175.0
		noise := rng.Float64()*12 - 6
		dune := 4.0 * math.Sin(float64(x)*0.15+float64(y)*0.08)
		v := base + noise + dune
		return color.RGBA{uint8(v * 1.0), uint8(v * 0.92), uint8(v * 0.65), 255}
	})

	generateTiles(dir, "concrete", 4, func(x, y int, rng *rand.Rand) color.RGBA {
		base := 140.0
		noise := rng.Float64()*6 - 3
		// Grid lines
		gridX := x % 32
		gridY := y % 32
		if gridX == 0 || gridY == 0 {
			return color.RGBA{uint8(base * 0.85), uint8(base * 0.85), uint8(base * 0.88), 255}
		}
		v := base + noise
		return color.RGBA{uint8(v * 0.95), uint8(v * 0.95), uint8(v * 0.98), 255}
	})

	generateTiles(dir, "rock", 4, func(x, y int, rng *rand.Rand) color.RGBA {
		base := 100.0
		noise := rng.Float64()*18 - 9
		crack := 8.0 * math.Sin(float64(x)*0.4+float64(y)*0.6)
		v := base + noise + crack
		return color.RGBA{uint8(v * 0.95), uint8(v * 0.92), uint8(v * 0.88), 255}
	})

	generateTiles(dir, "cliff", 4, func(x, y int, rng *rand.Rand) color.RGBA {
		base := 85.0
		noise := rng.Float64()*20 - 10
		// Vertical striations
		stria := 10.0 * math.Sin(float64(y)*0.6+float64(x)*0.1)
		v := base + noise + stria
		// Darker at top for 3D shadow
		shadow := 1.0 - float64(y)/128.0*0.3
		v *= shadow
		return color.RGBA{uint8(v * 0.9), uint8(v * 0.85), uint8(v * 0.78), 255}
	})

	generateTiles(dir, "snow", 4, func(x, y int, rng *rand.Rand) color.RGBA {
		base := 220.0
		noise := rng.Float64()*8 - 4
		sparkle := 5.0 * math.Sin(float64(x*7+y*13)*0.3)
		v := base + noise + sparkle
		if v > 255 {
			v = 255
		}
		return color.RGBA{uint8(v), uint8(v), uint8(v * 1.02), 255}
	})

	generateTiles(dir, "bridge", 4, func(x, y int, rng *rand.Rand) color.RGBA {
		base := 105.0
		noise := rng.Float64()*8 - 4
		// Wood planks
		plank := y % 8
		if plank == 0 {
			return color.RGBA{uint8(base * 0.7), uint8(base * 0.55), uint8(base * 0.3), 255}
		}
		v := base + noise
		return color.RGBA{uint8(v * 1.0), uint8(v * 0.78), uint8(v * 0.45), 255}
	})

	generateTiles(dir, "urban", 4, func(x, y int, rng *rand.Rand) color.RGBA {
		base := 130.0
		noise := rng.Float64()*10 - 5
		// Pavement pattern
		bx, by := x%16, y%16
		if bx < 1 || by < 1 {
			return color.RGBA{uint8(base * 0.75), uint8(base * 0.75), uint8(base * 0.78), 255}
		}
		v := base + noise
		return color.RGBA{uint8(v * 0.92), uint8(v * 0.90), uint8(v * 0.88), 255}
	})

	generateTiles(dir, "gem", 4, func(x, y int, rng *rand.Rand) color.RGBA {
		base := 90.0
		noise := rng.Float64()*12 - 6
		v := base + noise
		// Gem clusters (teal/cyan)
		cx, cy := x%20-10, y%20-10
		dist := math.Sqrt(float64(cx*cx + cy*cy))
		if dist < 6+rng.Float64()*3 {
			return color.RGBA{uint8(v * 0.3), uint8(v * 1.4), uint8(v * 1.4), 255}
		}
		return color.RGBA{uint8(v * 0.85), uint8(v * 0.82), uint8(v * 0.78), 255}
	})
}

func generateTiles(dir, prefix string, count int, colorFn func(x, y int, rng *rand.Rand) color.RGBA) {
	for i := 1; i <= count; i++ {
		rng := rand.New(rand.NewSource(int64(i * 12345)))
		img := image.NewRGBA(image.Rect(0, 0, 64, 64))
		for y := 0; y < 64; y++ {
			for x := 0; x < 64; x++ {
				img.SetRGBA(x, y, colorFn(x, y, rng))
			}
		}
		path := filepath.Join(dir, prefix+"_"+itoa(i)+".png")
		// Don't overwrite existing
		if _, err := os.Stat(path); err == nil {
			continue
		}
		f, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		png.Encode(f, img)
		f.Close()
	}
}

func itoa(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	return string(rune('0'+i/10)) + string(rune('0'+i%10))
}
