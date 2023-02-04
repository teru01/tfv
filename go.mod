module github.com/teru01/tfv

go 1.16

require (
	github.com/stretchr/testify v1.8.1
	github.com/teru01/tfv/core v0.0.0-00010101000000-000000000000
	github.com/urfave/cli v1.22.12
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/teru01/tfv/core => ./
