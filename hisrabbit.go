package hisrabbit

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	LogFormat string `long:"log-format" choice:"text" choice:"json" default:"text" description:"Log format"`
	Verbose   []bool `short:"v" long:"verbose" description:"Show verbose debug information, each -v bumps log level"`
	logLevel  slog.Level
}

var parser *flags.Parser

func Execute() int {
	parser = flags.NewParser(&opts, flags.HelpFlag)

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			parser.WriteHelp(os.Stdout)
			return 0
		}

		parser.WriteHelp(os.Stderr)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		return 1
	}

	if err := setLogLevel(); err != nil {
		slog.Error("error setting log level", "error", err)
		return 1
	}

	if err := setupLogger(); err != nil {
		slog.Error("error setting up logger", "error", err)
		return 1
	}

	if err := run(); err != nil {
		slog.Error("run failed", "error", err)
		return 1
	}

	return 0
}

func run() error {
	slog.Debug("Debug", "currrent level", opts.logLevel)
	slog.Info("Info", "currrent level", opts.logLevel)
	slog.Warn("Warn", "currrent level", opts.logLevel)
	slog.Error("Error", "currrent level", opts.logLevel)

	return nil
}
