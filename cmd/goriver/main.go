package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"go.uber.org/fx"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

// sharedFlags consists of anything common among commands.
type sharedFlags struct {
	databaseURL string `help:"The postgres database to connect to"`
}

type cli struct {
	sharedFlags

	Client clientCmd `cmd:"" help:"run the client"`
	Worker workerCmd `cmd:"" help:"run the worker"`
}

func run() error {
	var cli cli
	kctx := kong.Parse(
		&cli,
		kong.Description("river worker and client in one"),
		kong.ShortUsageOnError(),
		kong.DefaultEnvars("GORIVER"),
		kong.ConfigureHelp(kong.HelpOptions{
			Tree:     true,
			Compact:  true,
			Indenter: kong.TreeIndenter,
		}),
	)

	if err := kctx.Run(cli.sharedFlags); err != nil {
		slog.Error("something happened", slog.Any("error", err))
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

func runApp(app *fx.App, callbackFn func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), app.StartTimeout())
	defer cancel()
	if err := app.Start(ctx); err != nil {
		return err
	}

	callbackErr := callbackFn()

	stopCtx, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
	defer cancel()
	return errors.Join(callbackErr, app.Stop(stopCtx))
}
