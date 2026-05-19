package render

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var (
	bgColor    = color.RGBA{30, 30, 30, 255}
	fgColor    = color.RGBA{204, 204, 204, 255}
	promptColor = color.RGBA{80, 250, 123, 255}
	dimColor   = color.RGBA{100, 100, 100, 255}
	headerBg   = color.RGBA{40, 40, 40, 255}
)

// ProofDir returns the proof output directory
func ProofDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "shellcast", "proofs")
	os.MkdirAll(dir, 0755)
	return dir
}

// GenerateProof renders a command + output as a terminal-style PNG
func GenerateProof(filename, input, cleanOutput string) error {
	// Build lines
	var lines []pngLine
	lines = append(lines, pngLine{"$ " + input, promptColor})
	for _, l := range strings.Split(cleanOutput, "\n") {
		if len(l) > 120 {
			l = l[:117] + "..."
		}
		lines = append(lines, pngLine{l, fgColor})
	}

	// Dimensions
	charW := 7  // basicfont width
	lineH := 18
	padX := 24
	padY := 20
	headerH := 36

	maxW := 0
	for _, l := range lines {
		w := len(l.text) * charW
		if w > maxW {
			maxW = w
		}
	}
	imgW := maxW + padX*2
	if imgW < 600 {
		imgW = 600
	}
	if imgW > 1200 {
		imgW = 1200
	}
	imgH := headerH + len(lines)*lineH + padY*2

	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))

	// Fill background
	fillRect(img, 0, 0, imgW, imgH, bgColor)

	// Header bar (window chrome)
	fillRect(img, 0, 0, imgW, headerH, headerBg)
	// Dots
	drawCircle(img, 18, headerH/2, 6, color.RGBA{255, 95, 86, 255})
	drawCircle(img, 38, headerH/2, 6, color.RGBA{255, 189, 46, 255})
	drawCircle(img, 58, headerH/2, 6, color.RGBA{39, 201, 63, 255})
	// Title
	drawText(img, imgW/2-30, headerH/2+4, "shellcast", dimColor)

	// Draw lines
	face := basicfont.Face7x13
	y := headerH + padY + 13
	for _, l := range lines {
		drawTextFace(img, padX, y, l.text, l.clr, face)
		y += lineH
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

type pngLine struct {
	text string
	clr  color.RGBA
}

func fillRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			img.Set(px, py, c)
		}
	}
}

func drawCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				img.Set(cx+x, cy+y, c)
			}
		}
	}
}

func drawText(img *image.RGBA, x, y int, text string, c color.RGBA) {
	drawTextFace(img, x, y, text, c, basicfont.Face7x13)
}

func drawTextFace(img *image.RGBA, x, y int, text string, c color.RGBA, face font.Face) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(text)
}

// GenerateProofFile is a convenience wrapper
func GenerateProofFile(cmdID int64, input, cleanOutput string) (string, error) {
	dir := ProofDir()
	filename := filepath.Join(dir, fmt.Sprintf("proof_%d.png", cmdID))
	err := GenerateProof(filename, input, cleanOutput)
	if err != nil {
		return "", err
	}
	return filename, nil
}
