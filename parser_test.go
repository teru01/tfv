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
		out  map[string]string
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
			out: map[string]string{},
		},
		{
			name: "collect simplest variable",
			in:   `variable "foobar" {}`,
			out:  map[string]string{"foobar": `variable "foobar" {}`},
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
			out: map[string]string{
				"region": `variable "region" {
	description = "region"

	default     = "ap-northeast-1"
}`,
				"environment": `variable "environment" {
	description = "foo"
}`,
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
			out: map[string]string{
				"environment": `variable "environment" {
	description = "foo"
}`,
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
			out: map[string]string{
				"foo": `variable "foo" {
	type = object({
		min_capacity           = number
		peak_time_min_capacity = number
	})
}`,
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
	
# Firehose
resource "aws_cloudwatch_log_group" "sample" {
	name              = "/aws/kinesisfirehose/myvar/${var.hoge}"
	retention_in_days = var.retention_days
}`,
			out: map[string]struct{}{
				"hoge":           {},
				"retention_days": {},
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
