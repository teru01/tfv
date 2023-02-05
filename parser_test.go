package core

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectDeclaredVariables(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name string
		in   string
		out  tfVariables
	}

	cases := []testCase{
		{
			name: "collect nothing",
			in: `
resource "aws_ecr_repository" "app" {
	name = "myapp"
}

locals {
	env = "dev"
}
`,
			out: tfVariables{},
		},
		{
			name: "collect simplest variable",
			in:   `variable "foobar" {}`,
			out: tfVariables{
				"foobar": tfVariable{
					block: `variable "foobar" {}`,
				},
			},
		},
		{
			name: "collect simple variables",
			in: `
variable "region" {
	description = "region"

	default     = "ap-northeast-1"
}

variable "environment" {
	description = "foo"
}
`,
			out: tfVariables{
				"region": tfVariable{
					block: `variable "region" {
	description = "region"

	default     = "ap-northeast-1"
}`,
				},
				"environment": tfVariable{
					block: `variable "environment" {
	description = "foo"
}`,
				},
			},
		},
		{
			name: "collect simple with comment",
			in: `
# variable "region" {
# 	description = "region"
#
# 	default     = "ap-northeast-1"
# }

variable "environment" {
	description = "foo"
}
`,
			out: tfVariables{
				"environment": tfVariable{
					block: `variable "environment" {
	description = "foo"
}`,
				},
			},
		},
		{
			name: "collect nested block variables",
			in: `
variable "foo" {
	type = object({
		min_capacity           = number
		peak_time_min_capacity = number
	})
}`,
			out: tfVariables{
				"foo": tfVariable{
					block: `variable "foo" {
	type = object({
		min_capacity           = number
		peak_time_min_capacity = number
	})
}`,
				},
			},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := collectDeclaredVariables(bufio.NewReader(strings.NewReader(tt.in)))
			assert.NoError(t, err, tt.name)
			assert.Equal(t, tt.out, result, tt.name)
		})
	}
}

func TestCollectUsedVars(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name string
		in   string
		out  map[string]struct{}
	}

	cases := []testCase{
		{
			name: "collect nothing",
			in: `locals {
			env = "dev"
		}

		# Firehose
		resource "aws_cloudwatch_log_group" "sample" {
			name              = "/aws/kinesisfirehose/myvar"
			retention_in_days = 20
		}`,
			out: map[string]struct{}{},
		},
		{
			name: "collect simple",
			in: `locals {
	env = "dev"
}
	
# Firehose var.hose
resource "aws_cloudwatch_log_group" "sample" {
	name              = "/aws/kinesisfirehose/myvar/${var.hoge}"
	retention_in_days = var.retention_days
	value = foovar.value
}`,
			out: map[string]struct{}{
				"hoge":           {},
				"retention_days": {},
			},
		},
		{
			name: "collect only ${var.foo} style in quote",
			in: `locals {
			env = "dev"
		}

		# Firehose
		resource "aws_cloudwatch_log_group" "sample" {
			url              = lookup(var.somemap, "http://semvar.co.jp/var.orange/${var.hoge}/${var.foo}/%{ if var.name != "john" }$${var.my-first-name} = ${var.first_name}%{ else }unnamed%{ endif }")
		}`,
			out: map[string]struct{}{
				"somemap":    {},
				"hoge":       {},
				"foo":        {},
				"first_name": {},
				// var.name can not be collected for now.
			},
		},
		{
			name: "collect with heredoc",
			in: `
resource "aws_cloudwatch_log_group" "sample" {
	name              = "/aws/kinesisfirehose/myvar/${var.hoge}"
	retention_in_days = var.retention_days
	query = <<EOF
		inside var.heredoc_bare1 ${var.heredoc_parentheses2}
		inside var.heredoc_bare3 ${var.heredoc_parentheses4}
EOF
	num = var.num
	word = <<-EOL
inside var.eol ${var.heredoc_eol}
EOL
	count = var.count
}`,
			out: map[string]struct{}{
				"hoge":                 {},
				"retention_days":       {},
				"num":                  {},
				"heredoc_parentheses2": {},
				"heredoc_parentheses4": {},
				"heredoc_eol":          {},
				"count":                {},
			},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := collectUsedVariables(bufio.NewReader(strings.NewReader(tt.in)))
			assert.NoError(t, err)
			assert.Equal(t, tt.out, result)
		})
	}
}
