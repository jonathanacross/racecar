package main

import (
	"github.com/fogleman/gg"
)

// Vector represents a 2D point or vector using floating-point coordinates.
// Ebitengine's drawing functions often work with float64 for position.
type Vector struct {
	X float64
	Y float64
}

// Track defines the layout of a racetrack, consisting of outer and inner boundaries
// as polygons, and a finish line segment.
type Track struct {
	// OuterBounds is a slice of Vectors defining the vertices of the outer boundary of the track.
	// The last point should typically connect back to the first to form a closed polygon.
	OuterBounds []Vector

	// InnerBounds is a slice of Vectors defining the vertices of the inner boundary of the track.
	// This represents the hole in the track. The last point should connect back to the first.
	InnerBounds []Vector

	// FinishLineStart and FinishLineEnd define the two endpoints of the finish line segment.
	FinishLineStart Vector
	FinishLineEnd   Vector
}

// NewRectangularTrack creates and returns a simple rectangular track for demonstration purposes.
// This function can be called to initialize a default track easily.
func NewRectangularTrack(width, height float64) *Track {
	// Define padding from the edge of the logical screen
	const padding = 50.0
	const trackWidth = 100.0 // The width of the track itself

	outerX1 := padding
	outerY1 := padding
	outerX2 := width - padding
	outerY2 := height - padding

	innerX1 := outerX1 + trackWidth
	innerY1 := outerY1 + trackWidth
	innerX2 := outerX2 - trackWidth
	innerY2 := outerY2 - trackWidth

	return &Track{
		OuterBounds: []Vector{
			{X: outerX1, Y: outerY1}, // Top-left
			{X: outerX2, Y: outerY1}, // Top-right
			{X: outerX2, Y: outerY2}, // Bottom-right
			{X: outerX1, Y: outerY2}, // Bottom-left
			{X: outerX1, Y: outerY1}, // Close the polygon
		},
		InnerBounds: []Vector{
			{X: innerX1, Y: innerY1}, // Top-left
			{X: innerX2, Y: innerY1}, // Top-right
			{X: innerX2, Y: innerY2}, // Bottom-right
			{X: innerX1, Y: innerY2}, // Bottom-left
			{X: innerX1, Y: innerY1}, // Close the polygon
		},
		// Example finish line, crossing from left inner to left outer
		FinishLineStart: Vector{X: innerX1, Y: height / 2},
		FinishLineEnd:   Vector{X: outerX1, Y: height / 2},
	}
}

func DrawToImage() {
	width := 400
	height := 300

	points := []gg.Point{
        {X: 100, Y: 100},
        {X: 200, Y: 150},
        {X: 150, Y: 250},
        {X: 50, Y: 200},
    }

    dc := gg.NewContext(width, height) // Replace width and height with your desired dimensions

	if len(points) > 0 {
        dc.MoveTo(points[0].X, points[0].Y)
        for i := 1; i < len(points); i++ {
            dc.LineTo(points[i].X, points[i].Y)
        }
        dc.ClosePath() // Closes the path to form a complete polygon
    }

	dc.SetRGB(0.2, 0.4, 0.8) // Set fill color (e.g., blue)
    dc.FillPreserve()        // Fill the path
    dc.SetRGBA(0, 0, 0, 1)   // Set stroke color (e.g., black)
    dc.SetLineWidth(2)       // Set line width
    dc.Stroke()              // Stroke the path

	dc.SavePNG("polygon.png") // Save the drawing to a PNG file
}

func main() {
	DrawToImage()
}
