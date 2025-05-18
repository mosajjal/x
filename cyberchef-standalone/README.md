# CyberChef Standalone

A Go CLI tool that embeds CyberChef and serves it as a HTTP single-page application (SPA).

## Features

- Embeds CyberChef HTML and assets directly into the binary
- Serves CyberChef as a local web application
- Provides command-line options for customization
- Works offline, no internet connection required

## Installation

```bash
# Clone this repository
git clone github.com/mosajjal/x

# Change to the directory
cd x/cyberchef-standalone

# Download the latest CyberChef release and extract it to the `Cyberchef` directory
wget https://github.com/gchq/CyberChef/releases/download/v10.19.4/CyberChef_v10.19.4.zip

unzip CyberChef_v10.19.4.zip -d Cyberchef

# Build the binary
cd cyberchef-standalone
go build -o cyberchef-standalone
```

## Usage

```bash
# Start the server with default settings (localhost:8080)
./cyberchef-standalone serve

# Specify a different port
./cyberchef-standalone serve --port 3000

# Specify a different host
./cyberchef-standalone serve --host 0.0.0.0

# Open browser automatically
./cyberchef-standalone serve --open

# Get help
./cyberchef-standalone help
./cyberchef-standalone serve --help
```

## How It Works

This tool uses Go's `embed` package to embed the CyberChef files directly into the binary. When you run the serve command, it starts an HTTP server that serves the embedded files.

## License

This project is licensed under the MIT License. Note that CyberChef itself is licensed under Apache-2.0 License.
