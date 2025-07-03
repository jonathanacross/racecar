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

func norm(vec gg.Point) gg.Point {
	len := math.Sqrt(vec.X*vec.X + vec.Y + vec.Y)
	return gg.Point{
		X: vec.X / len,
		Y: vec.Y / len,
	}
}

func GetNewPath(numPoints int, width float64, height float64) []gg.Point {
	points := make([]gg.Point, numPoints)
	for i := 0; i < numPoints; i++ {
		points[i].X = rand.Float64() * width
		points[i].Y = rand.Float64() * height
	}
	return points
}

func clamp(x float64, lo float64, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func Perturb(ladder []gg.Point, width float64, height float64, roadWidth float64) {
	// compute total force on each vertex
	numPoints := len(ladder)
	forces := make([]gg.Point, numPoints)

	fBending := 0.05
	fLength := 0.05
	fNonAdj := 0.05
	targetLen := 50.0

	// totalLen := 0.0
	// for i := 0; i < numPoints; i++ {
	// 	j := (i + 1) % numPoints
	// 	totalLen += euclideanDistance(ladder[i], ladder[j])
	// }
	// avgLen := totalLen / float64(numPoints)

	for i := 0; i < numPoints; i++ {
		fmt.Printf("i = %v\n", i)
		// move each point toward average of neighbors
		j := (i + 1) % numPoints
		k := (i + 2) % numPoints
		targetLoc := gg.Point{
			X: 0.5 * (ladder[i].X + ladder[k].X),
			Y: 0.5 * (ladder[i].Y + ladder[k].Y),
		}
		forces[j].X += fBending * (targetLoc.X - ladder[j].X)
		forces[j].Y += fBending * (targetLoc.Y - ladder[j].Y)

		// try to make segments the same length
		// j := (i + 1) % numPoints
		dAdj := euclideanDistance(ladder[j], ladder[i])
		fRungInner := fLength * (dAdj - targetLen)
		innerVec := norm(gg.Point{
			X: ladder[j].X - ladder[i].X,
			Y: ladder[j].Y - ladder[i].Y,
		})
		forces[i].X += innerVec.X * fRungInner
		forces[i].Y += innerVec.Y * fRungInner
		forces[j].X -= innerVec.X * fRungInner
		forces[j].Y -= innerVec.Y * fRungInner

		// Try to make sure non-adjacent vertices don't get too close
		for m := 0; m < numPoints; m++ {
			if m == i || m == j || m == k {
				continue
			}
			dNonAdj := euclideanDistance(ladder[j], ladder[m])
			if dNonAdj < 2*roadWidth {
				totalFNonAdj := -fNonAdj * (2*roadWidth - dNonAdj)
				forces[j].X += totalFNonAdj * (ladder[m].X - ladder[j].X)
				forces[j].Y += totalFNonAdj * (ladder[m].Y - ladder[j].Y)
			}
		}
	}

	// apply forces
	margin := 50.0
	for i := 0; i < numPoints; i++ {
		ladder[i].X += 0.1 * forces[i].X
		ladder[i].Y += 0.1 * forces[i].Y

		// ensure stays in bounding box
		ladder[i].X = clamp(ladder[i].X, margin, width-margin)
		ladder[i].Y = clamp(ladder[i].Y, margin, height-margin)
	}
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

func DrawPoly(dc *gg.Context, poly []gg.Point, fillColor color.Color, strokeColor color.Color) {
	fmt.Println("a..")
	dc.MoveTo(poly[0].X, poly[0].Y)
	for i := 1; i < len(poly); i++ {
		dc.LineTo(poly[i].X, poly[i].Y)
	}
	dc.ClosePath()

	// fill the path, save the path
	// fmt.Println("b..")
	// r, g, b, a := fillColor.RGBA()
	// dc.SetRGBA(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, float64(a)/255.0)
	// dc.FillPreserve()

	// stroke the path
	fmt.Println("c..")
	r, g, b, a := strokeColor.RGBA()
	dc.SetRGBA(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, float64(a)/255.0)
	dc.SetLineWidth(2)
	fmt.Println("stroking..")
	dc.Stroke()
}

func DrawToImage(numPoints int, roadWidth float64) {
	width := 600
	height := 600
	//margin := 50

	//trackArea := Rect{left: float64(margin), top: float64(margin), right: float64(width - margin), bottom: float64(height - margin)}

	// cx := float64(width / 2)
	// cy := float64(height / 2)
	// rInner := float64(height / 4)
	// rOuter := float64(height / 3)

	dc := gg.NewContext(width, height)
	dc.FillPreserve()
	dc.SetRGBA(0.2, 0.2, 0.2, 1.0)
	dc.DrawRectangle(0, 0, float64(width), float64(height))
	dc.Stroke()

	ladder := GetNewPath(numPoints, float64(width), float64(height))

	DrawPoly(dc, ladder, color.RGBA{0, 0, 0, 0}, color.RGBA{0, 255, 0, 255})

	for iter := 0; iter < 5; iter++ {
		Perturb(ladder, float64(width), float64(height), roadWidth)
	}
	fmt.Println("finished perturbation")
	DrawPoly(dc, ladder, color.RGBA{0, 0, 0, 0}, color.RGBA{255, 255, 0, 255})
	fmt.Println("saving..")

	dc.SavePNG("ladder.png")
}

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("usage: forces numPoints roadWidth")
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
