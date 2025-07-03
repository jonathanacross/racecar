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
