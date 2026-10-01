// Package imageart converts decoded images to monochrome ASCII using bounded grids.
package imageart

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
)

// Control describes a numeric HTML input and its server-side bounds.
type Control struct {
	Name, Label, Unit       string
	Min, Max, Step, Default float64
}

var controls = []Control{
	{"characters", "Characters (columns)", "", 20, 300, 1, 100},
	{"brightness", "Brightness", "%", 0, 200, 1, 100},
	{"contrast", "Contrast", "%", 0, 200, 1, 100},
	{"saturation", "Saturation", "%", 0, 400, 1, 100},
	{"hue", "Hue", "degrees", 0, 360, 1, 0},
	{"grayscale", "Grayscale", "%", 0, 100, 1, 0},
	{"sepia", "Sepia", "%", 0, 100, 1, 0},
	{"invert", "Invert Colors", "%", 0, 100, 1, 0},
	{"threshold", "Threshold level", "", 0, 255, 1, 128},
	{"sharpness", "Sharpness strength", "", 0, 20, 1, 9},
	{"edges", "Edge strength", "", 0, 4, .1, 1},
	{"space-density", "Space Density", "", 0, 10, 1, 1},
	{"frame", "Transparent PNG frame", "px", 0, 100, 1, 0},
}

func Controls() []Control { return append([]Control(nil), controls...) }

type Options struct {
	Values                      map[string]float64
	Gradient, Quality           string
	Threshold, Sharpness, Edges bool
}

func Defaults() Options {
	o := Options{Values: make(map[string]float64), Gradient: "normal", Quality: "none"}
	for _, c := range controls {
		o.Values[c.Name] = c.Default
	}
	return o
}

var gradients = map[string]string{
	"normal":   "@%#*+=-:. ",
	"detailed": "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/\\|()1{}[]?-_+~<>i!lI;:,\"^`'. ",
	"blocks":   "@#+=. ", "binary": "# ", "reversed": " .:-=+*#%@",
}

func (o Options) Validate() error {
	for _, c := range controls {
		v, ok := o.Values[c.Name]
		if !ok || math.IsNaN(v) || math.IsInf(v, 0) || v < c.Min || v > c.Max || math.Abs(v/c.Step-math.Round(v/c.Step)) > 1e-6 {
			return fmt.Errorf("%s must be between %g and %g in steps of %g", c.Label, c.Min, c.Max, c.Step)
		}
	}
	if _, ok := gradients[o.Gradient]; !ok {
		return fmt.Errorf("choose a supported ASCII gradient")
	}
	switch o.Quality {
	case "none", "smooth", "normalize", "both":
	default:
		return fmt.Errorf("choose a supported quality enhancement")
	}
	return nil
}

func clamp(v float64) float64 { return math.Max(0, math.Min(255, v)) }

// RGB composites any image color onto white, including premultiplied alpha.
func RGB(c color.Color) [3]float64 {
	r, g, b, a := c.RGBA()
	return [3]float64{float64(r+65535-a) / 257, float64(g+65535-a) / 257, float64(b+65535-a) / 257}
}

func sample(src image.Image, x, y float64, smooth bool) [3]float64 {
	b := src.Bounds()
	x = math.Max(0, math.Min(float64(b.Dx()-1), x))
	y = math.Max(0, math.Min(float64(b.Dy()-1), y))
	at := func(x, y int) [3]float64 { return RGB(src.At(b.Min.X+x, b.Min.Y+y)) }
	if !smooth {
		return at(int(math.Round(x)), int(math.Round(y)))
	}
	x0, y0 := int(x), int(y)
	x1, y1 := min(x0+1, b.Dx()-1), min(y0+1, b.Dy()-1)
	a, b1, c, d := at(x0, y0), at(x1, y0), at(x0, y1), at(x1, y1)
	fx, fy := x-float64(x0), y-float64(y0)
	var rgb [3]float64
	for i := range rgb {
		rgb[i] = (a[i]*(1-fx)+b1[i]*fx)*(1-fy) + (c[i]*(1-fx)+d[i]*fx)*fy
	}
	return rgb
}

// Thumbnail creates a bounded opaque preview, also retained in the HTML form.
func Thumbnail(src image.Image) *image.RGBA {
	b := src.Bounds()
	scale := math.Min(1, 600/float64(max(b.Dx(), b.Dy())))
	w, h := max(1, int(float64(b.Dx())*scale)), max(1, int(float64(b.Dy())*scale))
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			rgb := sample(src, (float64(x)+.5)*float64(b.Dx())/float64(w)-.5, (float64(y)+.5)*float64(b.Dy())/float64(h)-.5, true)
			dst.SetRGBA(x, y, color.RGBA{uint8(rgb[0]), uint8(rgb[1]), uint8(rgb[2]), 255})
		}
	}
	return dst
}

func adjust(rgb [3]float64, o Options) float64 {
	v := o.Values
	r, g, b := rgb[0], rgb[1], rgb[2]
	a := v["hue"] * math.Pi / 180
	c, s := math.Cos(a), math.Sin(a)
	hr := (.213+c*.787-s*.213)*r + (.715-c*.715-s*.715)*g + (.072-c*.072+s*.928)*b
	hg := (.213-c*.213+s*.143)*r + (.715+c*.285+s*.140)*g + (.072-c*.072-s*.283)*b
	hb := (.213-c*.213-s*.787)*r + (.715-c*.715+s*.715)*g + (.072+c*.928+s*.072)*b
	gray := .2126*hr + .7152*hg + .0722*hb
	sat := v["saturation"] / 100 * (1 - v["grayscale"]/100)
	r, g, b = gray+(hr-gray)*sat, gray+(hg-gray)*sat, gray+(hb-gray)*sat
	sepia := v["sepia"] / 100
	r, g, b = r*(1-sepia)+(.393*r+.769*g+.189*b)*sepia, g*(1-sepia)+(.349*r+.686*g+.168*b)*sepia, b*(1-sepia)+(.272*r+.534*g+.131*b)*sepia
	channel := func(n float64) float64 {
		n = clamp((n*v["brightness"]/100-128)*v["contrast"]/100 + 128)
		inv := v["invert"] / 100
		return n*(1-inv) + (255-n)*inv
	}
	return .2126*channel(r) + .7152*channel(g) + .0722*channel(b)
}

func Render(src image.Image, o Options) (string, int, error) {
	if err := o.Validate(); err != nil {
		return "", 0, err
	}
	if src == nil {
		return "", 0, fmt.Errorf("image is empty")
	}
	b := src.Bounds()
	if b.Empty() {
		return "", 0, fmt.Errorf("image is empty")
	}
	w := int(o.Values["characters"])
	h := max(1, min(300, int(math.Round(float64(b.Dy())/float64(b.Dx())*float64(w)*7.2/14))))
	values := make([]float64, w*h)
	smooth := o.Quality == "smooth" || o.Quality == "both"
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			rgb := sample(src, (float64(x)+.5)*float64(b.Dx())/float64(w)-.5, (float64(y)+.5)*float64(b.Dy())/float64(h)-.5, smooth)
			values[y*w+x] = adjust(rgb, o)
		}
	}
	if o.Quality == "normalize" || o.Quality == "both" {
		low, high := 255.0, 0.0
		for _, v := range values {
			low = math.Min(low, v)
			high = math.Max(high, v)
		}
		if high > low {
			for i, v := range values {
				values[i] = (v - low) * 255 / (high - low)
			}
		}
	}
	filter := func(edges bool) {
		out := make([]float64, len(values))
		at := func(x, y int) float64 { return values[max(0, min(h-1, y))*w+max(0, min(w-1, x))] }
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if edges {
					gx := -at(x-1, y-1) + at(x+1, y-1) - 2*at(x-1, y) + 2*at(x+1, y) - at(x-1, y+1) + at(x+1, y+1)
					gy := -at(x-1, y-1) - 2*at(x, y-1) - at(x+1, y-1) + at(x-1, y+1) + 2*at(x, y+1) + at(x+1, y+1)
					out[y*w+x] = 255 - clamp(math.Hypot(gx, gy)*o.Values["edges"])
				} else {
					out[y*w+x] = clamp(at(x, y) + o.Values["sharpness"]/10*(4*at(x, y)-at(x-1, y)-at(x+1, y)-at(x, y-1)-at(x, y+1)))
				}
			}
		}
		values = out
	}
	if o.Sharpness {
		filter(false)
	}
	if o.Edges {
		filter(true)
	}
	gradient := gradients[o.Gradient]
	spaces := strings.Repeat(" ", int(o.Values["space-density"]))
	if o.Gradient == "reversed" {
		gradient = spaces + gradient
	} else {
		gradient += spaces
	}
	var out strings.Builder
	out.Grow(w*h + h)
	for y := 0; y < h; y++ {
		if y > 0 {
			out.WriteByte('\n')
		}
		for x := 0; x < w; x++ {
			v := values[y*w+x]
			if o.Threshold {
				if v >= o.Values["threshold"] {
					v = 255
				} else {
					v = 0
				}
			}
			out.WriteByte(gradient[int(math.Round(clamp(v)/255*float64(len(gradient)-1)))])
		}
	}
	return out.String(), h, nil
}
