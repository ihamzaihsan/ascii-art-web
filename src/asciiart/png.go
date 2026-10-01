package asciiart

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"
)

// PNGImage rasterizes printable ASCII with bundled Thinkertoy glyphs. Fixed-size
// cells preserve whitespace without a font library or platform font dependency.
func (g *Generator) PNGImage(text string, frame int) (image.Image, error) {
	if frame < 0 || frame > 100 || len(text) == 0 || len(text) > 100000 {
		return nil, fmt.Errorf("invalid artwork size or PNG frame")
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	columns := 0
	for _, line := range lines {
		columns = max(columns, len(line))
		for _, r := range line {
			if r < 32 || r > 126 {
				return nil, fmt.Errorf("PNG artwork must contain printable ASCII and newlines")
			}
		}
	}
	font := g.banners["thinkertoy"]
	cellWidth := 0
	for _, glyph := range font {
		for _, row := range glyph {
			cellWidth = max(cellWidth, len(row))
		}
	}
	cellWidth++
	const cellHeight = 14
	w, h := columns*cellWidth+16+frame*2, len(lines)*cellHeight+16+frame*2
	if w > 16384 || h > 16384 || w*h > 24000000 {
		return nil, fmt.Errorf("artwork is too large for PNG; use TXT instead")
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, image.Rect(frame, frame, w-frame, h-frame), image.NewUniform(color.White), image.Point{}, draw.Src)
	ink := color.RGBA{16, 27, 27, 255}
	for y, line := range lines {
		for x, ch := range []byte(line) {
			for gy, row := range font[ch-32] {
				for gx, dot := range []byte(row) {
					if dot != ' ' {
						dst.SetRGBA(frame+8+x*cellWidth+gx, frame+8+y*cellHeight+gy, ink)
					}
				}
			}
		}
	}
	return dst, nil
}
