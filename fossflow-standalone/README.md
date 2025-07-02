# fossflow Standalone

A Go CLI tool that embeds fossflow and serves it as a HTTP single-page application (SPA).

## Features

- Embeds fossflow HTML and assets directly into the binary
- Serves fossflow as a local web application
- Provides command-line options for customization
- Works offline, no internet connection required

## Installation

```bash
# Clone this repository
git clone github.com/mosajjal/x

# Change to the directory
cd x/fossflow-standalone

# run Docker build and export the binary
docker build --output . .

## Usage

```bash
# Start the server with default settings (localhost:8080)
./fossflow-standalone serve

# Specify a different port
./fossflow-standalone serve --port 3000

# Specify a different host
./fossflow-standalone serve --host 0.0.0.0

# Open browser automatically
./fossflow-standalone serve --open

# Get help
./fossflow-standalone help
./fossflow-standalone serve --help
```

## How It Works

This tool uses Go's `embed` package to embed the fossflow files directly into the binary. When you run the serve command, it starts an HTTP server that serves the embedded files.

## License

This project is licensed under the MIT License. Note that fossflow itself is licensed under Apache-2.0 License.
