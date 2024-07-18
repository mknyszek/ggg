package ggg

type Plot struct {
	layers []Layer
	opts   presentOpts
}

func NewPlot() *Plot {
	return &Plot{}
}

func (p *Plot) Layer(l Layer) *Plot {
	p.layers = append(p.layers, l)
	return p
}

func (p *Plot) Presentation(opts ...PresentationOption) *Plot {
	for _, opt := range opts {
		opt(&p.opts)
	}
	return p
}

type presentOpts struct {
	title  string
	x, y   axis
	legend legend
}

type axis struct {
	title            string
	min, max         float64
	userMin, userMax float64
	userLimits       bool
	logBase          int
	customTicks      []float64
}

type legend struct {
	visible   bool
	alignment int
}

type PresentationOption func(*presentOpts)

func Title(title string) PresentationOption {
	return func(opts *presentOpts) {
		opts.title = title
	}
}

func XAxis(title string, aOpts ...AxisOption) PresentationOption {
	return func(opts *presentOpts) {
		opts.x.title = title
		for _, aOpt := range aOpts {
			aOpt(&opts.x)
		}
	}
}

func YAxis(title string, aOpts ...AxisOption) PresentationOption {
	return func(opts *presentOpts) {
		opts.y.title = title
		for _, aOpt := range aOpts {
			aOpt(&opts.y)
		}
	}
}

type AxisOption func(*axis)

func Limits(min, max float64) AxisOption {
	return func(opts *axis) {
		opts.userMin, opts.userMax = min, max
		opts.userLimits = true
	}
}

func LogScale(base int) AxisOption {
	return func(opts *axis) {
		opts.logBase = base
	}
}

func Ticks(ticks ...float64) AxisOption {
	return func(opts *axis) {
		opts.customTicks = ticks
	}
}
