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
	return colRange(l.Data, l.X)
}

func (l *Layer1D[X, Y]) yRange() (lo, hi float64) {
	return colRange(l.Data, l.Y)
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
		key := l.Geom.grouping(l.Data, row)
		s, ok := smap[key]
		if !ok {
			n++
			s = &series[X, Y]{d: l.Data, x: l.X, y: l.Y}
			smap[key] = s
			ss = append(ss, s)
		}
		s.rows = append(s.rows, row)
	}
	for _, s := range ss {
		draw := l.Geom.drawer(theme, c, xScale, yScale, scaleFactor)

		// Sort the rows by X then Y.
		sort.Sort(s)

		// No statistic, take all points.
		if l.Stat == nil {
			for _, row := range s.rows {
				draw(l.Data, row, float64(l.X.Get(l.Data, row)), float64(l.Y.Get(l.Data, row)))
			}
			continue
		}

		// Apply statistic.
		for row, ygroup := range group(l.Data, s.rows, l.X, l.Y) {
			y := l.Stat(func(yield func(Y) bool) {
				for _, y := range ygroup {
					if !yield(y) {
						break
					}
				}
			})
			draw(l.Data, row, float64(l.X.Get(l.Data, row)), y)
		}
	}

	return c.Image(), nil
}

func group[X, Y Scalar](d *Dataset, rows []int, x Column[X], y Column[Y]) iter.Seq2[int, []Y] {
	return func(yield func(int, []Y) bool) {
		var lastX X
		var lastRow int
		var ys []Y
		first := true
		for _, r := range rows {
			x := x.Get(d, r)
			if first {
				first = false
				lastX = x
				lastRow = r
				ys = []Y{y.Get(d, r)}
				continue
			}
			if lastX != x {
				if !yield(lastRow, ys) {
					return
				}

				// Reset state.
				lastX = x
				lastRow = r
				ys = []Y{y.Get(d, r)}
				continue
			}
			ys = append(ys, y.Get(d, r))
		}
		if !first {
			if !yield(lastRow, ys) {
				return
			}
		}
	}
}

type series[X, Y Scalar] struct {
	d    *Dataset
	rows []int
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
	xi, xj := s.x.Get(s.d, s.rows[i]), s.x.Get(s.d, s.rows[j])
	if xi == xj {
		yi, yj := s.y.Get(s.d, s.rows[i]), s.y.Get(s.d, s.rows[j])
		return yi < yj
	}
	return xi < xj
}

func colRange[T Scalar](d *Dataset, c Column[T]) (lo, hi float64) {
	hi = math.Inf(-1)
	lo = math.Inf(1)
	for value := range c.All(d) {
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
