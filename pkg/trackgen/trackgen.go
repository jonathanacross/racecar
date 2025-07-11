package trackgen

import (
	"math"
	"math/rand/v2"
)

// Expands a polygon
func expand(poly []Point, r float64) []Point {
	numPoints := len(poly)
	expanded := make([]Point, numPoints)

	for i := 0; i < len(poly); i++ {
		prev := poly[i]
		curr := poly[(i+1)%numPoints]
		next := poly[(i+2)%numPoints]

		a := Norm(Point{
			X: curr.X - prev.X,
			Y: curr.Y - prev.Y,
		})
		b := Norm(Point{
			X: next.X - curr.X,
			Y: next.Y - curr.Y,
		})
		dir := Norm(Point{
			X: -a.Y - b.Y,
			Y: a.X + b.X,
		})

		mag := 0.5 * r * Len(Point{
			X: a.X + b.X,
			Y: a.Y + b.Y,
		})

		expanded[(i+1)%numPoints] = Point{
			X: curr.X + dir.X*mag,
			Y: curr.Y + dir.Y*mag,
		}
	}

	return expanded
}

// getPointsWithPoissonDiscSampling generates numPoints random points
// within bounds.  This uses Poisson disc sampling to ensure that
// points do not lie too close to each other.
func getPointsWithPoissonDiscSampling(numPoints int, bounds Rect) []Point {
	points := []Point{}

	// Determine an approximate 'minDistance' based on the desired number of points and the area.
	area := bounds.Width() * bounds.Height()
	minDistance := math.Sqrt(area / (float64(numPoints) * math.Pi))

	for len(points) < numPoints {
		candidateX := bounds.Left + rand.Float64()*bounds.Width()
		candidateY := bounds.Top + rand.Float64()*bounds.Height()
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
	return points
}

// getTrackSkeleton generates a random polygon with numPoints points
// lying within bounds.  The polygon is suitable to use as an initial
// skeleton for a road.
func getTrackSkeleton(numPoints int, bounds Rect) []Point {
	points := getPointsWithPoissonDiscSampling(numPoints, bounds)
	cycle := GetShortestCycle(points)
	OrientPositive(cycle)
	return cycle
}

func perturb(ladder []Point, bounds Rect, roadWidth float64) {
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

	// Apply forces.
	for i := 0; i < numPoints; i++ {
		ladder[i].X += forces[i].X
		ladder[i].Y += forces[i].Y

		// Ensure path stays in the bounding box.  Include some buffer
		// so that after expanding, the final road will be within bounds.
		ladder[i].X = Clamp(ladder[i].X, bounds.Left+roadWidth, bounds.Right-roadWidth)
		ladder[i].Y = Clamp(ladder[i].Y, bounds.Top+roadWidth, bounds.Bottom-roadWidth)
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
	return Rect{Left: minX, Top: minY, Right: maxX, Bottom: maxY}
}

// Rescales a set of points to fit in a new rectangle
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
			X: targetRect.Left + (p.X-srcRect.Left)*targetWidth/srcWidth,
			Y: targetRect.Top + (p.Y-srcRect.Top)*targetHeight/srcHeight,
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

type TrackDebugData struct {
	Inner     []Point
	Outer     []Point
	Orig      []Point
	Perturbed []Point
	Rounded   []Point
}

func BuildPossiblyIntersectingTrack(numPoints int, bounds Rect, roadWidth float64) TrackDebugData {
	points := getTrackSkeleton(numPoints, bounds)
	rescaledPointsOrig := rescale(points, bounds)
	rescaledPoints := make([]Point, len(rescaledPointsOrig))
	copy(rescaledPoints, rescaledPointsOrig)

	// Perturb the points so that after expanding, there is less likelihood of
	// self-intersections.
	for range 20 {
		perturb(rescaledPoints, bounds, roadWidth)
	}

	// TODO: enable after debugging
	// // Rescale the perturbed points so that they fill the bounding box
	// // (since perturbation tends to shrink paths a bit).
	// insetBounds := Rect{
	// 	left:   bounds.left + roadWidth,
	// 	top:    bounds.top + roadWidth,
	// 	right:  bounds.right - roadWidth,
	// 	bottom: bounds.bottom - roadWidth,
	// }
	// rescaledPoints = rescale(rescaledPoints, insetBounds)
	rounded := smoothCorners(rescaledPoints)
	inner := expand(rounded, roadWidth)
	outer := expand(rounded, -roadWidth)

	return TrackDebugData{
		Orig:      points,
		Perturbed: rescaledPoints,
		Rounded:   rounded,
		Inner:     inner,
		Outer:     outer,
	}
}

func BuildTrack(numPoints int, bounds Rect, roadWidth float64) (inner []Point, outer []Point) {
	for {
		trackData := BuildPossiblyIntersectingTrack(numPoints, bounds, roadWidth)
		if !IsSelfIntersecting(trackData.Inner) && !IsSelfIntersecting(trackData.Outer) {
			inner = trackData.Inner
			outer = trackData.Outer
			return
		}
	}
}
