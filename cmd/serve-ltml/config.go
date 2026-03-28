// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"time"

	flag "github.com/namsral/flag"
)

// Config holds validated runtime configuration for the LTML render server.
type Config struct {
	Listen         string
	BasePath       string
	MaxUploadBytes int64
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

// parseConfig reads flags and environment variables via namsral/flag and
// returns a validated Config. It calls flag.Parse internally. If
// configuration is invalid, parseConfig returns a non-nil error describing
// the problem; the caller should print it and exit.
func parseConfig() (*Config, error) {
	var (
		listen         string
		basePath       string
		maxUploadBytes int64
		readTimeout    time.Duration
		writeTimeout   time.Duration
	)

	flag.StringVar(&listen, "listen", ":8080", "address to listen on (LISTEN)")
	flag.StringVar(&basePath, "base-path", "", "path to static asset directory (BASE_PATH, required)")
	flag.Int64Var(&maxUploadBytes, "max-upload-bytes", 32<<20, "maximum multipart request size in bytes (MAX_UPLOAD_BYTES)")
	flag.DurationVar(&readTimeout, "read-timeout", 0, "HTTP server read timeout, e.g. 30s (READ_TIMEOUT)")
	flag.DurationVar(&writeTimeout, "write-timeout", 0, "HTTP server write timeout, e.g. 60s (WRITE_TIMEOUT)")

	flag.Parse()

	if basePath == "" {
		return nil, fmt.Errorf("base-path (or BASE_PATH) is required")
	}
	info, err := os.Stat(basePath)
	if err != nil {
		return nil, fmt.Errorf("base-path %q: %w", basePath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("base-path %q is not a directory", basePath)
	}

	return &Config{
		Listen:         listen,
		BasePath:       basePath,
		MaxUploadBytes: maxUploadBytes,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
	}, nil
}
