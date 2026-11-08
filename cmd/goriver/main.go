package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

type cli struct {
	Client clientCmd `help:"run the client" cmd:""`
	Worker workerCmd `help:"run the worker" cmd:""`
}

// func (c *cli) Run() error {
// 	slog.Info("hello")
// 	return nil
// }

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

	if err := kctx.Run(); err != nil {
		slog.Error("something happened", slog.Any("error", err))
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}
