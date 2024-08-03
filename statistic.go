package ggg

import (
	"fmt"
	"iter"

	"golang.org/x/perf/benchmath"
)

type Statistic[T Scalar] struct {
	f    func(iter.Seq[T], []float64)
	dims int
}

func (s Statistic[T]) Valid() bool {
	return s.f != nil
}

func (s Statistic[T]) Dimensions() int {
	return s.dims
}

func (s Statistic[T]) Apply(values iter.Seq[T]) []float64 {
	result := make([]float64, s.dims)
	s.f(values, result)
	return result
}

func (s Statistic[T]) ApplyInto(values iter.Seq[T], result []float64) {
	if len(result) != s.dims {
		panic(fmt.Sprintf("%d-dimensional statistic applied to %d-dimensional result", s.dims, len(result)))
	}
	s.f(values, result)
}

func Count[T Scalar]() Statistic[T] {
	return Statistic[T]{
		f: func(seq iter.Seq[T], result []float64) {
			var n int
			for _ = range seq {
				n++
			}
			result[0] = float64(n)
		},
		dims: 1,
	}
}

func Sum[T Scalar]() Statistic[T] {
	return Statistic[T]{
		f: func(seq iter.Seq[T], result []float64) {
			var sum T
			for v := range seq {
				sum += v
			}
			result[0] = float64(sum)
		},
		dims: 1,
	}
}

func Mean[T Scalar]() Statistic[T] {
	return Statistic[T]{
		f: func(seq iter.Seq[T], result []float64) {
			var sum T
			n := 0
			for v := range seq {
				sum += v
				n++
			}
			result[0] = float64(sum) / float64(n)
		},
		dims: 1,
	}
}

func Confidence[T Scalar](confidence float64) Statistic[T] {
	return Statistic[T]{
		f: func(seq iter.Seq[T], result []float64) {
			var f []float64
			for v := range seq {
				f = append(f, float64(v))
			}
			samp := benchmath.NewSample(f, &benchmath.DefaultThresholds)
			sum := benchmath.AssumeNothing.Summary(samp, confidence)
			result[0], result[1] = sum.Lo, sum.Hi
		},
		dims: 1,
	}
}

func ConfidenceNormal[T Scalar](confidence float64) Statistic[T] {
	return Statistic[T]{
		f: func(seq iter.Seq[T], result []float64) {
			var f []float64
			for v := range seq {
				f = append(f, float64(v))
			}
			samp := benchmath.NewSample(f, &benchmath.DefaultThresholds)
			sum := benchmath.AssumeNormal.Summary(samp, confidence)
			result[0], result[1] = sum.Lo, sum.Hi
		},
		dims: 1,
	}
}
