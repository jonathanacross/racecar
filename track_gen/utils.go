package main

import (
	"math"
)

type Point struct {
	X float64
	Y float64
}

// Rects are in graphics coordinates, so bottom value is greater than the top value.
type Rect struct {
	left   float64
	top    float64
	right  float64
	bottom float64
}

func Dist(p1 Point, p2 Point) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func Norm(vec Point) Point {
	len := math.Sqrt(vec.X*vec.X + vec.Y*vec.Y)
	if len == 0 {
		return Point{X: 0, Y: 0}
	}
	return Point{
		X: vec.X / len,
		Y: vec.Y / len,
	}
}

func WeightedAverage(a Point, b Point, lambda float64) Point {
	return Point{
		X: (1.0-lambda)*a.X + lambda*b.X,
		Y: (1.0-lambda)*a.Y + lambda*b.Y,
	}
}

func (r *Rect) Width() float64 {
	return r.right - r.left
}

func (r *Rect) Height() float64 {
	return r.bottom - r.top
}

func Clamp(x float64, lo float64, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

// Computes the area of a polygon using the trapezoid formula.
// A = 1/2 * \sum_{i} (y_i + y_{i+1})*(x_i - x_{i+1})
func Area(poly []Point) float64 {
	area := 0.0
	for i, curr := range poly {
		next := poly[(i+1)%len(poly)]
		area += (curr.Y + next.Y) * (curr.X - next.X)
	}
	return 0.5 * area
}

// Reverse a polygon to change its orientation.
func Reverse(poly []Point) {
	i := 0
	j := len(poly) - 1
	for i < j {
		poly[i], poly[j] = poly[j], poly[i]
		i++
		j--
	}
}

// Reorders the vertices of poly, if needed, so that the polygon
// has positive area.
func OrientPositive(poly []Point) {
	if Area(poly) < 0 {
		Reverse(poly)
	}
}

// orientation determines the orientation of an ordered triplet (p, q, r).
// The function returns:
// 0 --> p, q and r are collinear
// 1 --> Clockwise
// -1 --> Counterclockwise
func orientation(p Point, q Point, r Point) int {
	// Calculate the cross product (or determinant)
	// (q.Y - p.Y) * (r.X - q.X) - (q.X - p.X) * (r.Y - q.Y)
	val := (q.Y-p.Y)*(r.X-q.X) - (q.X-p.X)*(r.Y-q.Y)

	if val == 0 {
		return 0 // Collinear
	} else if val > 0 {
		return 1 // Clockwise
	} else {
		return -1 // Counterclockwise
	}
}

// onSegment checks if point q lies on line segment pr.
// This function assumes that p, q, and r are collinear.
func onSegment(p, q, r Point) bool {
	return q.X <= math.Max(p.X, r.X) && q.X >= math.Min(p.X, r.X) &&
		q.Y <= math.Max(p.Y, r.Y) && q.Y >= math.Min(p.Y, r.Y)
}

// SegmentsIntersect returns if two line segments (p1, q1) and (p2, q2) intersect.
func SegmentsIntersect(p1, q1, p2, q2 Point) bool {
	// Find the four orientations needed for general and special cases
	o1 := orientation(p1, q1, p2)
	o2 := orientation(p1, q1, q2)
	o3 := orientation(p2, q2, p1)
	o4 := orientation(p2, q2, q1)

	// If the orientations are different for both pairs of endpoints,
	// then the segments intersect.
	if o1 != 0 && o2 != 0 && o3 != 0 && o4 != 0 &&
		o1 != o2 && o3 != o4 {
		return true
	}

	// p1, q1 and p2 are collinear and p2 lies on segment p1q1
	if o1 == 0 && onSegment(p1, p2, q1) {
		return true
	}

	// p1, q1 and q2 are collinear and q2 lies on segment p1q1
	if o2 == 0 && onSegment(p1, q2, q1) {
		return true
	}

	// p2, q2 and p1 are collinear and p1 lies on segment p2q2
	if o3 == 0 && onSegment(p2, p1, q2) {
		return true
	}

	// p2, q2 and q1 are collinear and q1 lies on segment p2q2
	if o4 == 0 && onSegment(p2, q1, q2) {
		return true
	}

	return false
}

// IsSelfIntersecting checks if a closed polygon self-intersects.
func IsSelfIntersecting(polygon []Point) bool {
	n := len(polygon)
	if n < 3 {
		return false
	}

	// Naive algorithm: just check all possible pairs of segments.
	for i := 0; i < n; i++ {
		// Define the first segment (p1, q1).
		p1 := polygon[i]
		q1 := polygon[(i+1)%n]

		for j := i + 1; j < n; j++ {
			// Define the second segment (p2, q2).
			p2 := polygon[j]
			q2 := polygon[(j+1)%n]

			// Skip adjacent segments.
			if (i == j) || ((i+1)%n == j) || ((j+1)%n == i) {
				continue
			}

			// Check for intersection between the two non-adjacent segments
			if SegmentsIntersect(p1, q1, p2, q2) {
				return true
			}
		}
	}
	return false
}
