// Package asciiart renders printable ASCII text using eight-row banner glyphs.
package asciiart

import (
	"fmt"
	"io/fs"
	"strings"
)

const MaxInputLength = 100

var styles = []string{"standard", "shadow", "thinkertoy"}

type glyph [8]string

// Generator stores validated, immutable banners and is safe for concurrent use.
type Generator struct {
	banners map[string][95]glyph
}

// New loads all supported banners once, accepting LF and CRLF files.
func New(files fs.FS) (*Generator, error) {
	g := &Generator{banners: make(map[string][95]glyph)}
	for _, style := range styles {
		data, err := fs.ReadFile(files, style+".txt")
		if err != nil {
			return nil, fmt.Errorf("load %s banner: %w", style, err)
		}
		lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
		if len(lines) == 856 && lines[855] == "" {
			lines = lines[:855]
		}
		if len(lines) != 855 {
			return nil, fmt.Errorf("%s banner: expected 855 lines, got %d", style, len(lines))
		}
		var characters [95]glyph
		for i := range characters {
			if lines[i*9] != "" {
				return nil, fmt.Errorf("%s banner: missing separator for character %d", style, i+32)
			}
			copy(characters[i][:], lines[i*9+1:i*9+9])
		}
		g.banners[style] = characters
	}
	return g, nil
}

// Render preserves spaces and trailing newlines. Empty lines produce one newline;
// nonempty lines produce eight rows. Input length is measured after CRLF normalization.
func (g *Generator) Render(input, style string) (string, error) {
	characters, ok := g.banners[style]
	if !ok {
		return "", fmt.Errorf("choose standard, shadow, or thinkertoy")
	}
	input = strings.ReplaceAll(input, "\r\n", "\n")
	if len(input) == 0 || len(input) > MaxInputLength || strings.Trim(input, "\n") == "" {
		return "", fmt.Errorf("enter 1–%d characters with at least one printable character", MaxInputLength)
	}
	for _, character := range input {
		if character != '\n' && (character < 32 || character > 126) {
			return "", fmt.Errorf("use printable ASCII characters (spaces, letters, numbers, and punctuation) and newlines only")
		}
	}
	var output strings.Builder
	for _, line := range strings.Split(input, "\n") {
		if line == "" {
			output.WriteByte('\n')
			continue
		}
		for row := 0; row < 8; row++ {
			for _, character := range line {
				output.WriteString(characters[character-32][row])
			}
			output.WriteByte('\n')
		}
	}
	return output.String(), nil
}
