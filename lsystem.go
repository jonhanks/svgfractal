package main

import (
	"errors"
)

const BUF_SIZE = 1024 * 1024

// Lindenmayer system
type LSystem struct {
	rules      map[byte][]byte
	buf1       []byte
	buf2       []byte
	len1, len2 int
	cur        int
}

// create a New Lindenmayer system
func NewLSystem() *LSystem {
	return &LSystem{rules: make(map[byte][]byte), buf1: make([]byte, BUF_SIZE), buf2: make([]byte, BUF_SIZE), len1: 0, len2: 0, cur: 0}
}

// Setup the Lindenmayer object for computing a dragon fractal
func (sys *LSystem) InitDragon() {
	sys.rules = make(map[byte][]byte)
	sys.rules['X'] = []byte("X+YF")
	sys.rules['Y'] = []byte("FX-Y")

	sys.buf1[0] = 'F'
	sys.buf1[1] = 'X'
	sys.len1 = 2
	sys.len2 = 0
	sys.cur = 0
}

// Setup the Lindenmayer object for computing a plant
func (sys *LSystem) InitPlant1() {
	sys.rules = make(map[byte][]byte)
	sys.rules['X'] = []byte("F-[[X]+X]+F[+FX]-X")
	sys.rules['F'] = []byte("FF")

	sys.buf1[0] = 'X'
	sys.buf1[1] = 'F'
	sys.len1 = 2
	sys.len2 = 0
	sys.cur = 0
}

// Internal helper, get the current buffer
func (sys *LSystem) getCurBuf() ([]byte, int) {
	if sys.cur == 0 {
		return sys.buf1, sys.len1
	}
	return sys.buf2, sys.len2
}

// Internal helper, get the buffer to hold the next iteration of string
func (sys *LSystem) getNextBuf() []byte {
	if sys.cur == 0 {
		return sys.buf2
	}
	return sys.buf1
}

// Internal helper, set the next buffer to be the current buffer, with a specified length
func (sys *LSystem) swapBuffers(nextLength int) {
	sys.len1, sys.len2 = 0, 0
	if sys.cur == 0 {
		sys.cur = 1
		sys.len2 = nextLength
	} else {
		sys.cur = 0
		sys.len1 = nextLength
	}
}

// helper function, append a bytes slice to a buffer, return the length used, or an error if
// the length would overflow the given boundaries
func appendDest(dest []byte, curLength, maxLength int, value []byte) (int, error) {
	l := len(value)
	if curLength+l >= maxLength {
		return curLength, errors.New("overflow")
	}
	for i := 0; i < l; i++ {
		dest[i+curLength] = value[i]
	}
	return curLength + l, nil
}

// helper function, append a byte to a buffer, return the length used, or an error if
// the length would overflow the given boundaries
func appendDestByte(dest []byte, curLength, maxLength int, value byte) (int, error) {
	if curLength+1 >= maxLength {
		return curLength, errors.New("overflow")
	}
	dest[curLength] = value
	return curLength + 1, nil
}

// Iterate a L-System function through the specified number of iterations
func (sys *LSystem) IterateSystem(iterations int) {

	for ; iterations > 0; iterations-- {
		src, srcLen := sys.getCurBuf()
		dest := sys.getNextBuf()
		destLen := 0

		var err error

		for i := 0; i < srcLen; i++ {

			if replacement, ok := sys.rules[src[i]]; ok {
				if destLen, err = appendDest(dest, destLen, BUF_SIZE, replacement); err != nil {
					return
				}
			} else {
				if destLen, err = appendDestByte(dest, destLen, BUF_SIZE, src[i]); err != nil {
					return
				}
			}
		}

		sys.swapBuffers(destLen)
	}
}

func (sys *LSystem) FinalizePlant1(iterations int) {
	sys.InitPlant1()
	sys.IterateSystem(iterations)
}

// Iterate a dragon function through the specified number of iterations
func (sys *LSystem) FinalizeDragon(iterations int) {
	sys.InitDragon()
	sys.IterateSystem(iterations)
}

// Return a string version of the current output
func (sys *LSystem) String() string {
	buf, length := sys.getCurBuf()
	return string(buf[:length])
}

/*func main() {

	const MAX_ITER = 12

	dragon := NewLSystem()
	// for i := 0; i <= MAX_ITER; i++ {
	// 	dragon.IterateDragon(i)
	// 	fmt.Printf("%d: %s\n", i, dragon)
	// }
	dragon.FinalizeDragon(MAX_ITER)
	fmt.Printf("final: %s\n", dragon)
}*/
