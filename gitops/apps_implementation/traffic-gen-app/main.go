package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	// Setup structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting traffic generator")

	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://backend-prod:80"
		slog.Warn("BACKEND_URL not set, using default", "backend_url", backendURL)
	}

	intervalStr := os.Getenv("INTERVAL")
	if intervalStr == "" {
		intervalStr = "5s"
		slog.Info("INTERVAL not set, using default", "interval", intervalStr)
	}

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		slog.Error("Invalid INTERVAL format", "interval", intervalStr, "error", err)
		os.Exit(1)
	}

	endpoints := []string{
		"/",
		"/health",
		"/config",
		"/pods",
		"/metrics",
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	slog.Info("Traffic generator configured",
		"backend_url", backendURL,
		"interval", interval,
		"endpoints", endpoints,
		"timeout", client.Timeout,
	)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Starting traffic generation loop")
	// Initial request
	makeRequests(client, backendURL, endpoints)

	for range ticker.C {
		makeRequests(client, backendURL, endpoints)
	}
}

func makeRequests(client *http.Client, baseURL string, endpoints []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Debug("Starting request cycle", "endpoint_count", len(endpoints))

	for _, endpoint := range endpoints {
		url := baseURL + endpoint

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			slog.Error("Failed to create request", "url", url, "error", err)
			continue
		}

		start := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(start)

		if err != nil {
			slog.Warn("Request failed",
				"url", url,
				"error", err,
				"duration_ms", duration.Milliseconds(),
			)
			continue
		}
		defer resp.Body.Close()

		slog.Info("Request completed",
			"method", req.Method,
			"url", url,
			"status_code", resp.StatusCode,
			"duration_ms", duration.Milliseconds(),
		)

		if resp.StatusCode >= 400 {
			slog.Warn("Request returned error status",
				"url", url,
				"status_code", resp.StatusCode,
				"duration_ms", duration.Milliseconds(),
			)
		}
	}

	slog.Debug("Request cycle completed", "endpoint_count", len(endpoints))
}
