// Package main provides siemsend, a UNIX philosophy-inspired SIEM connector
// that reads JSONL from stdin and sends it to various SIEM backends.
package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{Use: "siemsend"}
	logger  *slog.Logger
)

// GenericOutput interface for all outputs for possible multi-output option
type GenericOutput interface {
	Send(string)
	Init() error
}

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
