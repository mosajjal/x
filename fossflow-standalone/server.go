package main

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

// Server represents the HTTP server for serving OpenFLOW
type Server struct {
	srv  *http.Server
	addr string
}

// NewServer creates a new HTTP server instance
func NewServer(addr string, openflowAssets fs.FS) *Server {
	mux := http.NewServeMux()

	// Register handlers
	fs := http.FileServer(http.FS(openflowAssets))
	mux.Handle("/", fs)

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		srv:  srv,
		addr: addr,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

// OpenBrowser opens the default browser to the specified URL
func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		log.Printf("Unsupported platform: %s\n", runtime.GOOS)
		return
	}

	if err != nil {
		log.Printf("Error opening browser: %v\n", err)
	}
}
