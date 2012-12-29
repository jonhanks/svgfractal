package main

import (
	"github.com/ajstarks/svgo"
)

type Turtle struct {
	location  Point
	direction Vector
	canvas    *svg.SVG
}

// Create a new turtle object
func NewTurtle(canvas *svg.SVG) *Turtle {
	return &Turtle{location: Point{X: 0.0, Y: 0.0}, direction: Vector{Point{X: 1.0, Y: 0.0}}, canvas: canvas}
}

// Move the turtle a total of distance units, also draw a line segment following that path
func (t *Turtle) Move(distance float64) {
	startX, startY := int(t.location.X), int(t.location.Y)
	t.location.X += distance * t.direction.X
	t.location.Y += distance * t.direction.Y
	t.canvas.Line(startX, startY, int(t.location.X), int(t.location.Y), "fill:none;stroke:black")
}

// Set the location of the turtle
func (t *Turtle) SetLocation(p Point) {
	t.location = p
}

// Set the direction of the turtle
func (t *Turtle) SetDirection(dir Vector) {
	t.direction = dir
	t.direction.MakeUnitLength()
}

// Rotate the turtle by the given angle (in radians)
func (t *Turtle) Turn(angle float64) {
	m := NewMatrix()
	m.Rotate(angle)
	v := MultMatrixVector(m, &t.direction)
	t.direction.X = v.X
	t.direction.Y = v.Y
}
