package csv

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mknyszek/ggg"
)

func TestRead(t *testing.T) {
	type expectation struct {
		column   ggg.Column[string]
		contents []string
	}
	type test struct {
		name    string
		input   string
		opts    []ReadOption
		expect  []expectation
		errLike string
	}
	for _, ts := range []test{
		{
			name:   "Empty",
			input:  "",
			expect: []expectation{},
		},
		{
			name: "OneColumn",
			input: `a
x
0
1
`,
			opts: []ReadOption{NoHeader(), Separator(',')},
			expect: []expectation{
				{ggg.NewColumn[string]("1"), []string{"a", "x", "0", "1"}},
			},
		},
		{
			name: "OneColumnHeader",
			input: `a
x
0
1
`,
			opts: []ReadOption{Header(), Separator(',')},
			expect: []expectation{
				{ggg.NewColumn[string]("a"), []string{"x", "0", "1"}},
			},
		},
		{
			name: "OneColumnComment",
			input: `// Zero is the hero.
a
// I am a comment.
x
0
// And I am one more.
1// Inline comment.
`,
			opts: []ReadOption{NoHeader(), Separator(','), CommentPrefix("//")},
			expect: []expectation{
				{ggg.NewColumn[string]("1"), []string{"a", "x", "0", "1"}},
			},
		},
		{
			name: "OneColumnHeaderComment",
			input: `// Zero is the hero.
a
// I am a comment.
x
0
// And I am one more.
1// Inline comment.
`,
			opts: []ReadOption{Header(), Separator(','), CommentPrefix("//")},
			expect: []expectation{
				{ggg.NewColumn[string]("a"), []string{"x", "0", "1"}},
			},
		},
		{
			name: "Comma",
			input: `a,b,c
x,y,z
`,
			opts: []ReadOption{NoHeader(), Separator(',')},
			expect: []expectation{
				{ggg.NewColumn[string]("1"), []string{"a", "x"}},
				{ggg.NewColumn[string]("2"), []string{"b", "y"}},
				{ggg.NewColumn[string]("3"), []string{"c", "z"}},
			},
		},
		{
			name: "CommaHeader",
			input: `a,b,c
x,y,z
`,
			opts: []ReadOption{Header(), Separator(',')},
			expect: []expectation{
				{ggg.NewColumn[string]("a"), []string{"x"}},
				{ggg.NewColumn[string]("b"), []string{"y"}},
				{ggg.NewColumn[string]("c"), []string{"z"}},
			},
		},
		{
			name: "Tab",
			input: `a	b	c
x	y	z
`,
			opts: []ReadOption{NoHeader(), Separator('\t')},
			expect: []expectation{
				{ggg.NewColumn[string]("1"), []string{"a", "x"}},
				{ggg.NewColumn[string]("2"), []string{"b", "y"}},
				{ggg.NewColumn[string]("3"), []string{"c", "z"}},
			},
		},
		{
			name: "TabHeader",
			input: `ay	bee	see
x	y	z
`,
			opts: []ReadOption{Header(), Separator('\t')},
			expect: []expectation{
				{ggg.NewColumn[string]("ay"), []string{"x"}},
				{ggg.NewColumn[string]("bee"), []string{"y"}},
				{ggg.NewColumn[string]("see"), []string{"z"}},
			},
		},
		{
			name: "TabComment",
			input: `// Haha.
ay	bee	see
// Comment.
x	y	z
1	2	3// Inline comment.
`,
			opts: []ReadOption{NoHeader(), Separator('\t'), CommentPrefix("//")},
			expect: []expectation{
				{ggg.NewColumn[string]("1"), []string{"ay", "x", "1"}},
				{ggg.NewColumn[string]("2"), []string{"bee", "y", "2"}},
				{ggg.NewColumn[string]("3"), []string{"see", "z", "3"}},
			},
		},
		{
			name: "TabCommentHeader",
			input: `// Haha.
a	b	c
// Comment.
x	y	z
1	2	3// Inline comment.
`,
			opts: []ReadOption{Header(), Separator('\t'), CommentPrefix("//")},
			expect: []expectation{
				{ggg.NewColumn[string]("a"), []string{"x", "1"}},
				{ggg.NewColumn[string]("b"), []string{"y", "2"}},
				{ggg.NewColumn[string]("c"), []string{"z", "3"}},
			},
		},
	} {
		t.Run(ts.name, func(t *testing.T) {
			d, err := Read(strings.NewReader(ts.input), ts.opts...)
			if err != nil {
				if ts.errLike != "" && !strings.Contains(err.Error(), ts.errLike) {
					t.Errorf("expected failure containing %q, got %v:", ts.errLike, err)
				} else {
					t.Errorf("unexpectedly failed to parse: %v", err)
				}
			} else {
				if d.Columns() != len(ts.expect) {
					t.Errorf("expected %d columns, got %d", len(ts.expect), d.Columns())
				}
				for _, e := range ts.expect {
					if !e.column.In(d) {
						t.Errorf("expected column %s in dataset, but not found", e.column)
						continue
					}
					if len(e.contents) != d.Rows() {
						t.Errorf("dataset has a different number of rows than expected: want %d, got %d", len(e.contents), d.Rows())
						continue
					}
					for i, want := range e.contents {
						if got := e.column.Get(d, i); got != want {
							t.Errorf("[col %s, row %d]: expected %q but found %q", e.column, i, want, got)
						}
					}
				}
			}
			if t.Failed() {
				var buf bytes.Buffer
				if err := d.Print(&buf); err != nil {
					t.Fatal(err)
				}
				t.Logf("Parse result:\n%s", buf.String())
			}
		})
	}
}
