package ggg

import "image/color"

type Mapping[O comparable] struct {
	selector func(*Dataset, int) any
	scale    func(*Dataset, int, *Theme) O
}

func Constant[O comparable](value O) Mapping[O] {
	return Mapping[O]{
		selector: func(_ *Dataset, _ int) any {
			return nil
		},
		scale: func(_ *Dataset, _ int, _ *Theme) O {
			return value
		},
	}
}

func ScaleLinear[I, O Scalar](col Column[I], i0, i1 I, o0, o1 O) Mapping[O] {
	s := scaleLinear(float64(i0), float64(i1), float64(o0), float64(o1))
	return Scale(col, func(i I) O {
		return O(s(float64(i)))
	})
}

func ScaleOrdinal[I, O comparable](col Column[I], i []I, o []O, alt O) Mapping[O] {
	im := make(map[I]O)
	for idx, in := range i {
		im[in] = o[idx]
	}
	return Scale(col, func(i I) O {
		if out, ok := im[i]; ok {
			return out
		}
		return alt
	})
}

func PaletteColor(i uint64) Mapping[color.Color] {
	return Mapping[color.Color]{
		selector: func(_ *Dataset, _ int) any {
			return nil
		},
		scale: func(_ *Dataset, _ int, th *Theme) color.Color {
			return th.SeriesPalette(uint64(i))
		},
	}
}

func NiceColors[I comparable](col Column[I]) Mapping[color.Color] {
	m := make(map[I]color.Color)
	n := uint64(0)
	return Mapping[color.Color]{
		selector: func(d *Dataset, row int) any {
			return col.Get(d, row)
		},
		scale: func(d *Dataset, row int, th *Theme) color.Color {
			i := col.Get(d, row)
			if c, ok := m[i]; ok {
				return c
			}
			c := th.SeriesPalette(n)
			m[i] = c
			n++
			return c
		},
	}
}

func Scale[I, O comparable](col Column[I], f func(I) O) Mapping[O] {
	return Mapping[O]{
		selector: func(d *Dataset, row int) any {
			return col.Get(d, row)
		},
		scale: func(d *Dataset, row int, _ *Theme) O {
			return f(col.Get(d, row))
		},
	}
}

func Identity[T comparable](col Column[T]) Mapping[T] {
	return Mapping[T]{
		selector: func(d *Dataset, row int) any {
			return col.Get(d, row)
		},
		scale: func(d *Dataset, row int, _ *Theme) T {
			return col.Get(d, row)
		},
	}
}

type Scalar interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}
