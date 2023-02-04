package core

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractVars(t *testing.T) {
	t.Skip()
	testcase := []struct {
		input    string
		expected []string
	}{
		{
			input:    `var.hoge`,
			expected: []string{"hoge"},
		},
		{
			input: `resource "aws_lb_listener" "https" {
	load_balancer_arn = var.foo_hoge
}`,
			expected: []string{"foo_hoge"},
		},
		{
			input: `resource "aws_lb_listener" "https" {
	load_balancer_arn = var.foo_hoge
	default_action {
		type = "forward"
		target_group_arn = var.bar_hoge
	}
}`,
			expected: []string{"foo_hoge", "bar_hoge"},
		},
		{
			input: `resource "aws_lb_listener" "https" {
	load_balancer_arn = ${var.foo_hoge} + "_" + var.bar_hoge[0]
	`,
			expected: []string{"foo_hoge", "bar_hoge"},
		},
	}
	for _, tc := range testcase {
		input := bufio.NewReader(bytes.NewBuffer([]byte(tc.input)))
		vars, err := extractVars(input)
		require.NoError(t, err)
		require.Equal(t, tc.expected, vars)
	}
}
