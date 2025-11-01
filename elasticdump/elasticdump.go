// Package main provides elasticdump, a tool to download entire Elasticsearch clusters
// with configurable filtering options for indices based on document count, size, and regex patterns.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
)

// Config holds the configuration for the elasticdump operation.
type Config struct {
	TargetIP       string
	TargetPort     uint
	MinDocCount    uint
	MinIndexSizeKB uint
	IndexRegex     *regexp.Regexp
}

var (
	logger *slog.Logger
)

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// parseFlags parses and validates command-line flags.
func parseFlags() (*Config, error) {
	targetIP := flag.String("targetIP", "", "Target Elasticsearch IP Address (required)")
	targetPort := flag.Uint("targetPort", 9200, "Target Elasticsearch port")
	minDocCount := flag.Uint("minDocCount", 100, "Minimum number of documents for each index")
	minIndexSizeKB := flag.Uint("minIndexSizeKB", 1024, "Minimum size of index for dump (in KB)")
	indexRegex := flag.String("indexRegex", ".*", "Only download indices matching this regex pattern")

	flag.Parse()

	if *targetPort > 65535 {
		return nil, fmt.Errorf("targetPort must be between 1 and 65535, got %d", *targetPort)
	}

	if *targetIP == "" {
		return nil, fmt.Errorf("targetIP is required")
	}

	re, err := regexp.Compile(*indexRegex)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	return &Config{
		TargetIP:       *targetIP,
		TargetPort:     *targetPort,
		MinDocCount:    *minDocCount,
		MinIndexSizeKB: *minIndexSizeKB,
		IndexRegex:     re,
	}, nil
}

// getNextScroll fetches the next batch of documents using scroll API.
func getNextScroll(ctx context.Context, client *http.Client, ip string, port uint, scrollID string, f *os.File) error {
	postData := []byte(fmt.Sprintf(`{"scroll_id": "%s", "scroll": "10m"}`, scrollID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s:%d/_search/scroll", ip, port),
		bytes.NewBuffer(postData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute scroll request: %w", err)
	}
	defer resp.Body.Close()

	resBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	hits := gjson.GetBytes(resBytes, "hits.hits")
	if hits.Raw == "" || hits.Raw == "[]" {
		return nil
	}

	if _, err := f.Write([]byte(hits.Raw)); err != nil {
		return fmt.Errorf("failed to write hits to file: %w", err)
	}

	nextScroll := gjson.GetBytes(resBytes, "_scroll_id")
	if nextScroll.Exists() {
		return getNextScroll(ctx, client, ip, port, nextScroll.String(), f)
	}

	return nil
}

// dumpIndexToJSON exports a single Elasticsearch index to a JSON file.
func dumpIndexToJSON(ctx context.Context, client *http.Client, cfg *Config, index string, done chan<- error) {
	defer func() {
		if r := recover(); r != nil {
			done <- fmt.Errorf("panic in dumpIndexToJSON: %v", r)
		}
	}()

	log := logger.With(
		slog.String("index", index),
		slog.String("target", cfg.TargetIP),
	)

	postData := []byte(`{"size": 1000}`)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s:%d/%s/_search?scroll=10m", cfg.TargetIP, cfg.TargetPort, index),
		bytes.NewBuffer(postData))
	if err != nil {
		done <- fmt.Errorf("failed to create request for index %s: %w", index, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		done <- fmt.Errorf("failed to query index %s: %w", index, err)
		return
	}
	defer resp.Body.Close()

	// Create output directory
	outputDir := fmt.Sprintf("./%s/", cfg.TargetIP)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		done <- fmt.Errorf("failed to create output directory: %w", err)
		return
	}

	filename := fmt.Sprintf("%sESDUMP-%s-%s-%s.json", outputDir, cfg.TargetIP, index, time.Now().Format(time.RFC3339))
	f, err := os.Create(filename)
	if err != nil {
		done <- fmt.Errorf("failed to create output file: %w", err)
		return
	}
	defer f.Close()

	resBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		done <- fmt.Errorf("failed to read response body: %w", err)
		return
	}

	hits := gjson.GetBytes(resBytes, "hits.hits")
	if hits.Raw != "" && hits.Raw != "[]" {
		if _, err := f.Write([]byte(hits.Raw)); err != nil {
			done <- fmt.Errorf("failed to write initial hits: %w", err)
			return
		}
	}

	nextScroll := gjson.GetBytes(resBytes, "_scroll_id")
	if nextScroll.Exists() {
		if err := getNextScroll(ctx, client, cfg.TargetIP, cfg.TargetPort, nextScroll.String(), f); err != nil {
			done <- fmt.Errorf("failed during scroll: %w", err)
			return
		}
	}

	log.Info("Successfully dumped index", slog.String("file", filename))
	done <- nil
}

// getIndexList retrieves the list of indices that match the configured criteria.
func getIndexList(ctx context.Context, client *http.Client, cfg *Config) ([]string, error) {
	var result []string

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("http://%s:%d/_cat/indices?format=json&bytes=kb", cfg.TargetIP, cfg.TargetPort),
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get index list: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	indices := gjson.ParseBytes(body)
	indices.ForEach(func(key, value gjson.Result) bool {
		docCount, _ := strconv.Atoi(gjson.Get(value.String(), "docs\\.count").String())
		indexSizeKB, _ := strconv.Atoi(gjson.Get(value.String(), "store\\.size").String())
		indexName := gjson.Get(value.String(), "index").String()

		log := logger.With(
			slog.String("index", indexName),
			slog.Int("docs", docCount),
			slog.Int("size_kb", indexSizeKB),
		)

		if uint(docCount) >= cfg.MinDocCount && uint(indexSizeKB) >= cfg.MinIndexSizeKB {
			if cfg.IndexRegex.MatchString(indexName) {
				result = append(result, indexName)
				log.Info("Index matched criteria")
			} else {
				log.Info("Index did not match regex pattern, skipping")
			}
		} else {
			log.Info("Index did not meet size/document requirements, skipping")
		}
		return true
	})

	return result, nil
}

func main() {
	logger.Info("ElasticDump starting...")

	cfg, err := parseFlags()
	if err != nil {
		logger.Error("Configuration error", slog.String("error", err.Error()))
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	logger.Info("Fetching index list",
		slog.String("target", cfg.TargetIP),
		slog.Uint64("port", uint64(cfg.TargetPort)))

	indexList, err := getIndexList(ctx, client, cfg)
	if err != nil {
		logger.Error("Failed to get index list", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if len(indexList) == 0 {
		logger.Warn("No indices matched the specified criteria")
		os.Exit(0)
	}

	logger.Info("Starting index dump",
		slog.Int("count", len(indexList)),
		slog.Any("indices", indexList))

	done := make(chan error, len(indexList))
	for _, index := range indexList {
		go dumpIndexToJSON(ctx, client, cfg, index, done)
	}

	// Wait for all dumps to complete
	errorCount := 0
	for i := 0; i < len(indexList); i++ {
		if err := <-done; err != nil {
			logger.Error("Index dump failed",
				slog.String("index", indexList[i]),
				slog.String("error", err.Error()))
			errorCount++
		}
	}

	if errorCount > 0 {
		logger.Error("Some dumps failed",
			slog.Int("failed", errorCount),
			slog.Int("total", len(indexList)))
		os.Exit(1)
	}

	logger.Info("ElasticDump completed successfully",
		slog.Int("indices_dumped", len(indexList)))
}
