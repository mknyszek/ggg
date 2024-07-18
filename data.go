package ggg

import (
	"fmt"
	"iter"
)

// Dataset represents some data. The data are structured as a series
// of rows with uniformly-typed columns.
type Dataset struct {
	columns []columnI
	rows    int
	colMap  map[string]int
}

type columnData[T any] struct {
	name   string
	values []T
}

func (c *columnData[T]) id() string {
	return c.name
}

func (c *columnData[T]) grow(n int) {
	c.values = append(c.values, make([]T, n)...)
}

type columnI interface {
	id() string
	grow(n int)
}

// Empty returns an empty dataset.
func Empty() *Dataset {
	return &Dataset{colMap: make(map[string]int)}
}

// Extend creates a new column with the provided name and type and
// extends the dataset with it. Returns a handle to the new column.
func Extend[T any](d *Dataset, name string) (Column[T], bool) {
	if _, ok := d.colMap[name]; ok {
		return Column[T]{}, false
	}
	d.columns = append(d.columns, &columnData[T]{name, make([]T, d.rows)})
	d.colMap[name] = len(d.columns)
	return Column[T]{d, len(d.columns) - 1}, true
}

// Append adds new rows to the dataset and returns an iterator producing
// those new rows.
func Append(d *Dataset, n int) iter.Seq[Row] {
	r := d.rows
	d.rows += n
	for _, c := range d.columns {
		c.grow(n)
	}
	return func(yield func(Row) bool) {
		for i := r; i < r+n; i++ {
			if !yield(Row{d, i}) {
				return
			}
		}
	}
}

// Column represents a column of uniformly-typed values in a particular Dataset.
//
// TODO(mknyszek): Consider relaxing columns and allow them to be used on other
// datasets, provided the column name and type match.
type Column[T any] struct {
	d *Dataset
	i int
}

// Values returns an iterator over all the values in this column.
func (c Column[T]) Values() iter.Seq[T] {
	cd, ok := c.d.columns[c.i].(*columnData[T])
	if !ok {
		panic(fmt.Sprintf("column type does not match %T", *new(T)))
	}
	return func(yield func(T) bool) {
		for _, v := range cd.values {
			if !yield(v) {
				return
			}
		}
	}
}

// Row represents a single row in a Dataset. Its fields
// may be accessed and mutated with Field and SetField.
type Row struct {
	d *Dataset
	i int
}

// Rows returns the number of rows in the dataset.
func (d *Dataset) Rows() int {
	return d.rows
}

// Rows returns an iterator over all the rows in the dataset.
func (d *Dataset) All() iter.Seq[Row] {
	return func(yield func(Row) bool) {
		for i := range d.rows {
			if !yield(Row{d, i}) {
				return
			}
		}
	}
}

// Field returns the entry of the provided row corresponding
// to the provided column.
//
// Panics if the column is not from this dataset.
func Field[T any](r Row, c Column[T]) T {
	if c.d != r.d {
		panic("attempted to access column in record for incorrect dataset")
	}
	cd, ok := r.d.columns[c.i].(*columnData[T])
	if !ok {
		panic(fmt.Sprintf("column type does not match %T", *new(T)))
	}
	return cd.values[r.i]
}

// SetField writes value into the entry of the provided row
// corresponding to the provided column.
//
// Panics if the column is not from this dataset.
func SetField[T any](r Row, c Column[T], value T) {
	if c.d != r.d {
		panic("attempted to access column in record for incorrect dataset")
	}
	cd, ok := r.d.columns[c.i].(*columnData[T])
	if !ok {
		panic(fmt.Sprintf("column type does not match %T: column type %T", *new(T), c))
	}
	cd.values[r.i] = value
}
