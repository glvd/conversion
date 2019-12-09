package conversion

import "github.com/glvd/go-fftool"

// FragmentOption ...
type FragmentOption func(f *Fragment)

// Fragment ...
type Fragment struct {
	scale     Scale
	output    string
	skip      []string
	input     string
	sharpness string
}

// Scale ...
func (s Fragment) Scale() Scale {
	return s.scale
}

// Sharpness ...
func (s Fragment) Sharpness() string {
	return s.sharpness
}

// Output ...
func (s Fragment) Output() string {
	return s.output
}

func parseScale(scale int64) Scale {
	if scale > 1080 {
		return fftool.Scale1080P
	} else if scale > 720 {
		return fftool.Scale720P
	}
	return fftool.Scale480P
}
