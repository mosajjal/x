package main // Changed package to main

import (
	"context"
	"io/fs" // Import io/fs
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

// Server represents the HTTP server for serving CyberChef
type Server struct {
	srv  *http.Server
	addr string
}

// NewServer creates a new HTTP server instance
func NewServer(addr string, cyberchefAssets fs.FS) *Server { // Accept fs.FS
	mux := http.NewServeMux()

	// Register handlers
	fs := http.FileServer(http.FS(cyberchefAssets)) // Use the passed fs.FS
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Redirect root to index.html if path is "/"
		// Need to adjust the path to match the structure within Cyberchef directory
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			r.URL.Path = "/CyberChef_v10.19.4.html"
		} else if r.URL.Path == "/favicon.ico" {
			// Serve the favicon from the correct path within the embedded FS
			r.URL.Path = "/assets/aecc661b69309290f600.ico"
		}
		// Ensure the path is prefixed with "Cyberchef" as per the embed directive
		// However, http.FS serves from the root of the embedded FS,
		// and the embed directive `//go:embed ../../Cyberchef` makes `Cyberchef` the root.
		// So, direct access to files within `Cyberchef` should work.
		// Let's verify the paths after attempting to build.
		fs.ServeHTTP(w, r)
	}))

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
