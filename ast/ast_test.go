package ast_test

import (
	"testing"

	"github.com/kylelemons/godebug/diff"
	"github.com/protocolbuffers/txtpbfmt/ast"
	"github.com/protocolbuffers/txtpbfmt/parser"
)

func TestGetFromPath(t *testing.T) {
	content := `first {
  second {
    third: "v1"
    third: "v2"
  }
  second {
    third: "v3"
    third: "v4"
  }
}
first {
  second {
    third: "v5"
    third: "v6"
  }
  second {
    third: "v7"
    third: "v8"
  }
}
`
	inputs := []struct {
		in   string
		path []string
		want string
	}{{
		in:   content,
		path: nil,
		want: ``,
	}, {
		in:   content,
		path: []string{"first", "second", "third"},
		want: `third: "v1"
third: "v2"
third: "v3"
third: "v4"
third: "v5"
third: "v6"
third: "v7"
third: "v8"
`,
	}, {
		in:   content,
		path: []string{"first", "second"},
		want: `second {
  third: "v1"
  third: "v2"
}
second {
  third: "v3"
  third: "v4"
}
second {
  third: "v5"
  third: "v6"
}
second {
  third: "v7"
  third: "v8"
}
`,
	}, {
		in:   content,
		path: []string{"first"},
		want: content,
	}}
	for _, input := range inputs {
		nodes, err := parser.Parse([]byte(input.in))
		if err != nil {
			t.Errorf("Parse %v returned err %v", input.in, err)
			continue
		}
		filtered := ast.GetFromPath(nodes, input.path)
		got := parser.Pretty(filtered, 0)
		if diff := diff.Diff(input.want, got); diff != "" {
			t.Errorf("GetFromPath %v %v returned diff (-want, +got):\n%s", input.in, input.path, diff)
		}
	}
}

func TestIsCommentOnly(t *testing.T) {
	inputs := []struct {
		in   string
		want []bool
	}{{
		in: `foo: 1
bar: 2`,
		want: []bool{false, false},
	},
		{
			in: `foo: 1
bar: 2
`,
			want: []bool{false, false},
		},
		{
			in: `foo: 1
bar: 2
# A long trailing comment
# over multiple lines.
`,
			want: []bool{false, false, true},
		},
		{
			in: `first {
  foo: true  # bar
}
`,
			want: []bool{false},
		},
		{
			in: `first {
  foo: true  # bar
}
# trailing comment
`,
			want: []bool{false, true},
		},
		{
			in:   `{}`,
			want: []bool{false},
		},
	}
	for _, input := range inputs {
		nodes, err := parser.Parse([]byte(input.in))
		if err != nil {
			t.Errorf("Parse %v returned err %v", input.in, err)
			continue
		}
		if len(nodes) != len(input.want) {
			t.Errorf("For %v, expect %v nodes, got %v", input.in, len(input.want), len(nodes))
		}
		for i, n := range nodes {
			if got := n.IsCommentOnly(); got != input.want[i] {
				t.Errorf("For %v, nodes[%v].IsCommentOnly() = %v, want %v", input.in, i, got, input.want[i])
			}
		}
	}
}

func TestFixInline(t *testing.T) {
	content := `first { }`

	inputs := []struct {
		in   string
		add  string
		want string
	}{{
		in:  content,
		add: "foo: true  # bar",
		want: `first {
  foo: true  # bar
}
`,
	}, {
		in: content,
		add: `
			# bar
			foo: true`,
		want: `first {
  # bar
  foo: true
}
`,
	}, {
		in: content,
		add: `
			# bar
			foo: true  # baz`,
		want: `first {
  # bar
  foo: true  # baz
}
`,
	}, {
		in: content,
		add: `
			foo {
				bar: true
			}`,
		want: `first {
  foo {
    bar: true
  }
}
`,
	}, {
		in:  content,
		add: `foo { bar: { baz: true } zip: "foo" }`,
		want: `first { foo { bar: { baz: true } zip: "foo" } }
`,
	}, {in: `foo {}`, add: ``, want: `foo {}
`}, {in: `foo {
}`, add: ``, want: `foo {
}
`}, {in: `foo <>`, add: ``, want: `foo {}
`}}
	for _, input := range inputs {
		nodes, err := parser.Parse([]byte(input.in))
		if err != nil {
			t.Errorf("Parse %v returned err %v", input.in, err)
			continue
		}
		if len(nodes) == 0 {
			t.Errorf("Parse %v returned no nodes", input.in)
			continue
		}
		add, err := parser.Parse([]byte(input.add))
		if err != nil {
			t.Errorf("Parse %v returned err %v", input.in, err)
			continue
		}
		nodes[0].Children = add
		nodes[0].Fix()
		got := parser.Pretty(nodes, 0)
		if diff := diff.Diff(input.want, got); diff != "" {
			t.Errorf("adding %v %v returned diff (-want, +got):\n%s", input.in, input.add, diff)
		}
	}
}
