package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
)

func inspectBytes(data []byte) (map[string]any, error) {
	if len(data) > 64<<20 {
		return nil, errors.New("image exceeds 64 MiB limit")
	}
	cfg, format, e := image.DecodeConfig(bytes.NewReader(data))
	if e != nil {
		return nil, errors.New("output is not a supported decodable PNG/JPEG/GIF")
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > 32_000_000 {
		return nil, errors.New("image dimensions exceed 32 million pixel limit")
	}
	result := map[string]any{"format": format, "width": cfg.Width, "height": cfg.Height, "bytes": len(data)}
	if format != "gif" {
		im, _, e := image.Decode(bytes.NewReader(data))
		if e != nil {
			return nil, errors.New("truncated or invalid image")
		}
		opaque := true
		if v, ok := im.(interface{ Opaque() bool }); ok {
			opaque = v.Opaque()
		}
		result["has_transparency"] = !opaque
		return result, nil
	}
	// Bound aggregate frame area before allocating decoded frames.
	if e := gifBudget(data); e != nil {
		return nil, e
	}
	g, e := gif.DecodeAll(bytes.NewReader(data))
	if e != nil {
		return nil, errors.New("truncated or invalid GIF")
	}
	total := 0
	delays := make([]int, len(g.Delay))
	for i, d := range g.Delay {
		delays[i] = d * 10
		total += d * 10
	}
	result["frames"] = len(g.Image)
	result["duration_ms"] = total
	result["frame_delays_ms"] = delays
	result["loop_count"] = g.LoopCount
	canvas := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
	var first []byte
	bg := color.Color(color.Transparent)
	if pal, ok := g.Config.ColorModel.(color.Palette); ok && int(g.BackgroundIndex) < len(pal) {
		bg = pal[g.BackgroundIndex]
	}
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)
	for i, frame := range g.Image {
		previous := append([]byte(nil), canvas.Pix...)
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)
		if i == 0 {
			first = append([]byte(nil), canvas.Pix...)
		}
		if i == len(g.Image)-1 {
			result["first_last_equal"] = bytes.Equal(first, canvas.Pix)
		}
		if i < len(g.Disposal) {
			switch g.Disposal[i] {
			case gif.DisposalBackground:
				draw.Draw(canvas, frame.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)
			case gif.DisposalPrevious:
				copy(canvas.Pix, previous)
			}
		}
	}
	result["seam_note"] = "Pixel equality measures endpoints only; motion and identity require visual review."
	return result, nil
}

// Read GIF block lengths without decompressing; bound total decoded frame pixels.
func gifBudget(b []byte) error {
	if len(b) < 13 {
		return errors.New("invalid GIF header")
	}
	pos := 13
	if b[10]&128 != 0 {
		pos += 3 * (1 << ((b[10] & 7) + 1))
	}
	skip := func() bool {
		for pos < len(b) {
			n := int(b[pos])
			pos++
			if n == 0 {
				return true
			}
			pos += n
		}
		return false
	}
	screenPixels := int64(int(b[6])|int(b[7])<<8) * int64(int(b[8])|int(b[9])<<8)
	frames := 0
	pixels := int64(0)
	for pos < len(b) {
		tag := b[pos]
		pos++
		switch tag {
		case 0x3b:
			return nil
		case 0x21:
			pos++
			if !skip() {
				return errors.New("invalid GIF extension")
			}
		case 0x2c:
			if pos+9 > len(b) {
				return errors.New("invalid GIF descriptor")
			}
			w := int(b[pos+4]) | int(b[pos+5])<<8
			h := int(b[pos+6]) | int(b[pos+7])<<8
			packed := b[pos+8]
			pos += 9
			frames++
			pixels += int64(w) * int64(h)
			if frames > 500 || pixels > 64_000_000 || screenPixels*int64(frames) > 64_000_000 {
				return errors.New("GIF exceeds 500 frames or 64 million aggregate pixels")
			}
			if packed&128 != 0 {
				pos += 3 * (1 << ((packed & 7) + 1))
			}
			pos++
			if !skip() {
				return errors.New("invalid GIF image blocks")
			}
		default:
			return errors.New("invalid GIF block")
		}
	}
	return errors.New("GIF trailer missing")
}
func validateImageOutputs(images [][]byte, paths []string) error {
	if len(images) != len(paths) {
		return errors.New("image count mismatch")
	}
	for i, b := range images {
		m, e := inspectBytes(b)
		if e != nil {
			return e
		}
		ext := strings.ToLower(filepath.Ext(paths[i]))
		expected := map[string]string{".png": "png", ".jpg": "jpeg", ".jpeg": "jpeg", ".gif": "gif"}[ext]
		if expected == "" || m["format"] != expected {
			return fmt.Errorf("actual format %s does not match output extension %s", m["format"], ext)
		}
	}
	return nil
}
func actualOutputs(images [][]byte) []map[string]any {
	out := make([]map[string]any, 0, len(images))
	for _, b := range images {
		m, _ := inspectBytes(b)
		out = append(out, m)
	}
	return out
}
func runInspect(argv []string) error {
	if len(argv) != 1 {
		return errors.New("usage: inspect <image-path>")
	}
	f, e := os.Open(argv[0])
	if e != nil {
		return e
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil {
		return e
	}
	if st.Size() > 64<<20 {
		return errors.New("image exceeds 64 MiB limit")
	}
	b, e := os.ReadFile(argv[0])
	if e != nil {
		return e
	}
	m, e := inspectBytes(b)
	if e == nil {
		printJSON(m)
	}
	return e
}
