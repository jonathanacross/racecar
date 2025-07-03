package main

import (
	"math"
	"math/rand/v2"
)

// Expands a polygon
// TODO: consider changing algorithm to standard polygon expansion
// so that sides are parallel.
func expand(poly []Point, r float64) []Point {
	numPoints := len(poly)
	expanded := make([]Point, numPoints)

	// TODO: change index to 0 based
	for i := 1; i <= len(poly); i++ {
		prev := poly[i-1]
		curr := poly[i%numPoints]
		next := poly[(i+1)%numPoints]

		dx := next.X - prev.X
		dy := next.Y - prev.Y
		d := math.Sqrt(dx*dx + dy*dy)

		expanded[i%numPoints] = Point{
			X: curr.X - (dy * r / d),
			Y: curr.Y + (dx * r / d),
		}
	}

	return expanded
}

// getTrackSkeleton generates a random polygon with numPoints points
// lying within bounds.  The polygon is suitable to use as an initial
// skeleton for where the road will follow (e.g., no self-intersections,
// points are not too close to each other).
// TODO: refactor the point generation code to a separate function.
func getTrackSkeleton(numPoints int, bounds Rect) []Point {
	points := []Point{}

	// Determine an approximate 'minDistance' based on the desired number of points and the area.
	area := bounds.Width() * bounds.Height()
	minDistance := math.Sqrt(area / (float64(numPoints) * math.Pi))

	for len(points) < numPoints {
		candidateX := bounds.left + rand.Float64()*bounds.Width()
		candidateY := bounds.top + rand.Float64()*bounds.Height()
		candidate := Point{X: candidateX, Y: candidateY}

		isTooClose := false
		// Check distance from this candidate to all previously accepted points.
		for _, existingPoint := range points {
			dist := Dist(existingPoint, candidate)
			if dist < minDistance {
				isTooClose = true
				break
			}
		}

		if !isTooClose {
			// If the candidate is not too close to any existing point, add it.
			points = append(points, candidate)
		}
	}

	return GetShortestCycle(points)
}

func perturb(ladder []Point, width float64, height float64, roadWidth float64) {
	// Compute total force on each vertex.
	numPoints := len(ladder)
	forces := make([]Point, numPoints)

	fBending := 0.1
	fLength := 0.05
	fNonAdj := 0.005
	targetLen := 50.0

	for i := 0; i < numPoints; i++ {
		// Move each point toward average of neighbors.
		j := (i + 1) % numPoints
		k := (i + 2) % numPoints
		targetLoc := Point{
			X: 0.5 * (ladder[i].X + ladder[k].X),
			Y: 0.5 * (ladder[i].Y + ladder[k].Y),
		}
		forces[j].X += fBending * (targetLoc.X - ladder[j].X)
		forces[j].Y += fBending * (targetLoc.Y - ladder[j].Y)

		// try to make segments the same length
		//j := (i + 1) % numPoints
		dAdj := Dist(ladder[j], ladder[i])
		fRungInner := fLength * (dAdj - targetLen)
		innerVec := Norm(Point{
			X: ladder[j].X - ladder[i].X,
			Y: ladder[j].Y - ladder[i].Y,
		})
		forces[i].X += innerVec.X * fRungInner
		forces[i].Y += innerVec.Y * fRungInner
		forces[j].X -= innerVec.X * fRungInner
		forces[j].Y -= innerVec.Y * fRungInner

		// Try to make sure non-adjacent vertices don't get too close
		// TODO: this should really be looking at closest point to each segment, not to each point.
		for m := 0; m < numPoints; m++ {
			if m == i || m == j || m == k {
				continue
			}
			dNonAdj := Dist(ladder[j], ladder[m])
			// TODO: roadwidth here really means half-road widths.  update naming
			if dNonAdj < 3*roadWidth {
				totalFNonAdj := -fNonAdj * (3*roadWidth - dNonAdj)
				forces[j].X += totalFNonAdj * (ladder[m].X - ladder[j].X)
				forces[j].Y += totalFNonAdj * (ladder[m].Y - ladder[j].Y)
			}
		}
	}

	// apply forces
	margin := 50.0
	for i := 0; i < numPoints; i++ {
		ladder[i].X += forces[i].X
		ladder[i].Y += forces[i].Y

		// ensure stays in bounding box
		ladder[i].X = Clamp(ladder[i].X, margin, width-margin)
		ladder[i].Y = Clamp(ladder[i].Y, margin, height-margin)
	}
}

func getBoundingBox(points []Point) Rect {
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
// TODO: include track width in computation
func rescale(points []Point, targetRect Rect) []Point {
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

	scaledPoints := make([]Point, len(points))

	for i, p := range points {
		scaledPoints[i] = Point{
			X: targetRect.left + (p.X-srcRect.left)*targetWidth/srcWidth,
			Y: targetRect.top + (p.Y-srcRect.top)*targetHeight/srcHeight,
		}
	}

	return scaledPoints
}

// Takes an existing polygon.  Subdivides each side into 3, and truncates
// all the corners.  The resulting polygon will have double the number of sides.
func smoothCorners(points []Point) []Point {
	smoothed := make([]Point, 2*len(points))

	for i := 0; i < len(points); i++ {
		j := (i + 1) % len(points)

		p1 := WeightedAverage(points[i], points[j], 0.25)
		p2 := WeightedAverage(points[i], points[j], 0.75)

		smoothed[2*i] = p1
		smoothed[2*i+1] = p2
	}

	return smoothed
}

func BuildTrack(numPoints int, bounds Rect, roadWidth float64) (inner []Point, outer []Point) {
	points := getTrackSkeleton(numPoints, bounds)
	rescaledPointsOrig := rescale(points, bounds)
	rescaledPoints := make([]Point, len(rescaledPointsOrig))
	copy(rescaledPoints, rescaledPointsOrig)

	// Perturb the points so that after expanding, there is less likelihood of
	// self-intersections.
	// TODO: add intersection check and redo if there are inersections at the end
	for range 20 {
		perturb(rescaledPoints, bounds.Width(), bounds.Height(), roadWidth)
	}
	rounded := smoothCorners(rescaledPoints)

	// TODO: need to orient poly to determine which is inner and which is outer.
	inner = expand(rounded, roadWidth)
	outer = expand(rounded, -roadWidth)
	return
}
