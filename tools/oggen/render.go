package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	canvasW        = 2400
	canvasH        = 1260
	marginX        = 180
	blockGap       = 64
	logoHeight     = 160
	iconHeight     = 1200
	iconOverhang   = 80
	accentBarH     = 8
	textMaxWidth   = 1200
	watermarkAlpha = 15
)

type faces struct {
	kicker, tagline, body, small font.Face
}

func loadFaces() (*faces, error) {
	regular, err := opentype.Parse(gomono.TTF)
	if err != nil {
		return nil, err
	}
	bold, err := opentype.Parse(gomonobold.TTF)
	if err != nil {
		return nil, err
	}
	newFace := func(f *opentype.Font, size float64) (font.Face, error) {
		return opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	}
	var fs faces
	if fs.kicker, err = newFace(regular, 56); err != nil {
		return nil, err
	}
	if fs.tagline, err = newFace(bold, 96); err != nil {
		return nil, err
	}
	if fs.body, err = newFace(regular, 44); err != nil {
		return nil, err
	}
	if fs.small, err = newFace(regular, 32); err != nil {
		return nil, err
	}
	return &fs, nil
}

type textBlock struct {
	lines []string
	face  font.Face
	col   color.NRGBA
}

func (b textBlock) height() int {
	return len(b.lines) * lineHeight(b.face)
}

func lineHeight(f font.Face) int {
	return f.Metrics().Height.Ceil()
}

type variant struct {
	texts Texts
	file  string
}

// variants lists the images to produce: og-image.png for the default texts,
// plus one og-image-<lang>.png per declared locale.
func (p Project) variants() []variant {
	if len(p.Locales) == 0 {
		return []variant{{Texts{Type: p.Type, Tagline: p.Tagline, Description: p.Description}, "og-image.png"}}
	}
	vs := []variant{{p.Locales[p.DefaultLocale], "og-image.png"}}
	langs := make([]string, 0, len(p.Locales))
	for lang := range p.Locales {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	for _, lang := range langs {
		vs = append(vs, variant{p.Locales[lang], "og-image-" + lang + ".png"})
	}
	return vs
}

func Render(root string, p Project, now time.Time) ([]string, error) {
	bg, err := parseHexColor(p.Background)
	if err != nil {
		return nil, err
	}
	logo, err := loadPNG(filepath.Join(root, p.LogoFull))
	if err != nil {
		return nil, fmt.Errorf("logoFull: %w", err)
	}
	logo = trimTransparent(logo)
	icon, err := loadPNG(filepath.Join(root, p.LogoIcon))
	if err != nil {
		return nil, fmt.Errorf("logoIcon: %w", err)
	}
	icon = trimTransparent(icon)
	fs, err := loadFaces()
	if err != nil {
		return nil, err
	}

	var outs []string
	for _, v := range p.variants() {
		out, err := renderOne(root, p, v, bg, logo, icon, fs, now)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", v.file, err)
		}
		outs = append(outs, out)
	}
	return outs, nil
}

func renderOne(root string, p Project, v variant, bg color.NRGBA, logo, icon image.Image, fs *faces, now time.Time) (string, error) {
	canvas := image.NewNRGBA(image.Rect(0, 0, canvasW, canvasH))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	iconScaled := scaleToHeight(icon, iconHeight)
	iconLeft := canvasW - iconScaled.Bounds().Dx() + iconOverhang
	iconTop := (canvasH - iconHeight) / 2
	iconRect := image.Rect(iconLeft, iconTop, iconLeft+iconScaled.Bounds().Dx(), iconTop+iconHeight)
	draw.DrawMask(canvas, iconRect, iconScaled, image.Point{},
		&image.Uniform{color.Alpha{watermarkAlpha}}, image.Point{}, draw.Over)

	if err := drawAccentBar(canvas, p.Accent); err != nil {
		return "", err
	}

	white := color.NRGBA{255, 255, 255, 255}
	dim60 := color.NRGBA{255, 255, 255, 153}
	dim50 := color.NRGBA{255, 255, 255, 128}
	dim40 := color.NRGBA{255, 255, 255, 102}

	var blocks []textBlock
	if v.texts.Type != "" {
		blocks = append(blocks, textBlock{[]string{strings.ToUpper(v.texts.Type)}, fs.kicker, dim50})
	}
	blocks = append(blocks, textBlock{strings.Split(v.texts.Tagline, "\n"), fs.tagline, white})
	if v.texts.Description != "" {
		blocks = append(blocks, textBlock{wrapText(v.texts.Description, fs.body, textMaxWidth), fs.body, dim60})
	}
	blocks = append(blocks, textBlock{[]string{p.URL}, fs.body, white})
	blocks = append(blocks, textBlock{[]string{now.Format("2006-01-02")}, fs.small, dim40})

	total := logoHeight + blockGap*len(blocks)
	for _, b := range blocks {
		total += b.height()
	}

	y := (canvasH - total) / 2
	logoScaled := scaleToHeight(logo, logoHeight)
	logoRect := image.Rect(marginX, y, marginX+logoScaled.Bounds().Dx(), y+logoHeight)
	draw.Draw(canvas, logoRect, logoScaled, image.Point{}, draw.Over)
	y += logoHeight + blockGap

	for _, b := range blocks {
		ascent := b.face.Metrics().Ascent.Ceil()
		for _, line := range b.lines {
			d := font.Drawer{
				Dst:  canvas,
				Src:  &image.Uniform{b.col},
				Face: b.face,
				Dot:  fixed.P(marginX, y+ascent),
			}
			d.DrawString(line)
			y += lineHeight(b.face)
		}
		y += blockGap
	}

	out := filepath.Join(root, "public", p.Name, v.file)
	if err := os.MkdirAll(filepath.Dir(out), 0o750); err != nil {
		return "", err
	}
	f, err := os.Create(out) // #nosec G304 -- output path built from the project name in og.config.json
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := png.Encode(f, canvas); err != nil {
		return "", err
	}
	return out, nil
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path) // #nosec G304 -- logo paths come from og.config.json in this repo
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func trimTransparent(src image.Image) image.Image {
	const alphaMin = 0x1000
	b := src.Bounds()
	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X-1, b.Min.Y-1
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := src.At(x, y).RGBA(); a >= alphaMin {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if maxX < minX {
		return src
	}
	out := image.NewNRGBA(image.Rect(0, 0, maxX-minX+1, maxY-minY+1))
	draw.Draw(out, out.Bounds(), src, image.Point{X: minX, Y: minY}, draw.Src)
	return out
}

func scaleToHeight(src image.Image, h int) *image.NRGBA {
	b := src.Bounds()
	w := b.Dx() * h / b.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
	return dst
}

func drawAccentBar(canvas *image.NRGBA, hexes []string) error {
	colors := make([]color.NRGBA, len(hexes))
	for i, h := range hexes {
		c, err := parseHexColor(h)
		if err != nil {
			return err
		}
		colors[i] = c
	}
	for x := 0; x < canvasW; x++ {
		c := gradientAt(colors, float64(x)/float64(canvasW-1))
		for y := 0; y < accentBarH; y++ {
			canvas.SetNRGBA(x, y, c)
		}
	}
	return nil
}

func gradientAt(cs []color.NRGBA, t float64) color.NRGBA {
	if len(cs) == 1 {
		return cs[0]
	}
	seg := t * float64(len(cs)-1)
	i := int(seg)
	if i >= len(cs)-1 {
		i = len(cs) - 2
	}
	f := seg - float64(i)
	lerp := func(a, b uint8) uint8 {
		return uint8(float64(a) + (float64(b)-float64(a))*f + 0.5)
	}
	return color.NRGBA{
		R: lerp(cs[i].R, cs[i+1].R),
		G: lerp(cs[i].G, cs[i+1].G),
		B: lerp(cs[i].B, cs[i+1].B),
		A: 255,
	}
}

func wrapText(s string, face font.Face, maxWidth int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		candidate := line + " " + w
		if font.MeasureString(face, candidate).Ceil() > maxWidth {
			lines = append(lines, line)
			line = w
		} else {
			line = candidate
		}
	}
	return append(lines, line)
}
