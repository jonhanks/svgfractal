package main

import (
	"container/list"
	"github.com/ajstarks/svgo"
)

// Holds the state of the turtle object
type turtleState struct {
	location  Point
	direction Vector
	penUp     bool
}

// The turtle object
type Turtle struct {
	turtleState
	canvas *svg.SVG
	stack  *list.List
}

// Create a new turtle object
func NewTurtle(canvas *svg.SVG) *Turtle {
	return &Turtle{canvas: canvas, stack: list.New(), turtleState: turtleState{location: Point{X: 0.0, Y: 0.0}, direction: Vector{Point{X: 1.0, Y: 0.0}}, penUp: false}}
}

// Move the turtle a total of distance units, also draws a line segment following that path if the pen is down
func (t *Turtle) Move(distance float64) {
	startX, startY := int(t.location.X), int(t.location.Y)
	t.location.X += distance * t.direction.X
	t.location.Y += distance * t.direction.Y
	if !t.penUp {
		t.canvas.Line(startX, startY, int(t.location.X), int(t.location.Y), "fill:none;stroke:black")
	}
}

func (t *Turtle) PenUp() {
	t.penUp = true
}

func (t *Turtle) PenDown() {
	t.penUp = false
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

func (t *Turtle) PushState() {
	state := t.turtleState
	t.stack.PushFront(&state)
}

func (t *Turtle) PopState() {
	if t.stack.Len() > 0 {
		val := t.stack.Remove(t.stack.Front())
		if state, ok := val.(*turtleState); ok {
			t.turtleState = *state
		}
	}
}
