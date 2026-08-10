package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeTestPNG(t *testing.T, path string, w, h int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestWrapText(t *testing.T) {
	fs, err := loadFaces()
	if err != nil {
		t.Fatal(err)
	}

	lines := wrapText("alpha beta gamma delta epsilon zeta eta theta", fs.body, 600)
	if len(lines) < 2 {
		t.Fatalf("expected wrapping, got %v", lines)
	}
	for _, l := range lines {
		if strings.HasPrefix(l, " ") || strings.HasSuffix(l, " ") {
			t.Errorf("line %q has surrounding spaces", l)
		}
	}

	if got := wrapText("word", fs.body, 600); len(got) != 1 || got[0] != "word" {
		t.Errorf("single word: %v", got)
	}
	if got := wrapText("", fs.body, 600); got != nil {
		t.Errorf("empty text: %v", got)
	}
}

func TestRenderEndToEnd(t *testing.T) {
	root := t.TempDir()
	writeTestPNG(t, filepath.Join(root, "assets", "logos", "full.png"), 300, 100)
	writeTestPNG(t, filepath.Join(root, "assets", "logos", "icon.png"), 100, 100)

	p := Project{
		Name:        "testproj",
		Type:        "Test Suite",
		Tagline:     "Line one.\nLine two.",
		Description: "A description long enough to wrap across more than one line at the configured maximum width for body text.",
		URL:         "https://example.com",
		Background:  "#2D2B55",
		Accent:      []string{"#6C5CE7", "#8577ED"},
		LogoFull:    "assets/logos/full.png",
		LogoIcon:    "assets/logos/icon.png",
	}

	out, err := Render(root, p, time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	want := filepath.Join(root, "assets", "testproj", "og-image.png")
	if out != want {
		t.Errorf("out = %s, want %s", out, want)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 2400 || img.Bounds().Dy() != 1260 {
		t.Fatalf("bounds = %v, want 2400x1260", img.Bounds())
	}

	nrgbaAt := func(x, y int) color.NRGBA {
		r, g, b, a := img.At(x, y).RGBA()
		return color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
	}
	if got := nrgbaAt(50, 1200); got != (color.NRGBA{R: 45, G: 43, B: 85, A: 255}) {
		t.Errorf("background pixel = %+v", got)
	}
	if got := nrgbaAt(0, 3); got != (color.NRGBA{R: 108, G: 92, B: 231, A: 255}) {
		t.Errorf("accent start pixel = %+v", got)
	}
}

func TestRenderMissingLogoFails(t *testing.T) {
	root := t.TempDir()
	p := Project{
		Name: "x", Tagline: "t", URL: "u",
		Background: "#000000", Accent: []string{"#FFFFFF"},
		LogoFull: "missing.png", LogoIcon: "missing.png",
	}
	if _, err := Render(root, p, time.Now()); err == nil {
		t.Fatal("want error for missing logo")
	}
	if _, err := os.Stat(filepath.Join(root, "assets", "x", "og-image.png")); err == nil {
		t.Fatal("no file must be written on failure")
	}
}
