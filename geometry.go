package openpose

// Point represents coordinate
type Point struct {
	// X coordinate of body part
	X float64
	// Y coordinate of body part
	Y float64
}

// Pt returns a new Point
func Pt(x float64, y float64) Point {
	return Point{x, y}
}

// IsZero check if the Point is zero
func (p Point) IsZero() bool {
	return p.X <= 1e-15 && p.Y <= 1e-15
}

// ZP represents zero position point
var ZP = Point{0, 0}

// Rectangle represents rectangle
type Rectangle struct {
	X int
	Y int
	W int
	H int
}

// Rect returns a new Rectangle
func Rect(x, y, w, h int) Rectangle {
	return Rectangle{x, y, w, h}
}

// Intersect returns the largest rectangle contained by both r and s. If the
// two rectangles do not overlap then the zero rectangle will be returned.
func (r Rectangle) Intersect(s Rectangle) Rectangle {
	if r.X < s.X {
		r.X = s.X
	}
	if r.Y < s.Y {
		r.Y = s.Y
	}
	if r.X+r.W > s.X+s.W {
		r.W = s.X + s.W - r.X
	}
	if r.Y+r.H > s.Y+s.W {
		r.H = s.Y + s.H - r.H
	}
	return r
}

// Area returns rectangle area size
func (r Rectangle) Area() int {
	return r.W * r.H
}

// ZR represents zero size rectangle
var ZR = Rect(0, 0, 0, 0)

// Size represents Size
type Size struct {
	W float64
	H float64
}

// ASize returns a Size
func ASize(w, h float64) Size {
	return Size{w, h}
}

// IsZero check if the Size is zero
func (s Size) IsZero() bool {
	return s.W <= 1e-15 && s.H <= 1e-15
}

// ZS returns a new Size with zero
var ZS = ASize(0, 0)
