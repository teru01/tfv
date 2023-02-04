package main

import (
	"log"
	"os"

	"github.com/teru01/tfv/core"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:   "tfv",
		Usage:  "generate Terraform variables declaration",
		Action: core.PrintVariables,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
}
