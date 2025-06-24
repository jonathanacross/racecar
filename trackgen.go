package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"strconv"

	"github.com/fogleman/gg"
)

// Rects are in graphics coordinates, so bottom value is greater than the top value.
type Rect struct {
	left   float64
	top    float64
	right  float64
	bottom float64
}

func (r *Rect) Width() float64 {
	return r.right - r.left
}

func (r *Rect) Height() float64 {
	return r.bottom - r.top
}

func euclideanDistance(p1, p2 gg.Point) float64 { // Fixed: Changed float6 4 to float64
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func GetTrackSkeleton2(numPoints int, boundsMinX, boundsMinY, boundsMaxX, boundsMaxY float64) []gg.Point {
	points := make([]gg.Point, numPoints)

	// Determine an approximate 'minDistance' based on the desired number of points and the area.
	area := (boundsMaxX - boundsMinX) * (boundsMaxY - boundsMinY)
	minDistance := math.Sqrt(area / (float64(numPoints) * math.Pi))

	for len(points) < numPoints { // Loop until we have enough points
		candidateX := boundsMinX + rand.Float64()*(boundsMaxX-boundsMinX)
		candidateY := boundsMinY + rand.Float64()*(boundsMaxY-boundsMinY)
		candidate := gg.Point{X: candidateX, Y: candidateY}

		isTooClose := false
		// Check distance from this candidate to all previously accepted points.
		for _, existingPoint := range points {
			dist := euclideanDistance(existingPoint, candidate)
			if dist < minDistance {
				isTooClose = true
				break // Candidate is too close, reject and try a new one
			}
		}

		if !isTooClose {
			// If the candidate is not too close to any existing point, add it.
			points = append(points, candidate)
		}
	}

	// for i := 0; i < numPoints; i++ {
	// 	valid := false
	// 	for !valid {
	// 		valid = true
	// 		x := boundsMinX + rand.Float64()*(boundsMaxX-boundsMinX)
	// 		y := boundsMinY + rand.Float64()*(boundsMaxY-boundsMinY)
	// 		candidate := gg.Point{X: x, Y: y}
	// 		// check distance from this point to the other ones; reject candidate if too close
	// 		for j := 1; j < i; j++ {
	// 			dist := euclideanDistance(points[j], candidate)
	// 			if dist < minDistance {
	// 				valid = false
	// 				break
	// 			}
	// 		}
	// 		if valid {
	// 			points[i] = candidate
	// 		}
	// 	}
	// }

	return GetShortestCycle(points)
}

func getBoundingBox(points []gg.Point) Rect {
	minX, minY := points[0].X, points[0].Y
	maxX, maxY := points[0].X, points[0].Y

	for _, p := range points {
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	return Rect{left: minX, top: minY, right: maxX, bottom: maxY}
}

// Rescales a set of points to fit in a new rectangle
func rescale(points []gg.Point, targetRect Rect) []gg.Point {
	srcRect := getBoundingBox(points)

	srcWidth := srcRect.Width()
	srcHeight := srcRect.Height()
	targetWidth := targetRect.Width()
	targetHeight := targetRect.Height()

	// Paranoia: handle the edge case where all points have the same coordinates
	if srcWidth == 0 {
		srcWidth = 1.0
	}
	if srcHeight == 0 {
		srcHeight = 1.0
	}

	scaledPoints := make([]gg.Point, len(points))

	for i, p := range points {
		scaledPoints[i] = gg.Point{
			X: targetRect.left + (p.X-srcRect.left)*targetWidth/srcWidth,
			Y: targetRect.top + (p.Y-srcRect.top)*targetHeight/srcHeight,
		}
	}

	return scaledPoints
}

// Computes the weighted average point ( (1 - lambda) * a + lambda*b)
func weightedAverage(a gg.Point, b gg.Point, lambda float64) gg.Point {
	return gg.Point{
		X: (1.0-lambda)*a.X + lambda*b.X,
		Y: (1.0-lambda)*a.Y + lambda*b.Y,
	}
}

// Takes an existing polygon.  Subdivides each side into 3, and truncates
// all the corners.  The resulting polygon will have double the number of sides.
func smoothCorners(points []gg.Point) []gg.Point {
	smoothed := make([]gg.Point, 2*len(points))

	for i := 0; i < len(points); i++ {
		j := (i + 1) % len(points)

		p1 := weightedAverage(points[i], points[j], 0.25)
		p2 := weightedAverage(points[i], points[j], 0.75)

		smoothed[2*i] = p1
		smoothed[2*i+1] = p2
	}

	return smoothed
}

func DrawToImage(gridX int, gridY int) {
	width := 600
	height := 600
	margin := 50

	trackArea := Rect{left: float64(margin), top: float64(margin), right: float64(width - margin), bottom: float64(height - margin)}

	points := GetTrackSkeleton2(gridX*gridY, trackArea.left, trackArea.top, trackArea.right, trackArea.bottom)
	rescaledPoints := rescale(points, trackArea)
	rounded := smoothCorners(rescaledPoints)
	rounded = smoothCorners(rounded)
	rounded = smoothCorners(rounded)

	dc := gg.NewContext(width, height)
	dc.FillPreserve()
	dc.SetRGBA(0.2, 0.2, 0.2, 1.0)
	dc.DrawRectangle(0, 0, float64(width), float64(height))
	dc.Stroke()

	dc.MoveTo(rounded[0].X, rounded[0].Y)
	for i := 1; i < len(rounded); i++ {
		dc.LineTo(rounded[i].X, rounded[i].Y)
	}
	dc.ClosePath()

	dc.SetRGB(0.2, 0.4, 0.8) // Set fill color (e.g., blue)
	dc.FillPreserve()        // Fill the path
	dc.SetRGBA(0, 0, 0, 1)   // Set stroke color (e.g., black)
	dc.SetLineWidth(2)       // Set line width
	dc.Stroke()              // Stroke the path

	// draw original polygon
	// dc.MoveTo(rescaledPoints[0].X, rescaledPoints[0].Y)
	// for i := 1; i < len(rescaledPoints); i++ {
	// 	dc.LineTo(rescaledPoints[i].X, rescaledPoints[i].Y)
	// }
	// dc.ClosePath()

	// dc.SetRGBA(0.0, 0.0, 0.0, 0.0)
	// dc.FillPreserve()        // Fill the path
	// dc.SetRGBA(0, 1.0, 0, 1) // Set stroke color (e.g., black)
	// dc.SetLineWidth(2)       // Set line width
	// dc.Stroke()              // Stroke the path

	dc.SavePNG("polygon.png") // Save the drawing to a PNG file
}

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("usage: trackgen x y")
		return
	}

	gridX, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("could not parse x as integer: %v\n", err)
		return
	}

	gridY, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("could not parse y as integer: %v\n", err)
		return
	}

	DrawToImage(gridX, gridY)
}
