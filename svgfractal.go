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
	"math"
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

type Matrix struct {
	Rows [][]float64
}

func NewMatrix() *Matrix {
	m := &Matrix{Rows: make([][]float64, 2)}
	for i := range m.Rows {
		m.Rows[i] = make([]float64, 2)
	}
	m.Identity()
	//m := Matrix{Rows: {{1.0, 0.0}, {0.0, 1.0}}}

	return m
}

// Set a matrix to be an identity matrix
func (m *Matrix) Identity() {
	for i := range m.Rows {
		for j := range m.Rows[i] {
			if i == j {
				m.Rows[i][j] = 1.0
			} else {
				m.Rows[i][j] = 0.0
			}
		}
	}
}

// rotate around the Z azis by theta degrees
func (m *Matrix) Rotate(theta float64) {
	cosT := math.Cos(theta)
	sinT := math.Sin(theta)
	m.Rows[0][0] = cosT
	m.Rows[0][1] = -sinT
	m.Rows[1][0] = sinT
	m.Rows[1][1] = cosT
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

// Take the product of m * v and store the result in dest
// This function is provided to reduce the amount of garbage
// created
func MultMatrixVectorWithDest(m *Matrix, v, dest *Vector) {
	dest.X = m.Rows[0][0]*v.X + m.Rows[0][1]*v.Y
	dest.Y = m.Rows[1][0]*v.X + m.Rows[1][1]*v.Y
}

// return the product of m * v
func MultMatrixVector(m *Matrix, v *Vector) *Vector {
	dest := &Vector{}
	MultMatrixVectorWithDest(m, v, dest)
	return dest
}

// Do the fractal
func doFractal(s *svg.SVG, l Line, depth int, rot *Matrix) {
	if depth <= 0 {
		l.Render(s)
	} else {
		//cdir := cross(l.Direction)
		cdir := MultMatrixVector(rot, &l.Direction)
		mid := l.At(0.5)
		l1, l2 := l.Split(0.333333)
		_, l3 := l2.Split(0.5)

		fmt.Printf("---\n\tl1: %v\n\tl2: %v\n---\n", l1, l3)
		fmt.Println(cdir)
		//l1.Render(s)
		doFractal(s, l1, depth-1, rot)
		//mid.Render(s)

		lmid := NewLine2(mid, *cdir, l1.Scale*1.0)
		//lmid.Render(s)

		mid2 := lmid.At(1.0)

		l2a := NewLine3(l1.At(1.0), mid2)
		//l2a.Render(s)

		l2b := NewLine3(mid2, l3.At(0.0))
		//l2b.Render(s)

		doFractal(s, l2a, depth-1, rot)
		doFractal(s, l2b, depth-1, rot)

		//l3.Render(s)
		doFractal(s, l3, depth-1, rot)
	}
}

func fractalLine(s *svg.SVG, x1, y1, x2, y2, complexity int, rotation float64) {
	l := NewLine(float64(x1), float64(y1), float64(x2), float64(y2))

	m := NewMatrix()
	m.Rotate(rotation)

	doFractal(s, l, complexity, m)
}

func lineFractal(w http.ResponseWriter, req *http.Request) {
	const (
		defaultComplexity = 6
		maxComplexity     = 9

		defaultPi = 0.5
	)

	_ = req.ParseForm()
	complexity, err := strconv.Atoi(req.FormValue("complexity"))
	if err != nil || complexity < 0 || complexity > maxComplexity {
		complexity = defaultComplexity
	}

	pi, err := strconv.ParseFloat(req.FormValue("pi"), 64)
	if err != nil {
		pi = defaultPi
	}
	if pi == 0.0 || pi == 1.0 {
		complexity = 0
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)
	width := 1000 + (4000 * complexity / maxComplexity)
	height := int(float64(width)*0.35) + 50
	s.Start(width, height)
	defer s.End()
	if pi < 0.0 {
		fractalLine(s, 0, 50, width-1, 50, complexity, -math.Pi*pi)
	} else {
		fractalLine(s, 0, height-50, width-1, height-50, complexity, -math.Pi*pi)
	}
}

func main() {
	fmt.Println("Starting server at localhost port 8080")
	fmt.Println("Pass the following url parameters:")
	fmt.Println("complexity=n (where n=[0,9]")
	fmt.Println("pi=n (where n is a real number")
	http.Handle("/", http.HandlerFunc(lineFractal))
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
