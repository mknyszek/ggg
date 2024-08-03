package ggg

import (
	"fmt"
	"image/color"

	"github.com/fogleman/gg"
)

type Geom struct {
	kind     kindGeom
	dims     int
	color    Mapping[color.Color]
	size     Mapping[float64]
	grouping func(d *Dataset, row int) any
}

type kindGeom int

const (
	kindBadGeom kindGeom = iota
	kindPoint
	kindLine
)

func (g *Geom) Dimensions() int {
	return g.dims
}

func Point(color Mapping[color.Color], size Mapping[float64]) *Geom {
	return &Geom{
		kind:  kindPoint,
		dims:  1,
		color: color,
		size:  size,
		grouping: func(d *Dataset, row int) any {
			return any2{color.selector(d, row), size.selector(d, row)}
		},
	}
}

func Line(color Mapping[color.Color], size Mapping[float64]) *Geom {
	return &Geom{
		kind:  kindLine,
		dims:  1,
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

func (g *Geom) drawer(th *Theme, c *gg.Context, xScale, yScale scaleFunc, scaleFactor float64) func(*Dataset, int, float64, []float64) {
	switch g.kind {
	case kindPoint:
		return func(d *Dataset, row int, x float64, y []float64) {
			if g.dims != len(y) {
				panic(fmt.Sprintf("%d-dimensional data applied to %d-dimensional geom", len(y), g.dims))
			}
			c.SetColor(g.color.scale(d, row, th))
			c.DrawCircle(xScale(x), yScale(y[0]), scaleFactor*g.size.scale(d, row, th))
			c.Fill()
		}
	case kindLine:
		var prev struct {
			x, y  float64
			valid bool
		}
		return func(d *Dataset, row int, x float64, y []float64) {
			if g.dims != len(y) {
				panic(fmt.Sprintf("%d-dimensional data applied to %d-dimensional geom", len(y), g.dims))
			}
			if prev.valid {
				c.SetColor(g.color.scale(d, row, th))
				c.MoveTo(xScale(prev.x), yScale(prev.y))
				c.LineTo(xScale(x), yScale(y[0]))
				c.SetLineWidth(scaleFactor * g.size.scale(d, row, th))
				c.Stroke()
			}
			prev.x = x
			prev.y = y[0]
			prev.valid = true
		}
	}
	panic("attempted to draw invalid Geom")
}
