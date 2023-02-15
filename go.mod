module github.com/teru01/tfv

go 1.16

require (
	github.com/hashicorp/hcl/v2 v2.16.0
	github.com/stretchr/testify v1.8.1
	github.com/teru01/tfv/core v0.0.0-00010101000000-000000000000
	github.com/urfave/cli/v2 v2.3.0
)

replace github.com/teru01/tfv/core => ./
