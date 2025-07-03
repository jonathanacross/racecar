package main

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"strconv"

	"github.com/fogleman/gg"
)

func DrawPoly(dc *gg.Context, poly []gg.Point, fillColor color.Color, strokeColor color.Color) {
	dc.MoveTo(poly[0].X, poly[0].Y)
	for i := 1; i < len(poly); i++ {
		dc.LineTo(poly[i].X, poly[i].Y)
	}
	dc.ClosePath()

	// TODO: for some reason, gg hangs when trying to fill paths.. need to debug
	// fill the path, and save the path
	// r, g, b, a := fillColor.RGBA()
	// dc.SetRGBA(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, float64(a)/255.0)
	// dc.FillPreserve()

	// stroke the path
	r, g, b, a := strokeColor.RGBA()
	dc.SetRGBA(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, float64(a)/255.0)
	dc.SetLineWidth(2)
	dc.Stroke()
}

func toGgPoly(points []Point) []gg.Point {
	result := make([]gg.Point, len(points))
	for i, p := range points {
		result[i] = gg.Point{X: p.X, Y: p.Y}
	}
	return result
}

func DrawToImage(width int, height int, numPoints int, roadWidth float64) {
	margin := math.Min(float64(width), float64(height)) / 10

	bounds := Rect{left: float64(margin), top: float64(margin), right: float64(width) - margin, bottom: float64(height) - margin}

	trackData := BuildPossiblyIntersectingTrack(numPoints, bounds, roadWidth)

	dc := gg.NewContext(width, height)
	dc.FillPreserve()
	dc.SetRGBA(0.2, 0.2, 0.2, 1.0)
	dc.DrawRectangle(0, 0, float64(width), float64(height))
	dc.Stroke()

	red := color.RGBA{255, 0, 0, 255}
	yellow := color.RGBA{255, 255, 0, 255}
	purple := color.RGBA{255, 0, 128, 255}
	darkBlue := color.RGBA{0, 0, 255, 255}
	lightBlue := color.RGBA{0, 128, 255, 255}
	darkGreen := color.RGBA{0, 255, 0, 255}
	lightGreen := color.RGBA{128, 255, 128, 255}

	innerColor := darkGreen
	if IsSelfIntersecting(trackData.inner) {
		innerColor = darkBlue
	}
	outerColor := lightGreen
	if IsSelfIntersecting(trackData.inner) {
		outerColor = lightBlue
	}
	DrawPoly(dc, toGgPoly(trackData.orig), color.RGBA{0, 0, 0, 0}, purple)
	DrawPoly(dc, toGgPoly(trackData.perturbed), color.RGBA{0, 0, 0, 0}, red)
	DrawPoly(dc, toGgPoly(trackData.rounded), color.RGBA{0, 0, 0, 0}, yellow)
	DrawPoly(dc, toGgPoly(trackData.inner), color.RGBA{0, 0, 0, 0}, innerColor)
	DrawPoly(dc, toGgPoly(trackData.outer), color.RGBA{0, 0, 0, 0}, outerColor)

	dc.SavePNG("polygon.png") // Save the drawing to a PNG file
}

func main() {
	args := os.Args[1:]
	if len(args) < 4 {
		fmt.Println("usage: trackgen width height numPoints roadWidth")
		return
	}

	width, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("could not parse width as integer: %v\n", err)
		return
	}

	height, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("could not parse height as integer: %v\n", err)
		return
	}

	numPoints, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("could not parse numPoints as integer: %v\n", err)
		return
	}

	roadWidth, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		fmt.Printf("could not parse roadWidth as float: %v\n", err)
		return
	}

	DrawToImage(width, height, numPoints, roadWidth)
}
