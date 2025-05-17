package webhookio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BodyEncodingType defines how the log data is encoded in the HTTP request.
type BodyEncodingType string

const (
	// RawBody sends the log data as the raw request body.
	// Configure 'ContentType' header for this.
	RawBody BodyEncodingType = "raw"
	// JSONStringEmbed wraps the log string in a JSON object.
	// e.g., {"message": "log line content"}
	// Content-Type will be application/json.
	JSONStringEmbed BodyEncodingType = "json_string_embed"
	// QueryParameter sends log data as a URL query parameter (for GET requests mostly).
	QueryParameter BodyEncodingType = "query_param"
)

// WebhookWriterConfig holds the configuration for the WebhookWriter.
type WebhookWriterConfig struct {
	URL           string            // The webhook URL.
	Method        string            // HTTP method (GET, POST, etc.). Defaults to POST.
	Headers       map[string]string // Additional static headers to send.
	Encoding      BodyEncodingType  // How to encode the log data in the request.
	QueryParamKey string            // Key for the query parameter if Encoding is QueryParameter. Defaults to "log".
	JSONEmbedKey  string            // Key for the JSON object if Encoding is JSONStringEmbed. Defaults to "message".
	Timeout       time.Duration     // Timeout for the HTTP request. Defaults to 10 seconds.
	ContentType   string            // Content-Type header for RawBody encoding. Defaults to "text/plain".
}

// WebhookWriter implements io.Writer to push log lines to a webhook.
type WebhookWriter struct {
	config     WebhookWriterConfig
	httpClient *http.Client
}

// NewWebhookWriter creates and initializes a new WebhookWriter.
func NewWebhookWriter(config WebhookWriterConfig) (*WebhookWriter, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("webhook URL cannot be empty")
	}
	if config.Method == "" {
		config.Method = http.MethodPost
	}
	config.Method = strings.ToUpper(config.Method)

	if config.Encoding == "" {
		config.Encoding = RawBody // Default encoding
	}

	if config.Encoding == QueryParameter && config.QueryParamKey == "" {
		config.QueryParamKey = "log"
	}
	if config.Encoding == JSONStringEmbed && config.JSONEmbedKey == "" {
		config.JSONEmbedKey = "message"
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.Encoding == RawBody && config.ContentType == "" {
		config.ContentType = "text/plain"
	}

	return &WebhookWriter{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// Write sends the log data p to the configured webhook.
func (w *WebhookWriter) Write(p []byte) (n int, err error) {
	var req *http.Request
	var reqBody io.Reader
	var finalURL string = w.config.URL

	logString := string(p) // Original log data as string

	switch w.config.Encoding {
	case RawBody:
		reqBody = bytes.NewBuffer(p)
	case JSONStringEmbed:
		jsonData, err := json.Marshal(map[string]string{
			w.config.JSONEmbedKey: logString,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to marshal log line to JSON: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	case QueryParameter:
		if w.config.Method != http.MethodGet {
			// While GET is typical, other methods can have query params.
			// For simplicity in this example, we'll focus on its common use with GET.
			// If not GET, it's often better to send data in the body.
		}
		parsedURL, err := url.Parse(w.config.URL)
		if err != nil {
			return 0, fmt.Errorf("failed to parse webhook URL: %w", err)
		}
		query := parsedURL.Query()
		query.Set(w.config.QueryParamKey, logString)
		parsedURL.RawQuery = query.Encode()
		finalURL = parsedURL.String()
		reqBody = nil // No body for GET requests typically
	default:
		return 0, fmt.Errorf("unsupported webhook encoding type: %s", w.config.Encoding)
	}

	req, err = http.NewRequest(w.config.Method, finalURL, reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Set static headers
	for key, value := range w.config.Headers {
		req.Header.Set(key, value)
	}

	// Set Content-Type based on encoding
	switch w.config.Encoding {
	case RawBody:
		req.Header.Set("Content-Type", w.config.ContentType)
	case JSONStringEmbed:
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		// You might want to log this error to a fallback mechanism if critical
		// For now, we just return it, and charmbracelet/log might print it to its other writers.
		return 0, fmt.Errorf("failed to send log to webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Log error or handle non-successful status codes
		bodyBytes, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("webhook returned non-success status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return len(p), nil
}
