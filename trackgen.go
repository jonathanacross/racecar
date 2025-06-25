package main

import (
	"fmt"
	"image/color"
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

// Expands a polygon
func Expand(poly []gg.Point, r float64, curvatures []float64) []gg.Point {
	numPoints := len(poly)
	expanded := make([]gg.Point, numPoints)

	// TODO: change index to 0 based
	for i := 1; i <= len(poly); i++ {
		prev := poly[i-1]
		curr := poly[i%numPoints]
		next := poly[(i+1)%numPoints]

		dx := next.X - prev.X
		dy := next.Y - prev.Y
		d := math.Sqrt(dx*dx + dy*dy)

		// Logic for curvatures
		radius := 1.0 / curvatures[i%numPoints]
		minRoadWidth := r
		if (radius > 0 && r > 0 && radius < r) ||
			(radius < 0 && r < 0 && radius > r) {
			minRoadWidth = radius
		}

		// logic for maxDists..
		// minRoadWidth := r
		// maxDist := curvatures[i%numPoints]
		// if r > 0 && maxDist < r {
		// 	minRoadWidth = maxDist
		// } else if r < 0 && maxDist > -r {
		// 	minRoadWidth = -maxDist
		// }

		expanded[i%numPoints] = gg.Point{
			X: curr.X - (dy * minRoadWidth / d),
			Y: curr.Y + (dx * minRoadWidth / d),
		}
	}

	return expanded
}

func GetCurvature(poly []gg.Point) []float64 {
	numPoints := len(poly)
	curvatures := make([]float64, numPoints)

	// For a function r(t) = (x(t), y(t), z(t))
	// the curvature is (r' x r'') / |r'|^3

	for i := 0; i < len(poly); i++ {
		// Use 5 consecutive points to estimate r', r''
		prev2 := poly[i]
		prev := poly[(i+1)%numPoints]
		curr := poly[(i+2)%numPoints]
		next := poly[(i+3)%numPoints]
		next2 := poly[(i+4)%numPoints]

		rPrime := gg.Point{
			X: 0.5 * (next.X - prev.X),
			Y: 0.5 * (next.Y - prev.Y),
		}

		// Estimate second derivative by differences of differences
		rPrimePrime := gg.Point{
			X: 0.25 * ((next2.X - curr.X) - (curr.X - prev2.X)),
			Y: 0.25 * ((next2.Y - curr.Y) - (curr.Y - prev2.Y)),
		}

		// Since r', r'' are 2D vectors, we don't need a full 3x3 determinant,
		// just the 2x2 determinant of the x, y components.
		cross := rPrime.X*rPrimePrime.Y - rPrime.Y*rPrimePrime.X

		rPrimeNorm := math.Sqrt(rPrime.X*rPrime.X + rPrime.Y*rPrime.Y)

		curvature := cross / (rPrimeNorm * rPrimeNorm * rPrimeNorm)

		curvatures[(i+2)%numPoints] = curvature
	}

	return curvatures
}

func GetMaxRadius(poly []gg.Point) []float64 {
	numPoints := len(poly)
	dists := make([]float64, numPoints)

	for i := 0; i < numPoints; i++ {
		// find the nearest point to this one.
		minDist := math.MaxFloat64
		for j := 0; j < numPoints; j++ {
			if (i == j) || (i == j-1) || (i == j+1) { //  TOOD: won't handle cyclick
				continue
			}
			dist := euclideanDistance(poly[i], poly[j])
			if dist < minDist {
				minDist = dist
			}
		}
		dists[i] = minDist / 3
	}

	return dists
}

func GetTrackSkeleton2(numPoints int, boundsMinX, boundsMinY, boundsMaxX, boundsMaxY float64) []gg.Point {
	points := []gg.Point{}

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

func DrawPoly(dc *gg.Context, poly []gg.Point, fillColor color.Color, strokeColor color.Color) {
	dc.MoveTo(poly[0].X, poly[0].Y)
	for i := 1; i < len(poly); i++ {
		dc.LineTo(poly[i].X, poly[i].Y)
	}
	dc.ClosePath()

	// fill the path, save the path
	r, g, b, a := fillColor.RGBA()
	dc.SetRGBA(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, float64(a)/255.0)
	dc.FillPreserve()

	// stroke the path
	r, g, b, a = strokeColor.RGBA()
	dc.SetRGBA(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, float64(a)/255.0)
	dc.SetLineWidth(2)
	dc.Stroke()
}

func DrawToImage(numPoints int, roadWidth float64) {
	width := 600
	height := 600
	margin := 50

	trackArea := Rect{left: float64(margin), top: float64(margin), right: float64(width - margin), bottom: float64(height - margin)}

	points := GetTrackSkeleton2(numPoints, trackArea.left, trackArea.top, trackArea.right, trackArea.bottom)
	rescaledPoints := rescale(points, trackArea)
	rounded := smoothCorners(rescaledPoints)
	rounded = smoothCorners(rounded)
	//rounded = smoothCorners(rounded)

	//maxDists := GetMaxRadius(rounded)
	curvatures := GetCurvature(rounded)
	for _, k := range curvatures {
		fmt.Println("%v", 1.0/k)
	}
	expanded := Expand(rounded, roadWidth, curvatures)
	expanded2 := Expand(rounded, -roadWidth, curvatures)

	dc := gg.NewContext(width, height)
	dc.FillPreserve()
	dc.SetRGBA(0.2, 0.2, 0.2, 1.0)
	dc.DrawRectangle(0, 0, float64(width), float64(height))
	dc.Stroke()

	//DrawPoly(dc, rescaledPoints, color.RGBA{0, 0, 0, 0}, color.RGBA{255, 0, 0, 255})
	DrawPoly(dc, rounded, color.RGBA{0, 0, 0, 0}, color.RGBA{0, 255, 0, 255})
	DrawPoly(dc, expanded, color.RGBA{0, 0, 0, 0}, color.RGBA{0, 0, 255, 255})
	// for i, k := range curvatures {
	// 	radius := 1.0 / k
	// 	point := expanded[i]
	// 	if i%5 == 0 && radius > 0 {
	// 		//if math.Abs(radius) < 150 {
	// 		dc.SetRGBA(1.0, 1.0, 0.0, 1.0)
	// 		dc.DrawCircle(point.X, point.Y, radius)
	// 		dc.Stroke()
	// 	}
	// }
	DrawPoly(dc, expanded2, color.RGBA{0, 0, 0, 0}, color.RGBA{0, 128, 255, 255})

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
