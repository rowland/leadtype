// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package serveltml

import (
	"fmt"
	"os"
	"time"

	flag "github.com/namsral/flag"
	"github.com/rowland/leadtype/ttf_fonts"
)

// Config holds validated runtime configuration for the LTML render server.
type Config struct {
	Listen         string
	BasePath       string
	FontDirs       []string
	OutputPath     string
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
		fontDir        string
		outputPath     string
		maxUploadBytes int64
		readTimeout    time.Duration
		writeTimeout   time.Duration
	)

	flag.StringVar(&listen, "listen", ":8080", "address to listen on (LISTEN)")
	flag.StringVar(&basePath, "assets", "", "path to static asset directory (ASSETS, required)")
	flag.StringVar(&basePath, "a", "", "path to static asset directory (shorthand)")
	flag.StringVar(&fontDir, "font-dir", "auto", "ordered comma-delimited font directories; auto adds system directories (FONT_DIR)")
	flag.StringVar(&outputPath, "output-path", "", "root directory for file output (OUTPUT_PATH, optional; enables X-Output-File)")
	flag.Int64Var(&maxUploadBytes, "max-upload-bytes", 32<<20, "maximum multipart request size in bytes (MAX_UPLOAD_BYTES)")
	flag.DurationVar(&readTimeout, "read-timeout", 0, "HTTP server read timeout, e.g. 30s (READ_TIMEOUT)")
	flag.DurationVar(&writeTimeout, "write-timeout", 0, "HTTP server write timeout, e.g. 60s (WRITE_TIMEOUT)")

	flag.Parse()

	if basePath == "" {
		return nil, fmt.Errorf("assets (or ASSETS) is required")
	}
	info, err := os.Stat(basePath)
	if err != nil {
		return nil, fmt.Errorf("assets %q: %w", basePath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("assets %q is not a directory", basePath)
	}

	fontDirs, err := ttf_fonts.ResolveFontDirs(fontDir)
	if err != nil {
		return nil, fmt.Errorf("font-dir: %w", err)
	}

	if outputPath != "" {
		info, err := os.Stat(outputPath)
		if err != nil {
			return nil, fmt.Errorf("output-path %q: %w", outputPath, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("output-path %q is not a directory", outputPath)
		}
	}

	return &Config{
		Listen:         listen,
		BasePath:       basePath,
		FontDirs:       fontDirs,
		OutputPath:     outputPath,
		MaxUploadBytes: maxUploadBytes,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
	}, nil
}
