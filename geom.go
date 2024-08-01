package ggg

import (
	"image/color"

	"github.com/fogleman/gg"
)

type Geom1D struct {
	kind     kindGeom1D
	color    Mapping[color.Color]
	size     Mapping[float64]
	grouping func(d *Dataset, row int) any
}

type kindGeom1D int

const (
	kindBadGeom1D kindGeom1D = iota
	kindPoints
	kindLine
)

func Points(color Mapping[color.Color], size Mapping[float64]) Geom1D {
	return Geom1D{
		kind:  kindPoints,
		color: color,
		size:  size,
		grouping: func(d *Dataset, row int) any {
			return any2{color.selector(d, row), size.selector(d, row)}
		},
	}
}

func Line(color Mapping[color.Color], size Mapping[float64]) Geom1D {
	return Geom1D{
		kind:  kindLine,
		color: color,
		size:  size,
		grouping: func(d *Dataset, row int) any {
			return any2{color.selector(d, row), size.selector(d, row)}
		},
	}
}

type any2 struct {
	a, b any
}

func (g Geom1D) drawer(th *Theme, c *gg.Context, xScale, yScale scaleFunc, scaleFactor float64) func(*Dataset, int, float64, float64) {
	switch g.kind {
	case kindPoints:
		return func(d *Dataset, row int, x, y float64) {
			c.SetColor(g.color.scale(d, row, th))
			c.DrawCircle(xScale(x), yScale(y), scaleFactor*g.size.scale(d, row, th))
			c.Fill()
		}
	case kindLine:
		var prev struct {
			x, y  float64
			valid bool
		}
		return func(d *Dataset, row int, x, y float64) {
			if prev.valid {
				c.SetColor(g.color.scale(d, row, th))
				c.MoveTo(xScale(prev.x), yScale(prev.y))
				c.LineTo(xScale(x), yScale(y))
				c.SetLineWidth(scaleFactor * g.size.scale(d, row, th))
				c.Stroke()
			}
			prev.x = x
			prev.y = y
			prev.valid = true
		}
	}
	panic("attempted to draw invalid Geom1D")
}
