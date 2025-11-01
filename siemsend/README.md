# SiemSend

[![Go Reference](https://pkg.go.dev/badge/github.com/mosajjal/go-exp/siemsend.svg)](https://pkg.go.dev/github.com/mosajjal/go-exp/siemsend)

A UNIX philosophy-inspired SIEM connector that reads JSON Lines from stdin and sends them to various SIEM backends. Designed for pipeline-based log processing and aggregation.

## Features

- 📊 **Multiple SIEM Backends**: Azure Sentinel, Elastic, OpenSearch, Splunk
- 🔄 **Streaming Architecture**: Reads from stdin, follows UNIX pipeline philosophy
- 📦 **Batch Processing**: Configurable batching for optimal performance
- 🗜️ **Compression**: Optional gzip compression for network efficiency
- 🔐 **Secure**: TLS support for encrypted transmission
- 📝 **Structured Logging**: JSON-formatted logs with detailed context

## Installation

```bash
go install github.com/mosajjal/go-exp/siemsend@latest
```

## Usage

### Azure Sentinel

```bash
cat logs.jsonl | siemsend sentinel \
  --workspaceId YOUR_WORKSPACE_ID \
  --sharedKey YOUR_SHARED_KEY \
  --logType CustomLogType
```

### Elasticsearch

```bash
cat logs.jsonl | siemsend elastic \
  --endpoint https://es.example.com:9200 \
  --index logs-2024-01 \
  --compress
```

### OpenSearch

```bash
cat logs.jsonl | siemsend opensearch \
  --endpoint https://opensearch.example.com:9200 \
  --index application-logs
```

### Splunk

```bash
cat logs.jsonl | siemsend splunk \
  --endpoint https://splunk.example.com:8088 \
  --token YOUR_HEC_TOKEN
```

## Input Format

SiemSend expects JSON Lines (JSONL) format - one JSON object per line:

```json
{"timestamp":"2024-01-15T10:30:00Z","level":"error","message":"Connection timeout"}
{"timestamp":"2024-01-15T10:30:01Z","level":"info","message":"Request processed"}
```

## License

See [LICENSE](LICENSE) file in the project root.