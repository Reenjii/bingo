package bingo

import (
	"crypto/sha1"
	"crypto/sha256"
	"image"
	"image/color"
)

// Deterministic byte machine.
type Dbm struct {
	data []byte
	i    int
}

// Create a new Deterministic Byte Machine.
func NewDbm(data string) *Dbm {
	dbm := &Dbm{i: 0}
	dbm.data = make([]byte, sha1.Size+sha256.Size+sha256.Size224)

	// initialize random data from input data
	bsha1 := sha1.Sum([]byte(data))
	bsha256 := sha256.Sum256([]byte(data))
	bsha224 := sha256.Sum224([]byte(data))
	copy(dbm.data, bsha1[:])
	copy(dbm.data[sha1.Size:], bsha256[:])
	copy(dbm.data[sha1.Size+sha256.Size:], bsha224[:])
	return dbm
}

// Get the next byte.
func (r *Dbm) get() byte {
	r.i = (r.i + 1) % len(r.data)
	return r.data[r.i]
}

// Get the next byte scaled to given integer.
func (r *Dbm) getScaled(s int) int {
	return int(float32(r.get()) / 255 * float32(s))
}

// Get the next byte scaled to the X dimension of Avatar.
func (r *Dbm) getX(avatar *Avatar) int {
	return r.getScaled(avatar.X)
}

// Get the next byte scaled to the Y dimension of Avatar.
func (r *Dbm) getY(avatar *Avatar) int {
	return r.getScaled(avatar.Y)
}

// Get a point scaled to the dimensions of Avatar.
func (r *Dbm) point(avatar *Avatar) image.Point {
	return image.Point{X: r.getScaled(avatar.X), Y: r.getScaled(avatar.Y)}
}

// Get a point scaled to the given dimensions.
func (r *Dbm) pointScaled(sx, sy int) image.Point {
	return image.Point{X: r.getScaled(sx), Y: r.getScaled(sy)}
}

// Get a rectangle.
func (r *Dbm) rectangle(avatar *Avatar) image.Rectangle {
	x, y := r.point(avatar), r.point(avatar)
	return image.Rect(x.X, x.Y, y.X, y.Y)
}

// Get a bool.
func (r *Dbm) bool() bool {
	return r.get()%2 == 0
}

// Get a color.
func (r *Dbm) color() color.RGBA {
	return color.RGBA{r.get(), r.get(), r.get(), 255}
}
