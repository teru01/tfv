package core

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildTfVars(t *testing.T) {
	t.Parallel()
	type input struct {
		file         string
		keysToDelete map[string]struct{}
	}
	type testCase struct {
		name string
		in   input
		out  string
	}

	cases := []testCase{
		{
			name: "nothing to delete",
			in: input{
				file: `
foo = "bar"
hoge = 3
qwerty = [
	"aaa",
	"bbb",
]
`,
				keysToDelete: map[string]struct{}{},
			},
			out: `
foo = "bar"
hoge = 3
qwerty = [
	"aaa",
	"bbb",
]
`,
		},
		{
			name: "simple delete",
			in: input{
				file: `
foo = "bar"
hoge = 3
qwerty = [
	"aaa",
	"bbb",
]
`,
				keysToDelete: map[string]struct{}{"hoge": {}},
			},
			out: `
foo = "bar"
qwerty = [
	"aaa",
	"bbb",
]
`,
		},
		{
			name: "delete with comments",
			in: input{
				file: `
foo = "bar"
hoge = 3
qwerty = [
	"aaa",
	"bbb",
]
my_config = {
	count = 3
	enabled = true
	version = "1.4.0"
}

# something = {
# 	editer = "vscode"
# }

task_count = 200
`,
				keysToDelete: map[string]struct{}{"foo": {}, "hoge": {}, "my_config": {}},
			},
			out: `
qwerty = [
	"aaa",
	"bbb",
]

# something = {
# 	editer = "vscode"
# }

task_count = 200
`,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := buildTfVars(bufio.NewReader(strings.NewReader(tt.in.file)), tt.in.keysToDelete)
			assert.NoError(t, err, tt.name)
			assert.Equal(t, tt.out, result, tt.name)
		})
	}
}
