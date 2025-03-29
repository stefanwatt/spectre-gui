package neovim

import "fmt"

type RGBA struct {
	R int
	G int
	B int
	A float64
}

type HSV struct {
	H, S, V float64
}

type XYZ struct {
	X, Y, Z float64
}

func (rgba *RGBA) copy() *RGBA {
	if rgba == nil {
		return nil
	}
	return &RGBA{
		R: rgba.R,
		G: rgba.G,
		B: rgba.B,
		A: rgba.A,
	}
}

func (rgba *RGBA) equals(other *RGBA) bool {
	if rgba == nil {
		return false
	}
	if other == nil {
		return false
	}
	return rgba.R == other.R && rgba.G == other.G && rgba.B == other.B
}

func (rgba *RGBA) String() string {
	if rgba == nil {
		return ""
	}
	return fmt.Sprintf("rgba(%d, %d, %d, %f)", rgba.R, rgba.G, rgba.B, rgba.A)
}

