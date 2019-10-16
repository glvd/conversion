package conversion

// Scale ...
type Scale int64

// HighScale ...
const HighScale Scale = 1080

// MiddleScale ...
const MiddleScale Scale = 720

// LowScale ...
const LowScale Scale = 480

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
func (s Fragment) Scale() int64 {
	return scale(s.scale)
}

// Sharpness ...
func (s Fragment) Sharpness() string {
	return s.sharpness
}

func toScale(scale int64) Scale {
	if scale > 1080 {
		return HighScale
	} else if scale > 720 {
		return MiddleScale
	}
	return LowScale
}

func scale(scale Scale) int64 {
	switch scale {
	case 480, 1080:
		return int64(scale)
	default:
		return 720
	}
}
