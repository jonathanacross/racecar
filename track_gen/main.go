package main

import (
	"fmt"
	"image/color"
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

	// fill the path, and save the path
	r, g, b, a := fillColor.RGBA()
	dc.SetRGBA(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, float64(a)/255.0)
	dc.FillPreserve()

	// stroke the path
	r, g, b, a = strokeColor.RGBA()
	dc.SetRGBA(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, float64(a)/255.0)
	dc.SetLineWidth(2)
	dc.Stroke()
}

func ConvertPolyToGgPoly(points []Point) []gg.Point {
	result := make([]gg.Point, len(points))
	for i, p := range(points) {
		result[i] = gg.Point{X: p.X, Y: p.Y}
	}
	return result
}

func DrawToImage(numPoints int, roadWidth float64) {
	width := 600
	height := 600
	margin := 50

	bounds := Rect{left: float64(margin), top: float64(margin), right: float64(width - margin), bottom: float64(height - margin)}

	inner, outer := BuildTrack(numPoints, bounds, roadWidth)

	innergg := ConvertPolyToGgPoly(inner)
	outergg := ConvertPolyToGgPoly(outer)

	dc := gg.NewContext(width, height)
	dc.FillPreserve()
	dc.SetRGBA(0.2, 0.2, 0.2, 1.0)
	dc.DrawRectangle(0, 0, float64(width), float64(height))
	dc.Stroke()

	DrawPoly(dc, innergg, color.RGBA{0, 0, 0, 0}, color.RGBA{0, 0, 255, 255})
	DrawPoly(dc, outergg, color.RGBA{0, 0, 0, 0}, color.RGBA{0, 128, 255, 255})

	dc.SavePNG("polygon.png") // Save the drawing to a PNG file
}

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("usage: trackgen numPoints roadWidth")
		return
	}

	numPoints, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("could not parse numPoints as integer: %v\n", err)
		return
	}

	roadWidth, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		fmt.Printf("could not parse roadWidth as float: %v\n", err)
		return
	}

	DrawToImage(numPoints, roadWidth)
}
