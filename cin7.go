package main

import (
	"context"
	"os"

	"github.com/sshaw/cin7/cmd/location"
	"github.com/sshaw/cin7/cmd/product"
	"github.com/sshaw/cin7/cmd/stock"
	"github.com/sshaw/cin7/cmd/webhook"
	"github.com/urfave/cli/v3"
)

const version = "0.0.1"

func main() {
	app := &cli.Command{
		Name:                   "cin7",
		Usage:                  "Utilities for working with Cin7",
		Version:                version,
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			&location.Cmd,
			&product.Cmd,
			&stock.Cmd,
			&webhook.Cmd,
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}
