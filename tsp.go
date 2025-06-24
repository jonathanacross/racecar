package main

import (
	"math"

	"github.com/fogleman/gg"
)

func distance(p1 gg.Point, p2 gg.Point) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func GetShortestCycle(points []gg.Point) []gg.Point {
	greedy := getShortestCycleGreedy(points)
	optimized := optimizeCycle(greedy)
	return optimized
}

// getShortestCycle implements the Nearest Neighbor algorithm to find an approximate
// shortest cycle through a given slice of points (Traveling Salesperson Problem).
// It starts from the first point in the slice and greedily picks the nearest unvisited point.
func getShortestCycleGreedy(points []gg.Point) []gg.Point {
	if len(points) == 0 {
		return nil
	}
	if len(points) == 1 {
		return []gg.Point{points[0]} // Cycle with one point is just the point itself
	}

	// visited keeps track of which points have been added to the tour.
	// Initialized to false by default
	visited := make([]bool, len(points))

	// tour will store the sequence of points in the cycle.
	tour := make([]gg.Point, 0, len(points))

	// Start with the first point in the provided slice.
	startPointIndex := 0
	currentPoint := points[startPointIndex]
	tour = append(tour, currentPoint)
	visited[startPointIndex] = true

	// Go through the rest of the points, finding the closest unvisited point
	for len(tour) < len(points) {
		nearestPointIndex := -1
		minDistance := math.MaxFloat64

		// Find the nearest unvisited point from the current point
		for i, p := range points {
			if !visited[i] { // Only consider unvisited points
				dist := distance(currentPoint, p)
				if dist < minDistance {
					minDistance = dist
					nearestPointIndex = i
				}
			}
		}

		currentPoint = points[nearestPointIndex]
		tour = append(tour, currentPoint)
		visited[nearestPointIndex] = true
	}

	return tour
}

// reverseSubsegment reverses the order of elements in a slice from index i to j (inclusive).
// This is a helper for the 2-Opt swap.
func reverseSubsegment(tour []gg.Point, i, j int) {
	for i < j {
		tour[i], tour[j] = tour[j], tour[i]
		i++
		j--
	}
}

// optimizeCycle applies the 2-Opt swap optimization to an existing tour to remove crossings
// and potentially shorten the total path length. It iterates until no further improvements are found.
// This simplified version directly checks for distance improvement without explicit intersection check.
func optimizeCycle(initialTour []gg.Point) []gg.Point {
	if len(initialTour) <= 3 {
		return initialTour // No meaningful 2-opt for 3 or fewer points
	}

	currentTour := make([]gg.Point, len(initialTour))
	copy(currentTour, initialTour) // Work on a copy

	improved := true
	for improved {
		improved = false
		// Iterate through all possible pairs of non-adjacent edges (i, i+1) and (j, j+1)
		// The tour is closed, so the last edge is (len-1, 0)
		numPoints := len(currentTour)

		for i := 0; i < numPoints-1; i++ {
			// Iterate j from i+2 to numPoints-1 (to avoid adjacent segments and last segment crossing first)
			for j := i + 2; j < numPoints; j++ {
				// Define the four points for the two segments:
				// Segment 1: (A, B) where A = currentTour[i], B = currentTour[i+1]
				// Segment 2: (C, D) where C = currentTour[j], D = currentTour[(j+1)%numPoints]

				p1 := currentTour[i]
				q1 := currentTour[i+1]

				p2 := currentTour[j]
				q2 := currentTour[(j+1)%numPoints]

				// Calculate original lengths
				originalLen1 := distance(p1, q1)
				originalLen2 := distance(p2, q2)
				originalTotal := originalLen1 + originalLen2

				// Calculate new lengths if swapped: (A, C) and (B, D)
				newLen1 := distance(p1, p2)
				newLen2 := distance(q1, q2)
				newTotal := newLen1 + newLen2

				// If swapping results in a shorter path, perform the swap
				if newTotal < originalTotal {
					// Perform the 2-opt swap: reverse the segment between (i+1) and j
					// Note: The segment to reverse is currentTour[i+1 ... j]
					reverseSubsegment(currentTour, i+1, j)
					improved = true // Mark that an improvement was made
				}
			}
		}
	}
	return currentTour
}
