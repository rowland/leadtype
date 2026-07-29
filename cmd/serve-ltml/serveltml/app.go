// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

// Package serveltml provides a reusable HTTP server that renders LTML
// documents to PDF.
//
// It accepts POST /render requests with multipart/form-data bodies:
//   - The first part must have field name "ltml" and contain the LTML
//     document. The preferred Content-Type is
//     application/vnd.rowland.leadtype.ltml+xml; application/xml, text/xml,
//     and an empty Content-Type are also accepted.
//   - Subsequent parts may have field name "file" and contain asset files
//     whose multipart filename is used as the virtual path during rendering.
//     These assets shadow same-named files in the configured assets directory
//     for the duration of the request only.
//
// Configuration is accepted as flags or environment variables:
//
//	LISTEN / -listen                      address to listen on (default :8080)
//	ASSETS / -assets / -a                path to static asset directory (required)
//	FONT_DIR / -font-dir                 ordered font directories (default auto)
//	MAX_UPLOAD_BYTES / -max-upload-bytes request size cap (default 32 MiB)
//	READ_TIMEOUT / -read-timeout         HTTP read timeout (default none)
//	WRITE_TIMEOUT / -write-timeout       HTTP write timeout (default none)
package serveltml

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	leadtype "github.com/rowland/leadtype"
)

// Main runs the serve-ltml command flow and returns a process exit code.
// If registerWidgets is non-nil, it is called before request handling starts.
func Main(stderr io.Writer, registerWidgets func() error) int {
	if stderr == nil {
		stderr = os.Stderr
	}
	if len(os.Args) == 2 && (os.Args[1] == "-version" || os.Args[1] == "--version") {
		fmt.Fprintf(stderr, "serve-ltml %s\n", leadtype.Version)
		return 0
	}
	if registerWidgets != nil {
		if err := registerWidgets(); err != nil {
			fmt.Fprintf(stderr, "serve-ltml: %v\n", err)
			return 1
		}
	}

	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintf(stderr, "serve-ltml: %v\n", err)
		return 1
	}

	mux := http.NewServeMux()
	mux.Handle("/render", newRenderHandler(cfg))

	srv := &http.Server{
		Addr:         cfg.Listen,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	log.Printf("serve-ltml: listening on %s (assets=%s)", cfg.Listen, cfg.BasePath)
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(stderr, "serve-ltml: %v\n", err)
		return 1
	}
	return 0
}
