package main

import (
	"fmt"
	"log"
	"os"

	"github.com/teru01/tfv/core"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Value: ".",
				Usage: "tfv collects variables from `DIR`",
			},
			&cli.BoolFlag{
				Name:  "sync",
				Usage: "execute in sync mode (tfv generates variables without unused variables)",
			},
			&cli.StringFlag{
				Name:  "tfvars-file",
				Usage: "load tfvars from `FILE`",
			},
			&cli.StringFlag{
				Name:  "suffix",
				Value: ".generated",
				Usage: "suffix of generated files",
			},
		},
		Name:  "tfv",
		Usage: "Terraform variables generator",
		Action: func(ctx *cli.Context) error {
			variables, tfvars, err := core.GenerateVariables(ctx)
			if err != nil {
				return err
			}
			fmt.Println(variables)
			fmt.Println(tfvars)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
}
