// svgfractal is a simple test of drawing a koch line/curve
//
// Warning, there is almost no parameterization yet.  This is quick code, but each
// detail level produces 10^n line segments.
//
// I output to svg so that I don't have to worry about a 2d graphics library
// in go.
package main

import (
	"fmt"
	"github.com/ajstarks/svgo"
	"net/http"
	"strconv"
)

// A point in 2d space
type Point struct {
	X, Y float64
}

// 2d vector
type Vector struct {
	Point
}

// A 2d line composed of a point, direction
// and a scale (used to view the line as a segment)
type Line struct {
	Start     Point
	Direction Vector
	Scale     float64
}

func (p Point) GoString() string {
	return fmt.Sprintf("p(%s, %s)", p.X, p.Y)
}

func (p Point) Render(s *svg.SVG) {
	s.Circle(int(p.X), int(p.Y), 1, "fill:none;stroke:black")
}

func (v Vector) GoString() string {
	return fmt.Sprintf("v(%s, %s)", v.X, v.Y)
}

/* rename these to useful names at some point */
func NewLine(x1, y1, x2, y2 float64) Line {
	return Line{Start: Point{X: x1, Y: y2}, Direction: Vector{Point{X: x2 - x1, Y: y2 - y1}}, Scale: 1.0}
}

func NewLine2(p Point, d Vector, s float64) Line {
	return Line{Start: p, Direction: d, Scale: s}
}

func NewLine3(p1, p2 Point) Line {
	return Line{Start: p1, Direction: Vector{Point{X: p2.X - p1.X, Y: p2.Y - p1.Y}}, Scale: 1.0}
}

// Split a line at point given origin + direction * scale * t
// return the line segments on each side of the point
func (l Line) Split(t float64) (Line, Line) {
	if t > 0.0 && t < 1.0 {

		newLine := Line{Start: l.At(t), Direction: l.Direction, Scale: (1.0 - t) * l.Scale}
		return Line{Start: l.Start, Direction: l.Direction, Scale: t * l.Scale}, newLine
	} else if t < 0.0 {
		return Line{Start: l.Start, Direction: l.Direction, Scale: t}, l
	} else if t == 0.0 {
		return Line{Start: l.Start, Direction: l.Direction, Scale: 0.0}, l
	} else if t == 1.0 {
		return l, Line{Start: l.Start, Direction: l.Direction, Scale: 0.0}
	} else {

	}
	return Line{}, Line{}
}

// return the point on a given line at scale factor t
func (l Line) At(t float64) Point {
	p := Point{X: l.Start.X + (t * l.Direction.X * l.Scale), Y: l.Start.Y + (t * l.Direction.Y * l.Scale)}
	fmt.Printf("At %v\n", p)
	return p
}

func (l Line) GoString() string {
	return fmt.Sprintf("L: %v - %v %s", l.Start, l.Direction, l.Scale)
}

// draw a line
func (l Line) Render(s *svg.SVG) {
	if l.Scale != 0.0 {
		end := l.At(1.0)
		s.Line(int(l.Start.X), int(l.Start.Y), int(end.X), int(end.Y), "fill:none;stroke:black")
	}
}

// Do a cross product of 2 2d vectors (temporarily viewing them in 3d for the math)
func cross(v Vector) Vector {
	x1, y1, z1 := v.X, v.Y, 0.0
	_, y2, z2 := 0.0, 0.0, 1.0

	return Vector{Point{X: y1*z2 - z1*y2, Y: z1*x1 - x1*z2}}
}

// Do the fractal
func doFractal(s *svg.SVG, l Line, depth int) {
	if depth <= 0 {
		l.Render(s)
	} else {
		cdir := cross(l.Direction)
		mid := l.At(0.5)
		l1, l2 := l.Split(0.333333)
		_, l3 := l2.Split(0.5)

		fmt.Printf("---\n\tl1: %v\n\tl2: %v\n---\n", l1, l3)
		fmt.Println(cdir)
		//l1.Render(s)
		doFractal(s, l1, depth-1)
		//mid.Render(s)

		lmid := NewLine2(mid, cdir, l1.Scale*1.0)
		//lmid.Render(s)

		mid2 := lmid.At(1.0)

		l2a := NewLine3(l1.At(1.0), mid2)
		//l2a.Render(s)

		l2b := NewLine3(mid2, l3.At(0.0))
		//l2b.Render(s)

		doFractal(s, l2a, depth-1)
		doFractal(s, l2b, depth-1)

		//l3.Render(s)
		doFractal(s, l3, depth-1)
	}
}

func fractalLine(s *svg.SVG, x1, y1, x2, y2, complexity int) {
	l := NewLine(float64(x1), float64(y1), float64(x2), float64(y2))

	doFractal(s, l, complexity)
}

func lineFractal(w http.ResponseWriter, req *http.Request) {
	const (
		defaultComplexity = 6
		maxComplexity     = 9
	)

	_ = req.ParseForm()
	complexity, err := strconv.Atoi(req.FormValue("complexity"))
	if err != nil || complexity < 0 || complexity > maxComplexity {
		complexity = defaultComplexity
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)
	width := 1000 + (4000 * complexity / maxComplexity)
	height := int(float64(width)*0.35) + 50
	s.Start(width, height)
	defer s.End()
	fractalLine(s, 0, height-50, width-1, height-50, complexity)
}

func main() {
	fmt.Println("Starting server at localhost port 8080")
	fmt.Println("Pass the url parameter complexity=n (where n=[0,9]")
	http.Handle("/", http.HandlerFunc(lineFractal))
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
