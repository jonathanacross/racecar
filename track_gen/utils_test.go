package main

import (
	"math"
	"reflect"
	"testing"
)

func TestDist(t *testing.T) {
	tests := []struct {
		name     string
		p1, p2   Point
		expected float64
	}{
		{
			name:     "Origin to (3,4)",
			p1:       Point{X: 0, Y: 0},
			p2:       Point{X: 3, Y: 4},
			expected: 5.0,
		},
		{
			name:     "Same point",
			p1:       Point{X: 5, Y: 5},
			p2:       Point{X: 5, Y: 5},
			expected: 0.0,
		},
		{
			name:     "Nonzero coordinates",
			p1:       Point{X: -1, Y: -1},
			p2:       Point{X: -4, Y: -5},
			expected: 5.0,
		},
		{
			name:     "Float coordinates",
			p1:       Point{X: 1.0, Y: 1.0},
			p2:       Point{X: 2.0, Y: 2.0},
			expected: math.Sqrt(2.0), // ~1.414
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Dist(tt.p1, tt.p2)
			if math.Abs(actual-tt.expected) > 1e-9 {
				t.Errorf("Dist(%v, %v) = %f; want %f",
					tt.p1, tt.p2, actual, tt.expected)
			}
		})
	}
}

func TestNorm(t *testing.T) {
	tests := []struct {
		name     string
		vec      Point
		expected Point
	}{
		{
			name:     "Vector (3,4)",
			vec:      Point{X: 3, Y: 4},
			expected: Point{X: 0.6, Y: 0.8},
		},
		{
			name:     "Zero vector",
			vec:      Point{X: 0, Y: 0},
			expected: Point{X: 0, Y: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Norm(tt.vec)
			if math.Abs(actual.X-tt.expected.X) > 1e-9 || math.Abs(actual.Y-tt.expected.Y) > 1e-9 {
				t.Errorf("Norm(%v) = %v; want %v",
					tt.vec, actual, tt.expected)
			}
		})
	}
}

func TestArea(t *testing.T) {
	tests := []struct {
		name     string
		polygon  []Point
		expected float64
	}{
		{
			name: "Square (CCW orientation)",
			polygon: []Point{
				{X: 0, Y: 0},
				{X: 10, Y: 0},
				{X: 10, Y: 10},
				{X: 0, Y: 10},
			},
			expected: 100.0,
		},
		{
			name: "Square (CW orientation)",
			polygon: []Point{
				{X: 0, Y: 0},
				{X: 0, Y: 10},
				{X: 10, Y: 10},
				{X: 10, Y: 0},
			},
			expected: -100.0,
		},
		{
			name: "Triangle",
			polygon: []Point{
				{X: 0, Y: 0},
				{X: 5, Y: 10},
				{X: 10, Y: 0},
			},
			expected: 50.0, // 0.5 * base * height = 0.5 * 10 * 10 = 50
		},
		{
			name:     "Empty polygon",
			polygon:  []Point{},
			expected: 0.0,
		},
		{
			name:     "Polygon with 1 point",
			polygon:  []Point{{X: 1, Y: 1}},
			expected: 0.0,
		},
		{
			name:     "Polygon with 2 points (degenerate line)",
			polygon:  []Point{{X: 0, Y: 0}, {X: 5, Y: 5}},
			expected: 0.0,
		},
		{
			name: "Concave",
			polygon: []Point{
				{X: 0, Y: 0},
				{X: 4, Y: 0},
				{X: 4, Y: 4},
				{X: 2, Y: 2},
				{X: 0, Y: 4},
			},
			expected: 12.0,
		},
		{
			name: "Turned square",
			polygon: []Point{
				{X: 0, Y: 0},
				{X: 2, Y: 1},
				{X: 1, Y: 3},
				{X: -1, Y: 2},
			},
			expected: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Area(tt.polygon)
			if math.Abs(actual-tt.expected) > 1e9 {
				t.Errorf("Area(%v) = %f; want %f",
					tt.polygon, actual, tt.expected)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name     string
		input    []Point
		expected []Point
	}{
		{
			name:     "Even number of points",
			input:    []Point{{X: 1, Y: 1}, {X: 2, Y: 2}, {X: 3, Y: 3}, {X: 4, Y: 4}},
			expected: []Point{{X: 4, Y: 4}, {X: 3, Y: 3}, {X: 2, Y: 2}, {X: 1, Y: 1}},
		},
		{
			name:     "Odd number of points",
			input:    []Point{{X: 1, Y: 1}, {X: 2, Y: 2}, {X: 3, Y: 3}},
			expected: []Point{{X: 3, Y: 3}, {X: 2, Y: 2}, {X: 1, Y: 1}},
		},
		{
			name:     "Two points",
			input:    []Point{{X: 1, Y: 1}, {X: 2, Y: 2}},
			expected: []Point{{X: 2, Y: 2}, {X: 1, Y: 1}},
		},
		{
			name:     "One point",
			input:    []Point{{X: 1, Y: 1}},
			expected: []Point{{X: 1, Y: 1}},
		},
		{
			name:     "Empty slice",
			input:    []Point{},
			expected: []Point{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the input slice because Reverse modifies it in-place
			inputCopy := make([]Point, len(tt.input))
			copy(inputCopy, tt.input)

			Reverse(inputCopy)
			if !reflect.DeepEqual(inputCopy, tt.expected) {
				t.Errorf("Reverse(%v) = %v; want %v",
					tt.input, inputCopy, tt.expected)
			}
		})
	}
}

func TestSegmentsIntersect(t *testing.T) {
	tests := []struct {
		name     string
		p1, q1   Point
		p2, q2   Point
		expected bool
	}{
		{
			name:     "Intersecting X-shape",
			p1:       Point{X: 1, Y: 1},
			q1:       Point{X: 10, Y: 10},
			p2:       Point{X: 1, Y: 8},
			q2:       Point{X: 10, Y: 1},
			expected: true,
		},
		{
			name:     "Non-intersecting parallel",
			p1:       Point{X: 0, Y: 0},
			q1:       Point{X: 5, Y: 0},
			p2:       Point{X: 0, Y: 1},
			q2:       Point{X: 5, Y: 1},
			expected: false,
		},
		{
			name:     "Non-intersecting separated",
			p1:       Point{X: 0, Y: 0},
			q1:       Point{X: 5, Y: 0},
			p2:       Point{X: 0, Y: 5},
			q2:       Point{X: 10, Y: 10},
			expected: false,
		},
		{
			name:     "Collinear overlapping",
			p1:       Point{X: 0, Y: 0},
			q1:       Point{X: 5, Y: 0},
			p2:       Point{X: 3, Y: 0},
			q2:       Point{X: 7, Y: 0},
			expected: true,
		},
		{
			name:     "Collinear non-overlapping",
			p1:       Point{X: 0, Y: 0},
			q1:       Point{X: 2, Y: 0},
			p2:       Point{X: 3, Y: 0},
			q2:       Point{X: 5, Y: 0},
			expected: false,
		},
		{
			name:     "Endpoint on segment",
			p1:       Point{X: 0, Y: 0},
			q1:       Point{X: 10, Y: 0},
			p2:       Point{X: 5, Y: 0},
			q2:       Point{X: 15, Y: 0},
			expected: true,
		},
		{
			name:     "Segments touch at endpoint",
			p1:       Point{X: 0, Y: 0},
			q1:       Point{X: 5, Y: 5},
			p2:       Point{X: 5, Y: 5},
			q2:       Point{X: 10, Y: 0},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := SegmentsIntersect(tt.p1, tt.q1, tt.p2, tt.q2)
			if actual != tt.expected {
				t.Errorf("SegmentsIntersect(%v, %v, %v, %v) = %t; want %t",
					tt.p1, tt.q1, tt.p2, tt.q2, actual, tt.expected)
			}
		})
	}
}

func TestIsSelfIntersecting(t *testing.T) {
	tests := []struct {
		name     string
		polygon  []Point
		expected bool
	}{
		{
			name: "Square",
			polygon: []Point{
				{X: 0, Y: 0},
				{X: 0, Y: 10},
				{X: 10, Y: 10},
				{X: 10, Y: 0},
			},
			expected: false,
		},
		{
			name: "Self-intersecting hourglass",
			polygon: []Point{
				{X: 0, Y: 0},
				{X: 10, Y: 0},
				{X: 0, Y: 10},
				{X: 10, Y: 10},
			},
			expected: true,
		},
		{
			name: "Non-degenerate Triangle",
			polygon: []Point{
				{X: 0, Y: 0},
				{X: 5, Y: 5},
				{X: 10, Y: 0},
			},
			expected: false,
		},
		{
			name: "Polygon with collinear overlapping segments",
			polygon: []Point{
				{X: 0, Y: 0},
				{X: 10, Y: 0},
				{X: 5, Y: 0},
				{X: 15, Y: 0},
			},
			expected: true,
		},
		{
			name:     "Polygon with less than 3 vertices",
			polygon:  []Point{{X: 0, Y: 0}, {X: 1, Y: 1}},
			expected: false,
		},
		{
			name:     "Empty polygon",
			polygon:  []Point{},
			expected: false,
		},
		{
			name: "Self-intersecting star",
			polygon: []Point{
				{X: 1, Y: 0},
				{X: 2, Y: 5},
				{X: 3, Y: 0},
				{X: 0, Y: 3},
				{X: 5, Y: 3},
			},
			expected: true,
		},
		{
			name: "Concave, non-self-intersecting",
			polygon: []Point{
				{X: 0, Y: 0},
				{X: 10, Y: 0},
				{X: 10, Y: 10},
				{X: 5, Y: 5},
				{X: 0, Y: 10},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsSelfIntersecting(tt.polygon)
			if actual != tt.expected {
				t.Errorf("IsSelfIntersecting(%v) = %t; want %t",
					tt.polygon, actual, tt.expected)
			}
		})
	}
}
