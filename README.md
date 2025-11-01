# Go Experiments Collection

[![Test and Lint](https://github.com/mosajjal/go-exp/actions/workflows/test.yml/badge.svg)](https://github.com/mosajjal/go-exp/actions/workflows/test.yml)
[![Build Release](https://github.com/mosajjal/go-exp/actions/workflows/release.yml/badge.svg)](https://github.com/mosajjal/go-exp/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://go.dev/doc/devel/release)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A curated collection of production-ready Go tools and utilities for security, networking, DevOps, and data operations. All projects follow modern Go 1.25+ idioms with structured logging (slog), comprehensive documentation, and automated testing.

## 🚀 Quick Start

Each tool is a self-contained Go module that can be installed independently:

```bash
# Clone the repository
git clone https://github.com/mosajjal/go-exp.git
cd go-exp

# Navigate to any project and build
cd <project-name>
go build

# Or install directly
go install github.com/mosajjal/go-exp/<project-name>@latest
```

## 📦 Projects Overview

### Security & Forensics

#### [allhash](./allhash)
Multi-hash calculator for forensic analysis and file verification.
- **Features**: MD5, SHA1, SHA256, TLSH, ssdeep support
- **Use Cases**: Digital forensics, file integrity verification, malware analysis
- **Status**: Production Ready

#### [redlineparser](./redlineparser)
Parser for Redline memory analysis output files.
- **Features**: Memory dump analysis, artifact extraction
- **Use Cases**: Incident response, memory forensics
- **Status**: Production Ready

#### [grizz](./grizz)
High-performance IPv4 lookup tool with extensive database support.
- **Features**: Fast IP geolocation, ASN lookup, threat intelligence
- **Use Cases**: Network security analysis, threat hunting
- **Status**: Production Ready

### Networking & Proxies

#### [fakedns](./fakedns)
Programmable DNS server with rule-based query responses.
- **Features**: CSV rule files, prefix/suffix/FQDN matching, upstream DNS fallback
- **Use Cases**: DNS testing, development environments, traffic routing
- **Status**: Production Ready

#### [gopherchain](./gopherchain)
ProxyChains implementation using tun2socks and network namespaces.
- **Features**: Transparent proxying, chain multiple proxies, network isolation
- **Use Cases**: Privacy, penetration testing, traffic anonymization
- **Status**: Production Ready

#### [sniplex](./sniplex)
TLS/MTProto SNI-based router without certificate hosting.
- **Features**: SNI inspection, transparent routing, no TLS termination
- **Use Cases**: Traffic routing, protocol multiplexing
- **Status**: Production Ready

#### [sockstls](./sockstls)
SOCKS5 proxy over TLS with authentication support.
- **Features**: TLS encryption, basic auth, SOCKS5 protocol
- **Use Cases**: Secure proxy access, encrypted tunneling
- **Status**: Production Ready

#### [sshproxy](./sshproxy)
Dedicated SSH server for port forwarding only.
- **Features**: SSH tunneling, port forwarding, no shell access
- **Use Cases**: Secure tunneling, bastion host alternative
- **Status**: Beta

#### [sctptunnel](./sctptunnel)
TCP over SCTP tunnel implementation.
- **Features**: SCTP protocol support, multi-streaming
- **Use Cases**: Network resilience, mobile networks
- **Status**: Work In Progress

#### [proxyenv](./proxyenv)
Environment-configured HTTP proxy with allowlist support.
- **Features**: Domain/IP allowlisting, zero-config via env vars
- **Use Cases**: Egress control, corporate proxy
- **Status**: Production Ready

### Data Operations

#### [elasticdump](./elasticdump)
Elasticsearch cluster data exporter with filtering.
- **Features**: Bulk export, index filtering, parallel downloads
- **Use Cases**: Elasticsearch backup, data migration
- **Status**: Production Ready

#### [siemsend](./siemsend)
Unix-pipeline SIEM connector supporting multiple backends.
- **Features**: Azure Sentinel, Splunk, Elasticsearch, OpenSearch
- **Use Cases**: Log aggregation, SIEM integration
- **Status**: Production Ready

#### [pebble-cli](./pebble-cli)
Command-line interface for PebbleDB key/value operations.
- **Features**: Full CRUD operations, batch processing
- **Use Cases**: Database management, debugging
- **Status**: Production Ready

### Development Tools

#### [spitcurl](./spitcurl)
HTTP(S) reverse proxy that outputs curl commands.
- **Features**: Request logging as curl commands, TLS support
- **Use Cases**: Debugging, API testing, request replay
- **Status**: Production Ready

#### [identme](./identme)
Minimal ident.me service implementation.
- **Features**: IP echo service, minimal footprint
- **Use Cases**: IP discovery, service testing
- **Status**: Production Ready

#### [webhookio](./webhookio)
Simple webhook receiver and forwarder.
- **Features**: HTTP webhook handling, payload logging
- **Use Cases**: Webhook testing, CI/CD integration
- **Status**: Production Ready

### Containerization

#### [containerize](./containerize)
Tool for containerizing Go applications.
- **Features**: Docker/OCI image creation
- **Use Cases**: Deployment automation, containerization
- **Status**: Production Ready

#### [fossflow-standalone](./fossflow-standalone)
Standalone workflow engine for automation.
- **Features**: Task orchestration, workflow management
- **Use Cases**: CI/CD, automation pipelines
- **Status**: Production Ready

#### [cyberchef-standalone](./cyberchef-standalone)
Standalone CyberChef server implementation.
- **Features**: Data transformation, encoding/decoding
- **Use Cases**: Data analysis, forensics
- **Status**: Production Ready

### Network Analysis

#### [pcap-freshner](./pcap-freshner)
PCAP file timestamp adjuster.
- **Features**: Timestamp modification, packet manipulation
- **Use Cases**: PCAP analysis, timeline adjustment
- **Status**: Production Ready

#### [mhs](./mhs)
Minimal HTTP(S) server for testing and development.
- **Features**: Quick HTTP server, static file serving
- **Use Cases**: Development, testing
- **Status**: Production Ready

#### [oauth2](./oauth2)
OAuth2 testing and development utility.
- **Features**: OAuth2 flow testing, token management
- **Use Cases**: OAuth2 development, API testing
- **Status**: Production Ready

## 🏗️ Architecture & Design

### Common Features

All projects in this collection share:

- **Modern Go (1.25+)**: Utilizing the latest Go features and idioms
- **Structured Logging**: Using `log/slog` for consistent, structured logging
- **Zero Dependencies**: Minimal external dependencies where possible
- **Production Ready**: Battle-tested in production environments
- **Well Documented**: Comprehensive code comments and README files
- **Tested**: Automated testing via GitHub Actions
- **Cross-Platform**: Linux, Windows, macOS support

### Code Quality

- Automated linting with `golangci-lint`
- Comprehensive test coverage
- Race condition detection
- Static binary compilation support
- Optimized build flags for minimal binary size

## 🛠️ Development

### Prerequisites

- Go 1.25 or later
- Make (optional, for convenience commands)
- Git

### Building All Projects

```bash
# Build all projects
for dir in */; do
  if [ -f "$dir/go.mod" ]; then
    echo "Building $dir..."
    (cd "$dir" && go build -ldflags="-s -w" -trimpath)
  fi
done
```

### Testing

```bash
# Test a specific project
cd <project-name>
go test -v -race ./...

# Test all projects
for dir in */; do
  if [ -f "$dir/go.mod" ]; then
    echo "Testing $dir..."
    (cd "$dir" && go test -v -race ./...)
  fi
done
```

### Linting

```bash
# Lint a specific project
cd <project-name>
golangci-lint run

# Lint all projects
for dir in */; do
  if [ -f "$dir/go.mod" ]; then
    echo "Linting $dir..."
    (cd "$dir" && golangci-lint run)
  fi
done
```

## 📊 CI/CD Pipeline

The repository uses GitHub Actions for:

1. **Continuous Testing**: Automated tests on every push/PR
2. **Linting**: Code quality checks with golangci-lint
3. **Building**: Verification of successful compilation
4. **Release**: Multi-platform binary releases (Linux, Windows, macOS on amd64/arm64)

See [.github/workflows/](./.github/workflows/) for pipeline definitions.

## 📝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Contribution Guidelines

- Follow existing code style and patterns
- Use `log/slog` for all logging
- Add tests for new functionality
- Update documentation (code comments + README)
- Ensure all CI checks pass

## 📄 License

Each project maintains its own license. Most are MIT-licensed. See individual project directories for details.

## 🤝 Support

- **Issues**: [GitHub Issues](https://github.com/mosajjal/go-exp/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mosajjal/go-exp/discussions)

## 🌟 Acknowledgments

Special thanks to all contributors and the Go community for making these tools possible.

## 📚 Additional Resources

- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Blog](https://go.dev/blog/)

---

**Made with ❤️ and Go**
