package asciiart

import (
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func testGenerator(t *testing.T) *Generator {
	t.Helper()
	g, err := New(os.DirFS("../../banners"))
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestKnownGlyphAndNewlines(t *testing.T) {
	g := testGenerator(t)
	want := " _  \n| | \n| | \n| | \n|_| \n(_) \n    \n    \n"
	for _, tc := range []struct{ input, want string }{
		{"!", want}, {"!!", " _   _  \n| | | | \n| | | | \n| | | | \n|_| |_| \n(_) (_) \n        \n        \n"},
		{"!\n!", want + want}, {"!\r\n!", want + want},
		{"\n!\n\n", "\n" + want + "\n\n"},
	} {
		got, err := g.Render(tc.input, "standard")
		if err != nil || got != tc.want {
			t.Errorf("Render(%q) = %q, %v; want %q", tc.input, got, err, tc.want)
		}
	}
}

func TestAllPrintableCharactersAndStyles(t *testing.T) {
	g := testGenerator(t)
	var input strings.Builder
	for c := byte(32); c <= 126; c++ {
		input.WriteByte(c)
	}
	for _, style := range styles {
		result, err := g.Render(input.String(), style)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(result, "\n") != 8 {
			t.Errorf("%s: expected eight rows", style)
		}
		space, err := g.Render(" ", style)
		if err != nil || strings.TrimSpace(space) != "" || len(space) <= 8 {
			t.Errorf("%s: spaces not preserved", style)
		}
	}
}

func TestInputValidation(t *testing.T) {
	g := testGenerator(t)
	for _, input := range []string{"", "\n", "\r\n", "\t", "hello\rworld", "café", "\x00", "\x7f", strings.Repeat("a", 101)} {
		if _, err := g.Render(input, "standard"); err == nil {
			t.Errorf("accepted invalid input %q", input)
		}
	}
	for _, style := range []string{"", "Standard", "../standard", "unknown"} {
		if _, err := g.Render("test", style); err == nil {
			t.Errorf("accepted style %q", style)
		}
	}
	if _, err := g.Render(strings.Repeat("a", 100), "standard"); err != nil {
		t.Fatal(err)
	}
}

func TestBannerValidation(t *testing.T) {
	data, err := fs.ReadFile(os.DirFS("../../banners"), "standard.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"truncated", []byte("\nshort\n")},
		{"bad separator", append([]byte("x"), data...)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := New(fstest.MapFS{"standard.txt": &fstest.MapFile{Data: tc.data}}); err == nil {
				t.Fatal("accepted malformed banner")
			}
		})
	}
	if _, err := New(fstest.MapFS{}); err == nil {
		t.Fatal("accepted missing banners")
	}
	files := fstest.MapFS{}
	for _, style := range styles {
		lf := strings.ReplaceAll(string(data), "\r\n", "\n")
		files[style+".txt"] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(lf, "\n", "\r\n"))}
	}
	if _, err := New(files); err != nil {
		t.Fatalf("CRLF banners: %v", err)
	}
}
