# webhookio

`webhookio` is a lightweight Go module that implements an `io.Writer` to send log lines or arbitrary data to a configurable HTTP webhook endpoint.
This is useful for integrating Go applications with external logging, alerting, or automation systems that accept webhooks.

## Features

- **Flexible HTTP Method**: Supports POST, GET, and other HTTP methods.
- **Multiple Encoding Modes**: Send data as raw body, JSON-embedded, or as a query parameter.
- **Custom Headers**: Add static headers to each request.
- **Timeout Control**: Configure HTTP request timeout.
- **Content-Type Control**: Set content type for raw body encoding.
- **Easy Integration**: Implements `io.Writer` for seamless use with logging libraries.

## Installation

```sh
go get github.com/mosajjal/x/webhookio
```

## Usage

```go
import (
    "log"
    "github.com/mosajjal/x/webhookio"
)

func main() {
    writer, err := webhookio.NewWebhookWriter(webhookio.WebhookWriterConfig{
        URL:     "https://example.com/webhook",
        Method:  "POST",
        Headers: map[string]string{"Authorization": "Bearer TOKEN"},
        Encoding: webhookio.JSONStringEmbed,
        JSONEmbedKey: "log_message",
        Timeout: 5 * time.Second,
    })
    if err != nil {
        log.Fatalf("Failed to create webhook writer: %v", err)
    }

    logger := log.New(writer, "", log.LstdFlags)
    logger.Println("Hello, webhook world!")
}
```

## Configuration

The `WebhookWriterConfig` struct supports the following fields:

| Field           | Type              | Description                                                                 | Default           |
|-----------------|-------------------|-----------------------------------------------------------------------------|-------------------|
| `URL`           | `string`          | The webhook endpoint URL.                                                   | **(required)**    |
| `Method`        | `string`          | HTTP method to use (e.g., POST, GET).                                       | `POST`            |
| `Headers`       | `map[string]string` | Additional headers to include in the request.                              | `nil`             |
| `Encoding`      | `BodyEncodingType` | How to encode the log data (`raw`, `json_string_embed`, `query_param`).    | `raw`             |
| `QueryParamKey` | `string`          | Query parameter key (for `query_param` encoding).                           | `"log"`           |
| `JSONEmbedKey`  | `string`          | JSON key (for `json_string_embed` encoding).                                | `"message"`       |
| `Timeout`       | `time.Duration`   | HTTP request timeout.                                                       | `10s`             |
| `ContentType`   | `string`          | Content-Type header (for `raw` encoding).                                   | `"text/plain"`    |

### Encoding Modes

- **RawBody**: Sends the log data as the raw HTTP request body.
- **JSONStringEmbed**: Wraps the log string in a JSON object (e.g., `{"message": "log content"}`).
- **QueryParameter**: Sends the log data as a URL query parameter (useful for GET requests).

## Error Handling

If the webhook returns a non-2xx status code or the request fails, the error is returned from the `Write` method. You may want to use a fallback logging mechanism in production.

## License

MIT License

---

*This module is ideal for integrating Go logs with external systems via webhooks in a simple, idiomatic way.*
