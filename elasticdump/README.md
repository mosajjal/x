# ElasticDump

[![Go Reference](https://pkg.go.dev/badge/github.com/mosajjal/go-exp/elasticdump.svg)](https://pkg.go.dev/github.com/mosajjal/go-exp/elasticdump)

A high-performance tool to dump entire Elasticsearch clusters with configurable filtering options. Supports concurrent downloads with structured logging.

## Features

- 🚀 **Concurrent Downloads**: Parallel index dumping for maximum throughput
- 🔍 **Smart Filtering**: Filter by document count, index size, and regex patterns
- 📊 **Structured Logging**: JSON-formatted logs with detailed context
- 🎯 **Precision Control**: Fine-grained control over what gets dumped
- 💾 **Efficient**: Uses Elasticsearch scroll API for memory-efficient large dataset handling

## Installation

```bash
go install github.com/mosajjal/go-exp/elasticdump@latest
```

Or build from source:

```bash
git clone https://github.com/mosajjal/go-exp.git
cd go-exp/elasticdump
go build
```

## Usage

### Basic Usage

Dump all indices from an Elasticsearch cluster:

```bash
elasticdump -targetIP 192.168.1.100
```

### Advanced Usage

Filter indices by size, document count, and name pattern:

```bash
elasticdump \
  -targetIP 192.168.1.100 \
  -targetPort 9200 \
  -minDocCount 1000 \
  -minIndexSizeKB 10240 \
  -indexRegex "^logs-.*-2024"
```

### Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `-targetIP` | (required) | Target Elasticsearch IP address or hostname |
| `-targetPort` | `9200` | Target Elasticsearch port |
| `-minDocCount` | `100` | Minimum number of documents for an index to be dumped |
| `-minIndexSizeKB` | `1024` | Minimum index size in KB for dumping |
| `-indexRegex` | `.*` | Only dump indices matching this regex pattern |

## Output Format

- Creates a directory named after the target IP
- Each index is saved as `ESDUMP-{IP}-{INDEX}-{TIMESTAMP}.json`
- JSON format contains raw Elasticsearch documents
- Logs are output in structured JSON format to stdout

### Example Output Structure

```
./192.168.1.100/
  ├── ESDUMP-192.168.1.100-logs-2024-01-01-2024-01-15T10:30:45Z.json
  ├── ESDUMP-192.168.1.100-logs-2024-01-02-2024-01-15T10:31:12Z.json
  └── ESDUMP-192.168.1.100-metrics-2024-01-01-2024-01-15T10:31:45Z.json
```

## Examples

### Dump Production Logs

```bash
# Dump all production log indices larger than 10GB
elasticdump \
  -targetIP prod-es.example.com \
  -minIndexSizeKB 10485760 \
  -indexRegex "^prod-logs-.*"
```

### Backup Recent Indices

```bash
# Dump indices from the last month
elasticdump \
  -targetIP 10.0.0.50 \
  -indexRegex ".*-2024-01-.*"
```

### Small Index Export

```bash
# Export small test indices
elasticdump \
  -targetIP localhost \
  -minDocCount 10 \
  -minIndexSizeKB 100 \
  -indexRegex "^test-.*"
```

## How It Works

1. **Discovery**: Queries `_cat/indices` API to get list of all indices
2. **Filtering**: Applies document count, size, and regex filters
3. **Scrolling**: Uses Elasticsearch scroll API for efficient pagination
4. **Parallel Dumping**: Dumps multiple indices concurrently
5. **Output**: Writes JSON documents to files in organized directory structure

## Performance Considerations

- Uses Elasticsearch scroll API with 10-minute scroll timeout
- Fetches 1000 documents per scroll request
- Concurrent index dumping (one goroutine per index)
- HTTP client timeout: 30 seconds per request

## Error Handling

- Graceful error handling with detailed logging
- Failed indices don't block other downloads
- Exit code 1 if any dumps fail
- Panic recovery with error reporting

## Logging

Uses structured JSON logging with contextual information:

```json
{
  "time": "2024-01-15T10:30:45Z",
  "level": "INFO",
  "msg": "Successfully dumped index",
  "index": "logs-2024-01-01",
  "target": "192.168.1.100",
  "file": "./192.168.1.100/ESDUMP-192.168.1.100-logs-2024-01-01-2024-01-15T10:30:45Z.json"
}
```

## Requirements

- Go 1.25 or later
- Network access to Elasticsearch cluster
- Sufficient disk space for dumps

## Roadmap

- [ ] TLS/HTTPS support with certificate validation
- [ ] Authentication (Basic Auth, API Keys)
- [ ] Compression support (gzip, zstd)
- [ ] Cloud storage upload (S3, GCS, Azure Blob)
- [ ] Restore functionality
- [ ] Progress bar for large dumps
- [ ] Resume capability for interrupted dumps
- [ ] Custom scroll size and timeout configuration
- [ ] Elasticsearch 8.x support

## Contributing

Contributions welcome! Please see the main repository [CONTRIBUTING.md](../CONTRIBUTING.md).

## License

See [LICENSE](LICENSE) file in the project root.

## Related Projects

- [Elasticsearch](https://www.elastic.co/elasticsearch/)
- [elasticdump (Node.js)](https://github.com/elasticsearch-dump/elasticsearch-dump)

---

**Note**: This tool is for data export and backup purposes. Always ensure you have proper authorization before dumping production data.
