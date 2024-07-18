package ggg

import (
	"fmt"
	"image"
	"iter"
	"math"
	"sort"

	"github.com/fogleman/gg"
)

type Layer interface {
	xRange() (lo, hi float64)
	yRange() (lo, hi float64)
	render(theme *Theme, width, height int, xScale, yScale scaleFunc) (image.Image, error)
}

type Layer1D[X, Y Scalar] struct {
	Data *Dataset
	X    Column[X]
	Y    Column[Y]
	Stat Statistic1D[Y]
	Geom Geom1D
}

func (l *Layer1D[X, Y]) xRange() (lo, hi float64) {
	return colRange(l.X)
}

func (l *Layer1D[X, Y]) yRange() (lo, hi float64) {
	return colRange(l.Y)
}

func (l *Layer1D[X, Y]) render(theme *Theme, width, height int, xScale, yScale scaleFunc) (image.Image, error) {
	if l.Geom.kind == kindBadGeom1D {
		return nil, fmt.Errorf("no initialized Geom for layer")
	}
	if l.Data == nil {
		return nil, fmt.Errorf("no intended dataset specified for layer")
	}
	if l.X == *new(Column[X]) {
		return nil, fmt.Errorf("no initialized X column for layer")
	}
	if l.Y == *new(Column[Y]) {
		return nil, fmt.Errorf("no initialized Y column for layer")
	}

	c := gg.NewContext(width, height)
	w, h := float64(width), float64(height)
	scaleFactor := math.Round(math.Sqrt(w * h / (1080 * 720)))
	n := 0
	smap := make(map[any]*series[X, Y])
	var ss []*series[X, Y]
	// Split the data into series.
	for row := range l.Data.Rows() {
		key := l.Geom.grouping(row)
		s, ok := smap[key]
		if !ok {
			n++
			s = &series[X, Y]{x: l.X, y: l.Y}
			smap[key] = s
			ss = append(ss, s)
		}
		s.rows = append(s.rows, row)
	}
	draw := l.Geom.drawer(theme, c, xScale, yScale, scaleFactor)
	for _, s := range ss {
		// Sort the rows by X then Y.
		sort.Sort(s)

		// No statistic, take all points.
		if l.Stat == nil {
			for _, row := range s.rows {
				draw(row, float64(Field(row, l.X)), float64(Field(row, l.Y)))
			}
			continue
		}

		// Apply statistic.
		for row, ygroup := range group(s.rows, l.X, l.Y) {
			y := l.Stat(func(yield func(Y) bool) {
				for _, y := range ygroup {
					if !yield(y) {
						break
					}
				}
			})
			draw(row, float64(Field(row, l.X)), y)
		}
	}

	return c.Image(), nil
}

func group[X, Y Scalar](rows []Row, x Column[X], y Column[Y]) iter.Seq2[Row, []Y] {
	return func(yield func(Row, []Y) bool) {
		var lastX X
		var lastRow Row
		var ys []Y
		first := true
		for _, r := range rows {
			x := Field(r, x)
			if first {
				first = false
				lastX = x
				lastRow = r
				ys = []Y{Field(r, y)}
				continue
			}
			if lastX != x {
				if !yield(lastRow, ys) {
					return
				}

				// Reset state.
				lastX = x
				lastRow = r
				ys = []Y{Field(r, y)}
				continue
			}
			ys = append(ys, Field(r, y))
		}
		if !first {
			if !yield(lastRow, ys) {
				return
			}
		}
	}
}

type series[X, Y Scalar] struct {
	rows []Row
	x    Column[X]
	y    Column[Y]
}

func (s *series[X, Y]) Len() int {
	return len(s.rows)
}

func (s *series[X, Y]) Swap(i, j int) {
	s.rows[i], s.rows[j] = s.rows[j], s.rows[i]
}

func (s *series[X, Y]) Less(i, j int) bool {
	xi, xj := Field(s.rows[i], s.x), Field(s.rows[j], s.x)
	if xi == xj {
		yi, yj := Field(s.rows[i], s.y), Field(s.rows[j], s.y)
		return yi < yj
	}
	return xi < xj
}

func colRange[T Scalar](c Column[T]) (lo, hi float64) {
	hi = math.Inf(-1)
	lo = math.Inf(1)
	for value := range c.Values() {
		v := float64(value)
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	if hi < lo {
		return 0, 0
	}
	return
}
