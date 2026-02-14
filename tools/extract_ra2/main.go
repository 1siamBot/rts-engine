// Package main extracts RA2 sprites from .mix archives to PNG.
//
// Usage:
//
//	go run tools/extract_ra2/main.go \
//	  -input "/tmp/ra2_extract/Command & Conquer Red Alert II/ra2.mix" \
//	  -output assets/ra2/
package main

import (
	"bytes"
	"crypto/cipher"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/blowfish"
)

// ─── MIX format constants ───────────────────────────────────────────────────

const (
	flagChecksum  = 0x00010000
	flagEncrypted = 0x00020000
)

// Westwood RSA public key for RA2/TS mix files (40‑byte modulus, exponent 0x10001).
// Two 40‑byte blocks are decrypted independently then concatenated → 56‑byte Blowfish key.
var (
	rsaModulus  = hexBig("0x51bcda086d39fce4565160d651713fa2e8aa54fa6682b04aabdd0e6af8b0c1e6d1fb4f3daa437f15")
	rsaExponent = big.NewInt(0x10001)
	// Private key (needed for "encrypt" in the RSA primitive – Westwood uses raw RSA, no padding).
	// For *decryption* of mix keysource the game applies: plaintext = ciphertext^e mod n (public op).
	// BUT ccmix uses the private key with ApplyFunction which is m^d mod n. Let's check:
	// Actually Westwood "encrypts" with the private key and the game "decrypts" with the public key.
	// So to READ a mix we do: plaintext = c^e mod n. That's the public key operation.
)

func hexBig(s string) *big.Int {
	v := new(big.Int)
	v.SetString(s, 0)
	return v
}

// ─── MIX types ──────────────────────────────────────────────────────────────

type mixEntry struct {
	ID     int32
	Offset uint32
	Size   uint32
}

type mixArchive struct {
	Flags      uint32
	FileCount  uint16
	BodySize   uint32
	Entries    []mixEntry
	HeaderSize int64
	data       io.ReadSeeker // underlying reader
}

// ─── CRC-based file ID (for TS/RA2 "new mix" format) ───────────────────────

var crcTable = makeCRCTable()

func makeCRCTable() [256]uint32 {
	var t [256]uint32
	for i := 0; i < 256; i++ {
		c := uint32(i)
		for j := 0; j < 8; j++ {
			if c&1 != 0 {
				c = 0xedb88320 ^ (c >> 1)
			} else {
				c >>= 1
			}
		}
		t[i] = c
	}
	return t
}

func mixID(name string) int32 {
	upper := strings.ToUpper(name)
	fname := upper
	l := len(fname)
	a := l >> 2
	if l&3 != 0 {
		fname += string(rune(l - (a << 2)))
		pad := 3 - (l & 3)
		for i := 0; i < pad; i++ {
			fname += string(fname[a<<2])
		}
	}
	data := []byte(fname)
	var rv uint32 = 0xffffffff
	for _, b := range data {
		rv = (rv >> 8) ^ crcTable[b^byte(rv&0xff)]
	}
	rv = ^rv
	return int32(rv)
}

// ─── Read MIX archive ──────────────────────────────────────────────────────

func readMix(r io.ReadSeeker) (*mixArchive, error) {
	m := &mixArchive{data: r}

	// Read first 4 bytes to check for flags or old format
	var first4 [4]byte
	if _, err := io.ReadFull(r, first4[:]); err != nil {
		return nil, err
	}
	flags := binary.LittleEndian.Uint32(first4[:])

	// Check if this is old format (first 2 bytes non-zero = file count)
	if first4[0] != 0 || first4[1] != 0 {
		// Old format: no flags, first 2 bytes = file count, next 4 = body size
		r.Seek(0, io.SeekStart)
		binary.Read(r, binary.LittleEndian, &m.FileCount)
		binary.Read(r, binary.LittleEndian, &m.BodySize)
		m.HeaderSize = 6
		return readMixIndex(r, m)
	}

	m.Flags = flags

	if flags&flagEncrypted != 0 {
		return readEncryptedMix(r, m)
	}
	// Unencrypted new format
	binary.Read(r, binary.LittleEndian, &m.FileCount)
	binary.Read(r, binary.LittleEndian, &m.BodySize)
	m.HeaderSize = 4 + 2 + 4
	return readMixIndex(r, m)
}

func readMixIndex(r io.ReadSeeker, m *mixArchive) (*mixArchive, error) {
	m.Entries = make([]mixEntry, m.FileCount)
	for i := range m.Entries {
		binary.Read(r, binary.LittleEndian, &m.Entries[i].ID)
		binary.Read(r, binary.LittleEndian, &m.Entries[i].Offset)
		binary.Read(r, binary.LittleEndian, &m.Entries[i].Size)
	}
	m.HeaderSize += int64(m.FileCount) * 12
	// round up for encrypted header padding
	return m, nil
}

func readEncryptedMix(r io.ReadSeeker, m *mixArchive) (*mixArchive, error) {
	// Read 80-byte RSA-encrypted key source
	var keysource [80]byte
	if _, err := io.ReadFull(r, keysource[:]); err != nil {
		return nil, fmt.Errorf("read keysource: %w", err)
	}

	// Decrypt keysource: split into two 40-byte blocks, reverse endianness, RSA decrypt
	bfKey := decryptKeySource(keysource[:])

	// Set up Blowfish ECB decryption
	bf, err := blowfish.NewCipher(bfKey)
	if err != nil {
		return nil, fmt.Errorf("blowfish init: %w", err)
	}

	// Read and decrypt first 8 bytes to get file_count and body_size
	var block [8]byte
	if _, err := io.ReadFull(r, block[:]); err != nil {
		return nil, err
	}
	decryptECB(bf, block[:])

	m.FileCount = binary.LittleEndian.Uint16(block[0:2])
	m.BodySize = binary.LittleEndian.Uint32(block[2:6])

	// Calculate remaining blocks needed for index
	indexBytes := int(m.FileCount)*12 - 2 // we already have 2 bytes from first block
	blockCount := indexBytes / 8
	if indexBytes%8 != 0 {
		blockCount++
	}

	// Build index buffer: first 2 bytes from block[6:8], then decrypt remaining blocks
	indexBuf := make([]byte, blockCount*8+2)
	copy(indexBuf[0:2], block[6:8])

	for i := 0; i < blockCount; i++ {
		if _, err := io.ReadFull(r, block[:]); err != nil {
			return nil, err
		}
		decryptECB(bf, block[:])
		copy(indexBuf[2+i*8:2+(i+1)*8], block[:])
	}

	// Parse index entries
	m.Entries = make([]mixEntry, m.FileCount)
	for i := range m.Entries {
		off := i * 12
		m.Entries[i].ID = int32(binary.LittleEndian.Uint32(indexBuf[off : off+4]))
		m.Entries[i].Offset = binary.LittleEndian.Uint32(indexBuf[off+4 : off+8])
		m.Entries[i].Size = binary.LittleEndian.Uint32(indexBuf[off+8 : off+12])
	}

	// Header size = 4 (flags) + 80 (keysource) + (blockCount+1)*8
	m.HeaderSize = 4 + 80 + int64(blockCount+1)*8

	return m, nil
}

func decryptKeySource(keysource []byte) []byte {
	// Reverse the 80 bytes (little-endian to big-endian for crypto)
	reversed := make([]byte, 80)
	for i := 0; i < 80; i++ {
		reversed[79-i] = keysource[i]
	}

	// Split into two 40-byte blocks
	block1 := new(big.Int).SetBytes(reversed[0:40])
	block2 := new(big.Int).SetBytes(reversed[40:80])

	// RSA public key operation: plaintext = ciphertext^e mod n
	plain1 := new(big.Int).Exp(block1, rsaExponent, rsaModulus)
	plain2 := new(big.Int).Exp(block2, rsaExponent, rsaModulus)

	// Combine: bfKey = (plain1 << 312) + plain2, then encode to 56 bytes
	combined := new(big.Int).Lsh(plain1, 312)
	combined.Add(combined, plain2)

	bfKeyBE := make([]byte, 56)
	b := combined.Bytes()
	// Pad to 56 bytes (big-endian)
	if len(b) <= 56 {
		copy(bfKeyBE[56-len(b):], b)
	} else {
		copy(bfKeyBE, b[len(b)-56:])
	}

	// Reverse to little-endian
	bfKey := make([]byte, 56)
	for i := 0; i < 56; i++ {
		bfKey[i] = bfKeyBE[55-i]
	}
	return bfKey
}

func decryptECB(c cipher.Block, data []byte) {
	bs := c.BlockSize()
	for i := 0; i+bs <= len(data); i += bs {
		c.Decrypt(data[i:i+bs], data[i:i+bs])
	}
}

// ─── Extract a file from a mix by ID ───────────────────────────────────────

func (m *mixArchive) findByID(id int32) *mixEntry {
	for i := range m.Entries {
		if m.Entries[i].ID == id {
			return &m.Entries[i]
		}
	}
	return nil
}

func (m *mixArchive) extractEntry(e *mixEntry) ([]byte, error) {
	offset := m.HeaderSize + int64(e.Offset)
	if _, err := m.data.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}
	buf := make([]byte, e.Size)
	if _, err := io.ReadFull(m.data, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (m *mixArchive) extractByName(name string) ([]byte, error) {
	id := mixID(name)
	e := m.findByID(id)
	if e == nil {
		return nil, fmt.Errorf("file %q (id %08x) not found in mix", name, uint32(id))
	}
	return m.extractEntry(e)
}

// ─── PAL (palette) format ──────────────────────────────────────────────────

type palette [256]color.RGBA

func parsePalette(data []byte) palette {
	var p palette
	for i := 0; i < 256 && i*3+2 < len(data); i++ {
		// RA2 palettes use 6-bit values (0-63), multiply by 4 to get 8-bit
		p[i] = color.RGBA{
			R: data[i*3] * 4,
			G: data[i*3+1] * 4,
			B: data[i*3+2] * 4,
			A: 255,
		}
	}
	// Index 0 is transparent in most cases
	p[0].A = 0
	return p
}

// ─── SHP (TS/RA2) format ───────────────────────────────────────────────────

type shpFrame struct {
	X, Y          uint16
	Width, Height uint16
	Compression   uint8
	_             [3]byte // alignment + radar color
	_             [4]byte // padding
	Offset        uint32
}

type shpFile struct {
	Width, Height uint16
	NumFrames     uint16
	Frames        []shpFrame
	Data          []byte
}

func parseSHP(data []byte) (*shpFile, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("shp too small")
	}
	r := bytes.NewReader(data)
	var zero uint16
	binary.Read(r, binary.LittleEndian, &zero)

	s := &shpFile{Data: data}
	binary.Read(r, binary.LittleEndian, &s.Width)
	binary.Read(r, binary.LittleEndian, &s.Height)
	binary.Read(r, binary.LittleEndian, &s.NumFrames)

	s.Frames = make([]shpFrame, s.NumFrames)
	for i := range s.Frames {
		// Each frame header is 24 bytes:
		// uint16 x, uint16 y, uint16 width, uint16 height
		// uint8 compression, uint8[3] radar/padding
		// uint32 unknown/zero, uint32 offset
		binary.Read(r, binary.LittleEndian, &s.Frames[i].X)
		binary.Read(r, binary.LittleEndian, &s.Frames[i].Y)
		binary.Read(r, binary.LittleEndian, &s.Frames[i].Width)
		binary.Read(r, binary.LittleEndian, &s.Frames[i].Height)

		var comp uint8
		binary.Read(r, binary.LittleEndian, &comp)
		s.Frames[i].Compression = comp

		var pad [7]byte
		r.Read(pad[:])

		var zero32 uint32
		binary.Read(r, binary.LittleEndian, &zero32)

		binary.Read(r, binary.LittleEndian, &s.Frames[i].Offset)
	}
	return s, nil
}

func (s *shpFile) decodeFrame(idx int, pal palette) *image.RGBA {
	if idx >= int(s.NumFrames) {
		return nil
	}
	f := &s.Frames[idx]
	w := int(f.Width)
	h := int(f.Height)
	if w == 0 || h == 0 {
		// Empty/shadow frame
		w = int(s.Width)
		h = int(s.Height)
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		return img
	}

	img := image.NewRGBA(image.Rect(0, 0, int(s.Width), int(s.Height)))

	off := int(f.Offset)
	if off >= len(s.Data) {
		return img
	}

	switch f.Compression {
	case 1:
		// Uncompressed
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				p := off + y*w + x
				if p < len(s.Data) {
					idx := s.Data[p]
					c := pal[idx]
					img.SetRGBA(int(f.X)+x, int(f.Y)+y, c)
				}
			}
		}
	case 2:
		// Compressed (scanline-based, each scanline has length prefix)
		pos := off
		for y := 0; y < h; y++ {
			if pos+2 > len(s.Data) {
				break
			}
			lineLen := int(binary.LittleEndian.Uint16(s.Data[pos : pos+2]))
			pos += 2
			x := 0
			end := pos + lineLen - 2
			if end > len(s.Data) {
				end = len(s.Data)
			}
			for pos < end && x < w {
				if pos >= len(s.Data) {
					break
				}
				v := s.Data[pos]
				pos++
				if v == 0 {
					// Skip transparent pixels
					if pos >= len(s.Data) {
						break
					}
					count := int(s.Data[pos])
					pos++
					x += count
				} else {
					c := pal[v]
					img.SetRGBA(int(f.X)+x, int(f.Y)+y, c)
					x++
				}
			}
			// Ensure we advance to end of scanline
			if pos < end {
				pos = end
			}
		}
	case 3:
		// Compressed with scanline offsets (most common in RA2)
		// Frame data starts with h uint16 offsets (relative to frame start), then pixel data
		pos := off
		for y := 0; y < h; y++ {
			if pos+2 > len(s.Data) {
				break
			}
			lineLen := int(binary.LittleEndian.Uint16(s.Data[pos : pos+2]))
			pos += 2
			x := 0
			end := pos + lineLen - 2
			if end > len(s.Data) {
				end = len(s.Data)
			}
			for pos < end && x < w {
				if pos >= len(s.Data) {
					break
				}
				v := s.Data[pos]
				pos++
				if v == 0 {
					if pos >= len(s.Data) {
						break
					}
					count := int(s.Data[pos])
					pos++
					x += count
				} else {
					c := pal[v]
					img.SetRGBA(int(f.X)+x, int(f.Y)+y, c)
					x++
				}
			}
			if pos < end {
				pos = end
			}
		}
	default:
		// Try as uncompressed
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				p := off + y*w + x
				if p < len(s.Data) {
					idx := s.Data[p]
					c := pal[idx]
					img.SetRGBA(int(f.X)+x, int(f.Y)+y, c)
				}
			}
		}
	}
	return img
}

// ─── Cameo SHP (PCX-like, 60x48 icons used in sidebar) ─────────────────────

// Cameos are regular SHP files, just smaller.

// ─── Main extraction logic ─────────────────────────────────────────────────

type extractTarget struct {
	Name     string // filename inside mix (e.g., "gayard.shp")
	Category string // output subdirectory
}

func main() {
	inputPath := flag.String("input", "", "Path to ra2.mix")
	outputPath := flag.String("output", "assets/ra2", "Output directory")
	listOnly := flag.Bool("list", false, "Just list mix contents (IDs)")
	dumpAll := flag.Bool("dump-all", false, "Dump all raw files from mix")
	flag.Parse()

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: extract_ra2 -input <ra2.mix> -output <dir>")
		os.Exit(1)
	}

	f, err := os.Open(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	mix, err := readMix(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read mix: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Mix: flags=0x%08x files=%d bodySize=%d headerSize=%d\n",
		mix.Flags, mix.FileCount, mix.BodySize, mix.HeaderSize)

	if *listOnly {
		for _, e := range mix.Entries {
			fmt.Printf("  ID=%08x offset=%d size=%d\n", uint32(e.ID), e.Offset, e.Size)
		}
		return
	}

	// Try extracting specific files by name (for debugging)
	extractName := flag.Lookup("extract-name")
	_ = extractName

	if *dumpAll {
		dumpDir := filepath.Join(*outputPath, "raw_dump")
		os.MkdirAll(dumpDir, 0755)
		for _, e := range mix.Entries {
			data, err := mix.extractEntry(&e)
			if err != nil {
				fmt.Printf("  extract %08x: %v\n", uint32(e.ID), err)
				continue
			}
			fname := fmt.Sprintf("%08x.bin", uint32(e.ID))
			os.WriteFile(filepath.Join(dumpDir, fname), data, 0644)
			fmt.Printf("  dumped %s (%d bytes)\n", fname, len(data))
		}
		return
	}

	// ── Step 1: Extract palette from inside the mix ──

	// Try to find palette in various locations
	var unitPal palette
	var havePal bool

	// The palettes are usually in cache.mix or inside ra2.mix directly
	// Try common palette names
	palNames := []string{"unittem.pal", "unit.pal", "temperat.pal", "isotem.pal"}
	for _, pn := range palNames {
		data, err := mix.extractByName(pn)
		if err == nil && len(data) >= 768 {
			unitPal = parsePalette(data)
			havePal = true
			fmt.Printf("Found palette: %s\n", pn)
			break
		}
	}

	// If not found directly, look in nested mixes
	nestedMixNames := []string{"cache.mix", "local.mix", "conquer.mix", "cameo.mix",
		"generic.mix", "isogen.mix", "isotemp.mix", "isosnow.mix", "isourb.mix"}

	type nestedMix struct {
		name string
		data []byte
		mix  *mixArchive
	}
	var nestedMixes []nestedMix

	for _, nm := range nestedMixNames {
		data, err := mix.extractByName(nm)
		if err != nil {
			continue
		}
		fmt.Printf("Found nested mix: %s (%d bytes)\n", nm, len(data))
		sub, err := readMix(bytes.NewReader(data))
		if err != nil {
			fmt.Printf("  failed to parse: %v\n", err)
			continue
		}
		fmt.Printf("  -> %d files\n", sub.FileCount)
		nestedMixes = append(nestedMixes, nestedMix{name: nm, data: data, mix: sub})

		if !havePal {
			for _, pn := range palNames {
				pdata, err := sub.extractByName(pn)
				if err == nil && len(pdata) >= 768 {
					unitPal = parsePalette(pdata)
					havePal = true
					fmt.Printf("  Found palette: %s in %s\n", pn, nm)
					break
				}
			}
		}
	}

	if !havePal {
		fmt.Println("WARNING: No palette found, using grayscale")
		for i := 0; i < 256; i++ {
			unitPal[i] = color.RGBA{uint8(i), uint8(i), uint8(i), 255}
		}
		unitPal[0].A = 0
	}

	// Save palette
	palDir := filepath.Join(*outputPath, "palettes")
	os.MkdirAll(palDir, 0755)
	if havePal {
		// Save palette as image for visualization
		palImg := image.NewRGBA(image.Rect(0, 0, 16, 16))
		for i := 0; i < 256; i++ {
			palImg.SetRGBA(i%16, i/16, unitPal[i])
		}
		savePNG(filepath.Join(palDir, "unittem_preview.png"), palImg)
	}

	// ── Extract INI files to identify correct art names ──
	iniDir := filepath.Join(*outputPath, "ini")
	os.MkdirAll(iniDir, 0755)
	iniNames := []string{"rules.ini", "art.ini", "rulesmd.ini", "artmd.ini"}
	for _, in := range iniNames {
		for _, nm := range nestedMixes {
			data, err := nm.mix.extractByName(in)
			if err == nil && len(data) > 100 {
				os.WriteFile(filepath.Join(iniDir, in), data, 0644)
				fmt.Printf("Extracted INI: %s from %s (%d bytes)\n", in, nm.name, len(data))
				break
			}
		}
	}

	// Also extract cameo.mix if found
	for _, nm := range nestedMixNames {
		if nm == "cameo.mix" {
			data, err := mix.extractByName(nm)
			if err == nil {
				sub, err := readMix(bytes.NewReader(data))
				if err == nil {
					fmt.Printf("Cameo.mix: %d files\n", sub.FileCount)
					cameoDir := filepath.Join(*outputPath, "ui", "cameos")
					os.MkdirAll(cameoDir, 0755)
					for _, e := range sub.Entries {
						edata, err := sub.extractEntry(&e)
						if err != nil || len(edata) < 8 {
							continue
						}
						if edata[0] == 0 && edata[1] == 0 {
							shp, err := parseSHP(edata)
							if err == nil && shp.NumFrames > 0 {
								img := shp.decodeFrame(0, unitPal)
								if img != nil {
									outFile := filepath.Join(cameoDir, fmt.Sprintf("%08x.png", uint32(e.ID)))
									savePNG(outFile, img)
								}
							}
						}
					}
				}
			}
		}
	}

	// ── Step 2: Define targets ──
	alliedBuildings := []string{
		"gacnst", "gapowr", "gapile", "gaweap", "garefn", "gatech", "gawall",
		"gayard", "gaspysat", "gaairc",
	}
	sovietBuildings := []string{
		"nacnst", "napowr", "napile", "naweap", "narefn", "natech", "nawall",
		"nayard", "nalasr", "naflak", "tesla",
	}
	turrets := []string{"gturret", "nturret", "atesla", "nasam", "gacsph", "nairon"}
	units := []string{"mcv", "mtnk", "htnk", "harv", "horv", "e1", "e2", "gi", "dog",
		"conscript", "snipe", "ivan", "tanya", "seal", "engineer", "flakt", "dest",
		"aegis", "carrier", "dred", "squid", "dolphin"}
	cameos := []string{
		// Allied buildings
		"powricon", "brrkicon", "gwepicon", "reficon", "radricon",
		"techicon", "wallicon", "ayaricon", "csphicon", "pillicon",
		"prisicon", "tpwricon", "gateicon",
		// Soviet buildings
		"npwricon", "handicon", "nwepicon", "nreficon", "nradicon",
		"ntchicon", "nwalicon", "tslaicon", "flakicon", "ironicon",
		"clonicon", "lasricon",
		// Allied units
		"giicon", "engnicon", "adogicon", "mtnkicon", "fvicon",
		"harvicon", "mcvicon", "gtnkicon", "sealicon", "spyicon",
		"tanyicon", "snipicon", "carricon", "desticon", "dlphicon",
		// Soviet units
		"dogicon", "e2icon", "desoicon", "rtnkicon", "v3icon",
		"dredicon", "sqdicon", "ivanicon", "yuriicon",
		// extra
		"agisicon", "htnkicon",
	}

	// Build extraction list
	type target struct {
		name   string
		ext    string
		outDir string
	}
	var targets []target

	for _, b := range alliedBuildings {
		targets = append(targets, target{b, ".shp", filepath.Join(*outputPath, "buildings", "allied")})
		targets = append(targets, target{b, ".tem", filepath.Join(*outputPath, "buildings", "allied")})
	}
	for _, b := range sovietBuildings {
		targets = append(targets, target{b, ".shp", filepath.Join(*outputPath, "buildings", "soviet")})
		targets = append(targets, target{b, ".tem", filepath.Join(*outputPath, "buildings", "soviet")})
	}
	for _, t := range turrets {
		targets = append(targets, target{t, ".shp", filepath.Join(*outputPath, "buildings", "turrets")})
	}
	for _, u := range units {
		targets = append(targets, target{u, ".shp", filepath.Join(*outputPath, "units")})
		targets = append(targets, target{u, ".vxl", filepath.Join(*outputPath, "units")})
	}
	for _, c := range cameos {
		targets = append(targets, target{c, ".shp", filepath.Join(*outputPath, "ui", "cameos")})
	}

	// ── Step 3: Search and extract ──
	extracted := 0
	for _, tgt := range targets {
		fname := tgt.name + tgt.ext
		var data []byte

		// Try main mix first
		data, err = mix.extractByName(fname)
		if err != nil {
			// Try nested mixes
			for _, nm := range nestedMixes {
				data, err = nm.mix.extractByName(fname)
				if err == nil {
					fmt.Printf("Found %s in %s\n", fname, nm.name)
					break
				}
			}
		}

		if err != nil || data == nil {
			continue
		}

		if tgt.ext == ".shp" && len(data) > 8 {
			os.MkdirAll(tgt.outDir, 0755)
			shp, err := parseSHP(data)
			if err != nil {
				fmt.Printf("  parse SHP %s: %v\n", fname, err)
				continue
			}
			fmt.Printf("Extracting %s: %dx%d, %d frames\n", fname, shp.Width, shp.Height, shp.NumFrames)

			// Save first frame as the main sprite
			if shp.NumFrames > 0 {
				img := shp.decodeFrame(0, unitPal)
				if img != nil {
					outFile := filepath.Join(tgt.outDir, strings.ToUpper(tgt.name)+".png")
					savePNG(outFile, img)
					extracted++
				}
			}

			// Also save all frames as a sprite sheet
			if shp.NumFrames > 1 {
				sheet := makeSpriteSheet(shp, unitPal)
				if sheet != nil {
					outFile := filepath.Join(tgt.outDir, strings.ToUpper(tgt.name)+"_sheet.png")
					savePNG(outFile, sheet)
				}
			}

			// Save raw SHP for reference
			rawDir := filepath.Join(tgt.outDir, "raw")
			os.MkdirAll(rawDir, 0755)
			os.WriteFile(filepath.Join(rawDir, fname), data, 0644)
		} else if tgt.ext == ".vxl" {
			// Save raw VXL (we don't render voxels to 2D here, just save for reference)
			rawDir := filepath.Join(tgt.outDir, "raw")
			os.MkdirAll(rawDir, 0755)
			os.WriteFile(filepath.Join(rawDir, fname), data, 0644)
			fmt.Printf("Saved VXL: %s (%d bytes)\n", fname, len(data))
			extracted++
		}
	}

	fmt.Printf("\nExtracted %d assets to %s\n", extracted, *outputPath)

	// ── Step 4: Also dump any SHP files we can find in all mixes ──
	fmt.Println("\nScanning all nested mixes for SHP/PAL files...")
	allShpDir := filepath.Join(*outputPath, "all_shp")
	os.MkdirAll(allShpDir, 0755)

	shpCount := 0
	for _, nm := range nestedMixes {
		for _, e := range nm.mix.Entries {
			data, err := nm.mix.extractEntry(&e)
			if err != nil || len(data) < 8 {
				continue
			}
			// Quick heuristic: SHP(TS) files start with 0x0000 and have reasonable dimensions
			if data[0] == 0 && data[1] == 0 {
				w := binary.LittleEndian.Uint16(data[2:4])
				h := binary.LittleEndian.Uint16(data[4:6])
				nf := binary.LittleEndian.Uint16(data[6:8])
				if w > 0 && w < 2000 && h > 0 && h < 2000 && nf > 0 && nf < 10000 {
					// Likely a SHP file
					expectedMinSize := 8 + int(nf)*24
					if len(data) > expectedMinSize {
						shp, err := parseSHP(data)
						if err == nil && shp.NumFrames > 0 {
							img := shp.decodeFrame(0, unitPal)
							if img != nil {
								outFile := filepath.Join(allShpDir,
									fmt.Sprintf("%s_%08x.png", nm.name, uint32(e.ID)))
								savePNG(outFile, img)
								shpCount++
							}
						}
					}
				}
			}
		}
	}
	fmt.Printf("Extracted %d additional SHP sprites from nested mixes\n", shpCount)
}

func makeSpriteSheet(shp *shpFile, pal palette) *image.RGBA {
	cols := 8
	if int(shp.NumFrames) < cols {
		cols = int(shp.NumFrames)
	}
	rows := (int(shp.NumFrames) + cols - 1) / cols

	w := int(shp.Width) * cols
	h := int(shp.Height) * rows
	sheet := image.NewRGBA(image.Rect(0, 0, w, h))

	for i := 0; i < int(shp.NumFrames); i++ {
		frame := shp.decodeFrame(i, pal)
		if frame == nil {
			continue
		}
		col := i % cols
		row := i / cols
		ox := col * int(shp.Width)
		oy := row * int(shp.Height)
		for y := 0; y < int(shp.Height); y++ {
			for x := 0; x < int(shp.Width); x++ {
				c := frame.RGBAAt(x, y)
				if c.A > 0 {
					sheet.SetRGBA(ox+x, oy+y, c)
				}
			}
		}
	}
	return sheet
}

func savePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("  create %s: %v\n", path, err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fmt.Printf("  encode %s: %v\n", path, err)
	}
}
