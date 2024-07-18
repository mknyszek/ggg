package ggg

import (
	"iter"

	"golang.org/x/perf/benchmath"
)

type Statistic1D[T Scalar] func(iter.Seq[T]) float64

func Sum[T Scalar](seq iter.Seq[T]) float64 {
	var sum T
	for v := range seq {
		sum += v
	}
	return float64(sum)
}

func Mean[T Scalar](seq iter.Seq[T]) float64 {
	var sum T
	n := 0
	for v := range seq {
		sum += v
		n++
	}
	return float64(sum) / float64(n)
}

type Statistic2D[T Scalar] func(iter.Seq[T]) (float64, float64)

func Confidence[T Scalar](confidence float64) func(seq iter.Seq[T]) (float64, float64) {
	return func(seq iter.Seq[T]) (float64, float64) {
		var f []float64
		for v := range seq {
			f = append(f, float64(v))
		}
		samp := benchmath.NewSample(f, &benchmath.DefaultThresholds)
		sum := benchmath.AssumeNothing.Summary(samp, confidence)
		return sum.Lo, sum.Hi
	}
}

func ConfidenceNormal[T Scalar](confidence float64) func(seq iter.Seq[T]) (float64, float64) {
	return func(seq iter.Seq[T]) (float64, float64) {
		var f []float64
		for v := range seq {
			f = append(f, float64(v))
		}
		samp := benchmath.NewSample(f, &benchmath.DefaultThresholds)
		sum := benchmath.AssumeNormal.Summary(samp, confidence)
		return sum.Lo, sum.Hi
	}
}
