package ggg

import (
	"cmp"
	"regexp"
)

type Filter interface {
	Accept(d *Dataset, row int) bool
}

func FilterBy[T any](c Column[T], f func(T) bool) Filter {
	return &filterFunc[T]{c, f}
}

type filterFunc[T any] struct {
	c Column[T]
	f func(T) bool
}

func (f *filterFunc[T]) Accept(d *Dataset, row int) bool {
	return f.f(f.c.Get(d, row))
}

func EqualTo[T comparable](c Column[T], value T) Filter {
	return FilterBy(c, func(t T) bool {
		return t == value
	})
}

func NotEqualTo[T comparable](c Column[T], value T) Filter {
	return FilterBy(c, func(t T) bool {
		return t != value
	})
}

func LessThan[T cmp.Ordered](c Column[T], value T) Filter {
	return FilterBy(c, func(t T) bool {
		return t < value
	})
}

func LessThanOrEqualTo[T cmp.Ordered](c Column[T], value T) Filter {
	return FilterBy(c, func(t T) bool {
		return t <= value
	})
}

func GreaterThan[T cmp.Ordered](c Column[T], value T) Filter {
	return FilterBy(c, func(t T) bool {
		return t > value
	})
}

func GreaterThanOrEqualTo[T cmp.Ordered](c Column[T], value T) Filter {
	return FilterBy(c, func(t T) bool {
		return t >= value
	})
}

func Matches(c Column[string], r *regexp.Regexp) Filter {
	return FilterBy(c, func(t string) bool {
		return r.MatchString(t)
	})
}

func In[T comparable](c Column[T], values ...T) Filter {
	return FilterBy(c, func(t T) bool {
		for _, v := range values {
			if t == v {
				return true
			}
		}
		return false
	})
}

func Not(f Filter) Filter {
	return &filterNot{f}
}

type filterNot struct {
	f Filter
}

func (f *filterNot) Accept(d *Dataset, row int) bool {
	return !f.f.Accept(d, row)
}

func And(f ...Filter) Filter {
	return &filterAnd{f}
}

type filterAnd struct {
	fs []Filter
}

func (f *filterAnd) Accept(d *Dataset, row int) bool {
	for _, filter := range f.fs {
		if !filter.Accept(d, row) {
			return false
		}
	}
	return true
}

func Or(f ...Filter) Filter {
	return &filterOr{f}
}

type filterOr struct {
	fs []Filter
}

func (f *filterOr) Accept(d *Dataset, row int) bool {
	for _, filter := range f.fs {
		if filter.Accept(d, row) {
			return true
		}
	}
	return false
}
