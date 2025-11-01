# FakeDNS

[![Go Reference](https://pkg.go.dev/badge/github.com/mosajjal/go-exp/fakedns.svg)](https://pkg.go.dev/github.com/mosajjal/go-exp/fakedns)

A programmable DNS server that returns static IP addresses for configured domain patterns with support for upstream DNS fallback. Ideal for development environments, testing, and traffic routing scenarios.

## Features

- 🎯 **Pattern Matching**: Support for prefix, suffix, and exact FQDN matching
- 🔄 **Upstream Fallback**: Automatic fallback to upstream DNS for unmatched domains
- 📝 **CSV Rule Files**: Simple CSV format for domain-to-IP mappings
- 🌐 **Remote Rules**: Load rules from HTTP/HTTPS URLs
- ⚡ **High Performance**: Uses Ternary Search Trees (TST) for efficient lookups
- 🔌 **Protocol Support**: Both UDP and TCP DNS servers
- 📊 **Structured Logging**: JSON-formatted logs with detailed context

## Installation

```bash
go install github.com/mosajjal/go-exp/fakedns@latest
```

Or build from source:

```bash
git clone https://github.com/mosajjal/go-exp.git
cd go-exp/fakedns
go build
```

## Usage

### Basic Usage

Start a DNS server on port 53 with upstream DNS:

```bash
fakedns -udp 53 -upstream udp://1.1.1.1:53
```

### With Rule File

Use a CSV rule file to define domain-to-IP mappings:

```bash
fakedns -udp 53 -upstream udp://8.8.8.8:53 -rule ./rules.csv
```

### TCP and UDP

Run both TCP and UDP DNS servers:

```bash
fakedns -udp 53 -tcp 53 -upstream udp://1.0.0.1:53 -rule ./rules.csv
```

### Remote Rule File

Load rules from a URL:

```bash
fakedns -udp 53 -upstream udp://1.1.1.1:53 -rule https://example.com/dns-rules.csv
```

## Rule File Format

The rule file is a CSV with three columns: `domain,type,ip`

### Types

- `fqdn`: Exact domain match
- `prefix`: Match domain and all subdomains
- `suffix`: Match all domains ending with the pattern

### Example rules.csv

```csv
example.com,fqdn,192.168.1.100
google,prefix,10.0.0.1
.internal,suffix,172.16.0.1
test.local,fqdn,127.0.0.1
dev.,prefix,192.168.100.50
```

### Matching Examples

Given the rules above:

| Query | Match Type | Result |
|-------|-----------|--------|
| `example.com` | FQDN | `192.168.1.100` |
| `google.com` | Prefix | `10.0.0.1` |
| `google.co.uk` | Prefix | `10.0.0.1` |
| `app.internal` | Suffix | `172.16.0.1` |
| `test.local` | FQDN | `127.0.0.1` |
| `dev.example.com` | Prefix | `192.168.100.50` |
| `unknown.com` | None | Upstream DNS |

## Configuration

### Command Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-udp` | `53` | UDP port to listen on (0 to disable) |
| `-tcp` | `0` | TCP port to listen on (0 to disable) |
| `-upstream` | `udp://1.0.0.1:53` | Upstream DNS server URL |
| `-rule` | `""` | Rule file path or URL |

### Upstream DNS Formats

- `udp://1.1.1.1:53` - Standard DNS over UDP
- `tcp://1.1.1.1:53` - DNS over TCP
- `tls://1.1.1.1:853` - DNS over TLS (DoT)
- `https://1.1.1.1/dns-query` - DNS over HTTPS (DoH)

## Use Cases

### Development Environment

Route development domains to local services:

```csv
*.dev.local,suffix,127.0.0.1
api.staging,prefix,192.168.1.50
```

### Testing

Mock DNS responses for testing:

```csv
payment-gateway.com,fqdn,127.0.0.1
external-api,prefix,127.0.0.1
```

### Traffic Routing

Route specific domains through proxies:

```csv
internal-app,prefix,10.0.0.100
vpn-only,prefix,172.16.0.1
```

### Ad Blocking

Redirect unwanted domains:

```csv
ads.example.com,fqdn,0.0.0.0
tracking,prefix,0.0.0.0
```

## Architecture

### Data Structures

- **Ternary Search Trees (TST)**: Used for efficient prefix and suffix matching
- **Hash Maps**: Used for exact FQDN lookups
- **Reversed Strings**: Suffix matching converts to prefix matching by reversing strings

### Lookup Algorithm

1. Check exact FQDN match in hash map (O(1))
2. Check prefix match in TST (O(log n))
3. Check suffix match in TST with reversed string (O(log n))
4. If no match, query upstream DNS

### Performance

- Prefix/Suffix lookups: O(log n) where n is the number of rules
- FQDN lookups: O(1)
- Memory efficient with TST data structure
- Concurrent request handling with Go routines

## Logging

FakeDNS uses structured logging with context:

```json
{
  "time": "2024-01-15T10:30:45Z",
  "level": "INFO",
  "msg": "returned sniproxy address for domain",
  "fqdn": "example.com.",
  "service": "dns"
}
```

## Security Considerations

- **No DNSSEC**: FakeDNS doesn't support DNSSEC validation
- **Trusted Networks**: Best used in trusted network environments
- **Rule Validation**: Validate rule files before deployment
- **Access Control**: Use firewall rules to restrict access

## Limitations

- No DNSSEC support
- No IPv6 (AAAA records) handling in fake responses
- No caching layer (relies on upstream DNS caching)
- Single-threaded rule loading

## Troubleshooting

### Port Already in Use

```bash
# Check what's using port 53
sudo ss -pltun -at '( dport = :53 or sport = :53 )'

# Use alternative ports
fakedns -udp 5353 -tcp 5353
```

### Permission Denied (Port < 1024)

```bash
# Option 1: Run with sudo
sudo fakedns -udp 53

# Option 2: Use high ports
fakedns -udp 5353

# Option 3: Grant capability (Linux)
sudo setcap CAP_NET_BIND_SERVICE=+eip ./fakedns
```

### Rules Not Loading

```bash
# Check file permissions
ls -la rules.csv

# Check file format (must be CSV)
cat rules.csv | head -5

# Enable debug logging
# (modify source to set slog.LevelDebug)
```

## Contributing

Contributions welcome! Please see the main repository [CONTRIBUTING.md](../CONTRIBUTING.md).

## License

See [LICENSE](LICENSE) file in the project root.

## Related Projects

- [CoreDNS](https://coredns.io/) - Production-grade DNS server
- [dnsmasq](https://thekelleys.org.uk/dnsmasq/doc.html) - Lightweight DNS forwarder
- [dnsclient](https://github.com/mosajjal/dnsclient) - Go DNS client library

---

**Note**: This tool is intended for development and testing purposes in controlled environments.