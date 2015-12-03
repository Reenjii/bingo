package bingo

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math"
	"sort"
)

type Avatar struct {
	X int
	Y int
}

// Gradient.
func gradient(img *image.RGBA, from, to color.RGBA, x, y int, horizontal bool) {
	s := [3]float32{
		(float32(to.R) - float32(from.R)) / float32(x),
		(float32(to.G) - float32(from.G)) / float32(x),
		(float32(to.B) - float32(from.B)) / float32(x),
	}

	for i := 0; i < x; i++ {
		for j := 0; j < y; j++ {
			a, b := i, j
			if horizontal {
				a, b = j, i
			}
			img.SetRGBA(
				a,
				b,
				color.RGBA{
					uint8(float32(from.R) + float32(i)*s[0]),
					uint8(float32(from.G) + float32(i)*s[1]),
					uint8(float32(from.B) + float32(i)*s[2]),
					255,
				})
		}
	}
}

// Rectangle.
func rectangle(img *image.RGBA, rect image.Rectangle, c color.RGBA) {
	for i := rect.Min.X; i <= rect.Max.X; i++ {
		for j := rect.Min.Y; j <= rect.Max.Y; j++ {
			img.SetRGBA(i, j, c)
		}
	}
}

// Circle.
func circle(img *image.RGBA, center image.Point, r int, c color.RGBA) {
	for i := center.X - r; i <= center.X+r; i++ {
		for j := center.Y - r; j <= center.Y+r; j++ {
			if math.Pow(float64(i)-float64(center.X), 2)+math.Pow(float64(j)-float64(center.Y), 2) <= math.Pow(float64(r), 2)+float64(r)*0.8 {
				img.SetRGBA(i, j, c)
			}
		}
	}
}

// Ellipse.
func ellipse(img *image.RGBA, center, r image.Point, c color.RGBA) {
	for i := center.X - r.X; i <= center.X+r.X; i++ {
		for j := center.Y - r.Y; j <= center.Y+r.Y; j++ {
			if math.Pow((float64(i)-float64(center.X))/float64(r.X), 2)+math.Pow((float64(j)-float64(center.Y))/float64(r.Y), 2) <= 1.1 {
				img.SetRGBA(i, j, c)
			}
		}
	}
}

// Polygon.
func (avatar *Avatar) polygon(img *image.RGBA, points []image.Point, c color.RGBA) {
	// For each row
	for j := 0; j <= avatar.Y; j++ {

		// Build the list of Xs at which the row crosses a polygon edge
		intersect := make([]int, 0, len(points))
		adj := len(points) - 1
		for i, p := range points {
			q := points[adj]

			if (j > p.Y && j <= q.Y) || (j > q.Y && j <= p.Y) {
				x := int(float64(p.X) + (float64(j)-float64(p.Y))/(float64(q.Y)-float64(p.Y))*(float64(q.X)-float64(p.X)))
				intersect = append(intersect, x)
			}

			adj = i
		}

		// Sort the list f Xs
		sort.Ints(intersect)

		// Fill the pixels between node pairs
		for i := 0; i < len(intersect); i += 2 {
			for k := intersect[i]; k < intersect[i+1]; k++ {
				img.SetRGBA(k, j, c)
			}
		}
	}
}

// Draw a random shape.
func (avatar *Avatar) shape(img *image.RGBA, r *Dbm) {
	s := int(r.get()) % 7
	switch s {
	case 0:
		rectangle(img, r.rectangle(avatar), r.color())
	case 1:
		circle(img, r.point(avatar), r.getScaled(avatar.X/2), r.color())
	case 2:
		ellipse(img, r.point(avatar), r.pointScaled(avatar.X/2, avatar.Y/2), r.color())
	case 3, 4, 5, 6:
		points := make([]image.Point, 0, s)
		for i := 0; i < s; i++ {
			points = append(points, r.point(avatar))
		}
		avatar.polygon(img, points, r.color())
	}
}

// Creates an Avatar from input data.
// Returns a png image encoded as a base64 string.
func (avatar *Avatar) Avatar(data string) string {
	// Create a deterministic byte machine
	r := NewDbm(data)

	// Create a new RGBA image
	var img = image.NewRGBA(image.Rect(0, 0, avatar.X, avatar.Y))

	// Draw a background gradient
	gradient(img, r.color(), r.color(), avatar.X, avatar.Y, r.bool())

	// Draw 3 shapes
	for i := 0; i < 3; i++ {
		avatar.shape(img, r)
	}

	// Encode image as png
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		panic(err)
	}

	// Export as base64 string
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	return b64
}
