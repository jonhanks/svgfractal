// svgfractal is a simple test of drawing a koch line/curve
//
// Warning, there is almost no parameterization yet.  This is quick code, but each
// detail level produces 10^n line segments.
//
// I output to svg so that I don't have to worry about a 2d graphics library
// in go.
package main

import (
	"flag"
	"fmt"
	"github.com/ajstarks/svgo"
	"html/template"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
)

var (
	addr = flag.String("addr", "localhost:8080", "Port to listen on")
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

type CachedTemplate struct {
	t    *template.Template
	mod  int64
	path string
}

type PeanoOptions struct {
	Height        float64
	DisplayCenter bool
}

var (
	templates = make(map[string]*CachedTemplate)
)

func NewTemplate(path string) *CachedTemplate {
	fInfo, err := os.Stat(path)
	if err != nil {
		panic("Error loading template file " + path + " " + err.Error())
	}
	return &CachedTemplate{t: template.Must(template.ParseFiles(path)), mod: fInfo.ModTime().Unix(), path: path}
}

func (t *CachedTemplate) Execute(w io.Writer, d interface{}) error {
	fInfo, err := os.Stat(t.path)
	if err == nil && fInfo.ModTime().Unix() > t.mod {
		newT, err := template.ParseFiles(t.path)
		if err == nil {
			t.t = newT
			t.mod = fInfo.ModTime().Unix()
		}
	}
	return t.t.Execute(w, d)
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
	sinT, cosT := math.Sincos(theta)
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

func (v *Vector) Reverse() {
	v.X = -v.X
	v.Y = -v.Y
}

/* rename these to useful names at some point */
func NewLine(x1, y1, x2, y2 float64) Line {
	return Line{Start: Point{X: x1, Y: y1}, Direction: Vector{Point{X: x2 - x1, Y: y2 - y1}}, Scale: 1.0}
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
	//fmt.Printf("At %v\n", p)
	return p
}

func (l Line) GoString() string {
	return fmt.Sprintf("L: %s - %s %s", l.Start, l.Direction, l.Scale)
}

// draw a line
func (l Line) Render(s *svg.SVG) {
	if l.Scale != 0.0 {
		end := l.At(1.0)
		s.Line(int(l.Start.X), int(l.Start.Y), int(end.X), int(end.Y), "fill:none;stroke:black")
	}
}

// Return the length of the line
func (l Line) Length() float64 {
	if l.Direction.X == 0.0 {
		return l.Direction.Y * l.Scale
	} else if l.Direction.Y == 0.0 {
		return l.Direction.X * l.Scale
	}
	x := l.Direction.X * l.Scale
	y := l.Direction.Y * l.Scale
	return math.Sqrt(x*x + y*y)
}

func (l *Line) SetLength(newLength float64) {
	if l.Direction.X == 0.0 && l.Direction.Y == 0.0 {
		return
	}
	length := l.Length()
	if length != 1.0 {
		l.Direction.X /= length
		l.Direction.Y /= length
	}
	l.Scale = newLength
}

func (l *Line) Reverse() {
	l.Direction.Reverse()
}

// Find the perpendicular vector to the 2d vector v.
// This is done by computing the cross product of v with
// the Z unit vector
func cross(v Vector) *Vector {
	x1, y1, z1 := v.X, v.Y, 0.0
	_, y2, z2 := 0.0, 0.0, 1.0

	return &Vector{Point{X: y1*z2 - z1*y2, Y: z1*x1 - x1*z2}}
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
func doKochCurve(s *svg.SVG, l Line, depth int, rot *Matrix) {
	if depth <= 0 {
		l.Render(s)
	} else {
		//cdir := cross(l.Direction)
		cdir := MultMatrixVector(rot, &l.Direction)
		mid := l.At(0.5)
		l1, l2 := l.Split(0.333333)
		_, l3 := l2.Split(0.5)

		//fmt.Printf("---\n\tl1: %v\n\tl2: %v\n---\n", l1, l3)
		//fmt.Println(cdir)
		//l1.Render(s)
		doKochCurve(s, l1, depth-1, rot)
		//mid.Render(s)

		lmid := NewLine2(mid, *cdir, l1.Scale*1.0)
		//lmid.Render(s)

		mid2 := lmid.At(1.0)

		l2a := NewLine3(l1.At(1.0), mid2)
		//l2a.Render(s)

		l2b := NewLine3(mid2, l3.At(0.0))
		//l2b.Render(s)

		doKochCurve(s, l2a, depth-1, rot)
		doKochCurve(s, l2b, depth-1, rot)

		//l3.Render(s)
		doKochCurve(s, l3, depth-1, rot)
	}
}

func kochCurve(s *svg.SVG, x1, y1, x2, y2, complexity int, rotation float64) {
	l := NewLine(float64(x1), float64(y1), float64(x2), float64(y2))

	//fmt.Printf("kochCurve (%d, %d) - (%d, %d)\n", x1, y1, x2, y2)
	//fmt.Printf("line: %s\n\n", l)

	m := NewMatrix()
	m.Rotate(rotation)

	doKochCurve(s, l, complexity, m)
}

func kochCurveHandler(w http.ResponseWriter, req *http.Request) {
	const (
		defaultComplexity = 6
		maxComplexity     = 9

		defaultPi = 0.5
		maxPi     = 1.0
		minPi     = -1.0
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
	if pi < minPi {
		pi = minPi
	} else if pi > maxPi {
		pi = maxPi
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
		kochCurve(s, 0, 50, width-1, 50, complexity, -math.Pi*pi)
	} else {
		kochCurve(s, 0, height-50, width-1, height-50, complexity, -math.Pi*pi)
	}
}

func kochSnowflakeHandler(w http.ResponseWriter, req *http.Request) {
	const (
		defaultComplexity = 5
		maxComplexity     = 8
	)

	_ = req.ParseForm()
	complexity, err := strconv.Atoi(req.FormValue("complexity"))
	if err != nil || complexity < 0 || complexity > maxComplexity {
		complexity = defaultComplexity
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)
	width := 1000 + (4000 * complexity / maxComplexity)
	height := width

	offset := width / 4
	pi := .5
	rotation := -math.Pi * pi

	s.Start(width, height)
	defer s.End()
	kochCurve(s, offset, offset, width-offset, offset, complexity, rotation)
	kochCurve(s, width-offset, offset, width/2, height-offset, complexity, rotation)
	kochCurve(s, width/2, height-offset, offset, offset, complexity, rotation)
}

// Do the fractal
func doPeanoCurve(s *svg.SVG, l Line, options *PeanoOptions, depth int) {
	if depth <= 0 {
		l.Render(s)
	} else {
		height := options.Height

		length := l.Length()

		perpendicular := cross(l.Direction)

		//fmt.Printf("segment len %f perp %s\n", length, perpendicular)

		intersect1 := l.At(1.0 / 3.0)
		intersect2 := l.At(2.0 / 3.0)

		l1 := NewLine2(intersect1, *perpendicular, 1.0)
		l1.SetLength(length * height)

		l2 := NewLine2(intersect2, *perpendicular, 1.0)
		l2.SetLength(length * height)

		l3 := NewLine3(l1.At(1.0), l2.At(1.0))

		//perpendicular.Reverse()

		l4 := NewLine2(intersect1, *perpendicular, 1.0)
		l4.SetLength(-length * height)

		l5 := NewLine2(intersect2, *perpendicular, 1.0)
		l5.SetLength(-length * height)

		l6 := NewLine3(l4.At(1.0), l5.At(1.0))

		l7 := NewLine3(l.Start, intersect1)

		l8 := NewLine3(intersect2, l.At(1.0))

		doPeanoCurve(s, l1, options, depth-1)
		doPeanoCurve(s, l2, options, depth-1)
		doPeanoCurve(s, l3, options, depth-1)
		doPeanoCurve(s, l4, options, depth-1)
		doPeanoCurve(s, l5, options, depth-1)
		doPeanoCurve(s, l6, options, depth-1)
		doPeanoCurve(s, l7, options, depth-1)
		doPeanoCurve(s, l8, options, depth-1)

		if options.DisplayCenter {
			l9 := NewLine3(intersect1, intersect2)
			doPeanoCurve(s, l9, options, depth-1)
		}

	}
}

func peanoCurve(s *svg.SVG, x1, y1, x2, y2 int, options *PeanoOptions, complexity int) {
	l := NewLine(float64(x1), float64(y1), float64(x2), float64(y2))

	doPeanoCurve(s, l, options, complexity)
}

func peanoCurveHandler(w http.ResponseWriter, req *http.Request) {
	const (
		defaultComplexity = 5
		maxComplexity     = 8
		defaultHeight     = 1.0 / 3.0
		maxHeight         = 0.5
		minHeight         = 0.0
	)

	options := PeanoOptions{}

	_ = req.ParseForm()
	complexity, err := strconv.Atoi(req.FormValue("complexity"))
	if err != nil || complexity < 0 || complexity > maxComplexity {
		complexity = defaultComplexity
	}

	options.Height, err = strconv.ParseFloat(req.FormValue("height"), 64)
	if err != nil || options.Height < minHeight || options.Height > maxHeight {
		options.Height = defaultHeight
	}

	center := req.FormValue("center")
	if center == "true" {
		options.DisplayCenter = true
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)
	width := 1000 + (4000 * complexity / maxComplexity)
	height := width

	s.Start(width, height)
	defer s.End()

	peanoCurve(s, 0, height/2, width-1, height/2, &options, complexity)
}

func doDragonCurve(s *svg.SVG, l Line, useRot1 bool, rot1, rot2 *Matrix, depth int) bool {
	if depth <= 0 {
		l.Render(s)
		if useRot1 {
			fmt.Printf("1")
		} else {
			fmt.Printf("2")
		}
	} else {
		m := rot1
		if !useRot1 {
			m = rot2
		}
		perp := MultMatrixVector(m, &l.Direction)
		center := l.At(0.5)

		center_line := NewLine2(center, *perp, 1.0)
		center_line.SetLength(l.Length()/2.0)

		peak := center_line.At(1.0)

		useRot1 = doDragonCurve(s, NewLine3(l.At(0.0), peak), useRot1, rot1, rot2, depth-1)
		useRot1 = doDragonCurve(s, NewLine3(peak, l.At(1.0)), useRot1, rot1, rot2, depth-1)
	}
	if depth == 1 {
		return !useRot1
	}
	return useRot1
}

func dragonCurve(s *svg.SVG, x1, y1, x2, y2, complexity int) {
	l := NewLine(float64(x1), float64(y1), float64(x2), float64(y2))

	left, right := NewMatrix(), NewMatrix()
	left.Rotate(math.Pi/2.0)
	right.Rotate(-math.Pi/2.0)

	_ = doDragonCurve(s, l, true, left, right, complexity)
}

func dragonCurveHandler(w http.ResponseWriter, req *http.Request) {
	const (
		defaultComplexity = 5
		maxComplexity = 12
	)

	_ = req.ParseForm()
	complexity, err := strconv.Atoi(req.FormValue("complexity"))
	if err != nil || complexity < 0 || complexity > maxComplexity {
		complexity = defaultComplexity
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)

	width := 1000 + (4000*complexity/maxComplexity)
	height := width

	s.Start(width, height)
	defer s.End()

	dragonCurve(s, width/3, height/2, width-(width/3), height/2, complexity)
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	m := make(map[string]string)
	m[""] = ""
	t, ok := templates["index"]
	if ok {
		if err := t.Execute(w, &m); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "Missing template file", http.StatusInternalServerError)
}

func initTemplates() {
	for _, name := range []string{"index"} {
		fileName := fmt.Sprintf("templates/%s.html", name)
		templates[name] = NewTemplate(fileName)
	}
}

func main() {
	flag.Parse()

	initTemplates()

	fmt.Println("Starting server at localhost port 8080")
	fmt.Println("\nKoch curves/waves:")
	fmt.Println("Pass the following url parameters:")
	fmt.Println("complexity=n (where n in an integer in [0,9]")
	fmt.Println("pi=n (where n is a real number [-1.0, 1.0])")
	fmt.Println("\nPeano curves:")
	fmt.Println("complexity=n (where n is an integer in [0,9])")
	fmt.Println("height=n (where n is a real number [0.0, 1.0])")

	fmt.Println("\nDragon curves:")
	fmt.Println("complexity=n (where n is an integer in [0,9]")

	http.Handle("/", http.HandlerFunc(indexHandler))
	http.Handle("/linear/koch/curve/", http.HandlerFunc(kochCurveHandler))
	http.Handle("/linear/koch/snowflake/", http.HandlerFunc(kochSnowflakeHandler))
	http.Handle("/linear/peano/curve/", http.HandlerFunc(peanoCurveHandler))
	http.Handle("/linear/dragon/curve/", http.HandlerFunc(dragonCurveHandler))

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
