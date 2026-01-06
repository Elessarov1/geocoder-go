package main

import (
	"Geocoder/cmd/start"
	"Geocoder/internal/common/version"
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/urfave/cli/v3"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	app := &cli.Command{
		Name:    "geocoder",
		Usage:   "geocoder",
		Version: version.Version(),
		Commands: []*cli.Command{
			start.CmdStart(),
		},
	}

	if err := app.Run(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start application: %v\n", err)
		os.Exit(1)
	}
}
