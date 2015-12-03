package bingo

import (
	"image"
	"testing"
)

func TestByte(t *testing.T) {

	// Create a new Deterministic Byte Machine
	dbm := NewDbm("awesome data")

	// Create an Avatar
	avatar := &Avatar{100, 100}

	bytes := []byte{140, 66, 202}
	for _, b := range bytes {
		get := dbm.get()
		if get != b {
			t.Errorf("dbm.get() == %d, want %d", get, b)
		}
	}

	ints := []int{0, 1, 3}
	for _, i := range ints {
		get := dbm.getScaled(10)
		if get != i {
			t.Errorf("dbm.getScaled() == %d, want %d", get, i)
		}
	}

	intsX := []int{52, 7, 96}
	for _, i := range intsX {
		get := dbm.getX(avatar)
		if get != i {
			t.Errorf("dbm.getX() == %d, want %d", get, i)
		}
	}

	intsY := []int{56, 30, 57}
	for _, i := range intsY {
		get := dbm.getY(avatar)
		if get != i {
			t.Errorf("dbm.getY() == %d, want %d", get, i)
		}
	}

	points := []image.Point{
		{54, 51},
		{64, 51},
		{54, 99},
	}
	for _, p := range points {
		get := dbm.point(avatar)
		if get != p {
			t.Errorf("dbm.point() == %v, want %v", get, p)
		}
	}

	points = []image.Point{
		{19, 73},
		{19, 69},
		{7, 68},
	}
	for _, p := range points {
		get := dbm.pointScaled(25, 75)
		if get != p {
			t.Errorf("dbm.pointScaled(25, 75) == %v, want %v", get, p)
		}
	}

	rectangles := []image.Rectangle{
		image.Rect(38, 35, 73, 85),
		image.Rect(32, 85, 77, 89),
		image.Rect(22, 21, 92, 67),
	}
	for _, r := range rectangles {
		get := dbm.rectangle(avatar)
		if get != r {
			t.Errorf("dbm.rectangle() == %v, want %v", get, r)
		}
	}

	bools := []bool{false, true, true}
	for _, b := range bools {
		get := dbm.bool()
		if get != b {
			t.Errorf("dbm.bool() == %v, want %v", get, b)
		}
	}
}
