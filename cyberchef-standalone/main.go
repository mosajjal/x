// Package main is the entry point for the CyberChef Standalone application.
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

//go:embed all:Cyberchef
var embeddedFS embed.FS

var cyberchefFS fs.FS

// NewRootCmd creates the root command for the CLI application
func NewRootCmd(cyberchefFS fs.FS) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cyberchef-standalone",
		Short: "A CLI tool that serves CyberChef as a standalone web application",
		Long: `CyberChef Standalone is a Go CLI tool that embeds CyberChef and serves it as a web application.
It allows you to run CyberChef locally without needing an internet connection.`,
	}

	rootCmd.AddCommand(NewServeCmd(cyberchefFS)) // Call NewServeCmd directly
	return rootCmd
}

// NewServeCmd creates a new serve command for the CLI application
func NewServeCmd(cyberchefFS fs.FS) *cobra.Command {
	var port int
	var host string
	var open bool

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve CyberChef as a web application",
		Long:  `Start a web server and serve CyberChef as a single-page application.`,
		RunE: func(_ *cobra.Command, _ []string) error { // Mark unused parameters
			addr := fmt.Sprintf("%s:%d", host, port)
			srv := NewServer(addr, cyberchefFS) // Call NewServer directly

			// Set up channel to listen for interrupt signal
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			// Start server in a goroutine
			errChan := make(chan error, 1)
			go func() {
				fmt.Printf("Starting CyberChef server at http://%s\n", addr)
				if open {
					OpenBrowser(fmt.Sprintf("http://%s", addr)) // Call OpenBrowser directly
				}
				errChan <- srv.Start()
			}()

			// Wait for interrupt signal or server error
			select {
			case err := <-errChan:
				return err
			case <-sigChan:
				fmt.Println("\nShutting down server...")
				return srv.Shutdown()
			}
		},
	}

	// Add flags
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to serve on")
	serveCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Host to serve on")
	serveCmd.Flags().BoolVarP(&open, "open", "o", false, "Open browser automatically")

	return serveCmd
}

func init() {
	var err error
	cyberchefFS, err = fs.Sub(embeddedFS, "Cyberchef")
	if err != nil {
		log.Fatalf("Failed to get Cyberchef subfolder from embedded filesystem: %v", err)
	}
}

func main() {
	rootCmd := NewRootCmd(cyberchefFS) // Call NewRootCmd directly
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
