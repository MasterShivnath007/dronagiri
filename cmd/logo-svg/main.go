// Command logo-svg renders the braille emblem as an SVG of dots.
//
// It exists because browsers cannot be trusted with braille glyph widths:
// GitHub's monospace stack has no Braille coverage, so the browser substitutes
// fonts mid-line and the artwork shears. A dot-matrix drawing has no such
// problem and matches the original painting's halftone style.
package main

import (
	"fmt"
	"os"

	"github.com/MasterShivnath007/dronagiri/internal/ui"
)

const (
	spacing = 7.0
	radius  = 2.3
	margin  = 28.0
	colour  = "#F97316"
)

// brailleDots maps a braille bit to its cell position. The Unicode Braille
// Patterns block lays dots out as two columns of four rows.
var brailleDots = [8][2]int{
	{0, 0}, // U+2801 dot 1
	{0, 1}, // dot 2
	{0, 2}, // dot 3
	{1, 0}, // dot 4
	{1, 1}, // dot 5
	{1, 2}, // dot 6
	{0, 3}, // dot 7
	{1, 3}, // dot 8
}

func main() {
	path := "docs/logo.svg"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	lines := ui.LogoLines(120)
	if len(lines) == 0 {
		fmt.Fprintln(os.Stderr, "no logo available")
		os.Exit(1)
	}

	type point struct{ x, y int }
	var dots []point
	maxX, maxY := 0, 0

	for row, line := range lines {
		col := 0
		for _, r := range line {
			if r < 0x2800 || r > 0x28FF {
				col++
				continue
			}
			mask := int(r - 0x2800)
			for bit := 0; bit < 8; bit++ {
				if mask&(1<<bit) == 0 {
					continue
				}
				x := col*2 + brailleDots[bit][0]
				y := row*4 + brailleDots[bit][1]
				dots = append(dots, point{x, y})
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
			col++
		}
	}

	width := float64(maxX+1)*spacing + margin*2
	height := float64(maxY+1)*spacing + margin*2

	var b []byte
	add := func(format string, args ...any) {
		b = append(b, fmt.Sprintf(format, args...)...)
	}

	add(`<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" role="img" aria-label="Hanuman flying, carrying the Dronagiri mountain">`, width, height, width, height)
	add("\n<title>Dronagiri package manager</title>\n")
	add(`<rect width="100%%" height="100%%" fill="#0B1220"/>` + "\n")
	add(`<g fill="%s">`+"\n", colour)

	for _, d := range dots {
		cx := margin + float64(d.x)*spacing
		cy := margin + float64(d.y)*spacing
		add(`<circle cx="%.1f" cy="%.1f" r="%.1f"/>`+"\n", cx, cy, radius)
	}
	add("</g>\n</svg>\n")

	if err := os.WriteFile(path, b, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "could not write", path, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s — %d dots, %.0fx%.0f px\n", path, len(dots), width, height)
}
