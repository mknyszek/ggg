package csv

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/mknyszek/ggg"
)

func Read(r io.Reader, options ...ReadOption) (*ggg.Dataset, error) {
	var opts readOptions
	for _, opt := range DefaultReadOptions {
		opt.set(&opts)
	}
	for _, opt := range options {
		opt.set(&opts)
	}
	d := ggg.Empty()
	cn := 0
	line := 1
	dataLine := 0
	newLine := true
	expectColumns := 0
	var dColumns []ggg.Column[string]
	for tok, err := range tokens(r, &opts) {
		if err == io.EOF {
			if line > 1 && cn > 0 && cn != expectColumns {
				return nil, fmt.Errorf("line %d: expected %d columns, found %d", line, expectColumns, cn)
			}
			break
		}
		if err != nil {
			return nil, err
		}
		switch tok {
		case string(opts.separator):
			// Nothing to do.
		case "\n":
			if expectColumns == 0 {
				expectColumns = cn
			} else if cn != expectColumns {
				return nil, fmt.Errorf("line %d: expected %d columns, found %d", line, expectColumns, cn)
			}
			cn = 0
			if line > 1 || !opts.header {
				dataLine++
			}
			line++
			newLine = true
		default:
			cn++
			if line == 1 {
				var name string
				if opts.header {
					name = tok
				} else {
					name = strconv.Itoa(cn)
				}
				col := ggg.NewColumn[string](name)
				dColumns = append(dColumns, col)
				d.AddColumn(col)
				if opts.header {
					// It's a header, so don't set it as data.
					break
				}
			}
			if newLine {
				d.Grow(1)
				newLine = false
			}
			dColumns[cn-1].Set(d, dataLine, tok)
		}
	}
	return d, nil
}

type readOptions struct {
	separator rune
	comment   string
	comment0  rune
	header    bool
}

var DefaultReadOptions = []ReadOption{
	Separator(','),
	Header(),
}

type ReadOption interface {
	set(*readOptions)
}

func tokens(r io.Reader, opts *readOptions) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		tokenize(r, opts, yield)
	}
}

func tokenize(rd io.Reader, opts *readOptions, yield func(string, error) bool) {
	var r io.RuneReader
	if rr, ok := rd.(io.RuneReader); ok {
		r = rr
	} else {
		r = bufio.NewReader(rd)
	}
	var field strings.Builder
	for {
		c, _, err := r.ReadRune()
		if err != nil {
			yield("", err)
			return
		}
		switch c {
		case opts.separator, '\n':
			if !yield(field.String(), nil) {
				return
			}
			field.Reset()
			if !yield(string(c), nil) {
				return
			}
		case '\r':
			continue
		case opts.comment0:
			if field.Len() != 0 {
				if !yield(field.String(), nil) {
					return
				}
				field.Reset()
			}
			// Look ahead to see if we match the comment.
			i := utf8.RuneLen(opts.comment0)
			var lastC rune
			for i < len(opts.comment) {
				c, n, err := r.ReadRune()
				if err != nil {
					yield("", err)
					return
				}
				if string(c) != opts.comment[i:i+n] {
					lastC = c
					break
				}
				i += n
			}
			if i != len(opts.comment) {
				field.WriteString(opts.comment[:i])
				field.WriteRune(lastC)
				break
			}
			for {
				c, _, err := r.ReadRune()
				if err != nil {
					yield("", err)
					return
				}
				if c == '\n' {
					break
				}
			}
		default:
			field.WriteRune(c)
		}
	}
}

func Separator(r rune) ReadOption {
	return readOption{
		do: func(opts *readOptions) {
			opts.separator = r
		},
	}
}

func Header() ReadOption {
	return readOption{
		do: func(opts *readOptions) {
			opts.header = true
		},
	}
}

func NoHeader() ReadOption {
	return readOption{
		do: func(opts *readOptions) {
			opts.header = false
		},
	}
}

func CommentPrefix(s string) ReadOption {
	return readOption{
		do: func(opts *readOptions) {
			opts.comment = s
			c0, _ := utf8.DecodeRuneInString(s)
			opts.comment0 = c0
		},
	}
}

type readOption struct {
	do func(*readOptions)
}

func (r readOption) set(opts *readOptions) {
	r.do(opts)
}
