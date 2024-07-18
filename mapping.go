package ggg

import "image/color"

type Mapping[O comparable] struct {
	selector func(Row) any
	scale    func(*Theme, Row) O
}

func Constant[O comparable](value O) Mapping[O] {
	return Mapping[O]{
		selector: func(_ Row) any {
			return nil
		},
		scale: func(_ *Theme, _ Row) O {
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
		selector: func(_ Row) any {
			return nil
		},
		scale: func(th *Theme, _ Row) color.Color {
			return th.SeriesPalette(uint64(i))
		},
	}
}

func NiceColors[I comparable](col Column[I]) Mapping[color.Color] {
	m := make(map[I]color.Color)
	n := uint64(0)
	return Mapping[color.Color]{
		selector: func(r Row) any {
			return Field(r, col)
		},
		scale: func(th *Theme, r Row) color.Color {
			i := Field(r, col)
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
		selector: func(r Row) any {
			return Field(r, col)
		},
		scale: func(_ *Theme, r Row) O {
			return f(Field(r, col))
		},
	}
}

func Identity[T comparable](col Column[T]) Mapping[T] {
	return Mapping[T]{
		selector: func(r Row) any {
			return Field(r, col)
		},
		scale: func(_ *Theme, r Row) T {
			return Field(r, col)
		},
	}
}

type Scalar interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}
