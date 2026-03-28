// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

// serve-ltml is an HTTP server that renders LTML documents to PDF.
//
// It accepts POST /render requests with multipart/form-data bodies:
//   - The first part must have field name "ltml" and contain the LTML document.
//     The preferred Content-Type is application/vnd.rowland.leadtype.ltml+xml;
//     application/xml, text/xml, and an empty Content-Type are also accepted.
//   - Subsequent parts may have field name "file" and contain asset files whose
//     multipart filename is used as the virtual path during rendering. These
//     assets shadow same-named files in the configured base-path for the
//     duration of the request only.
//
// Configuration is accepted as flags or environment variables:
//
//	LISTEN / -listen            address to listen on (default :8080)
//	BASE_PATH / -base-path      path to static asset directory (required)
//	MAX_UPLOAD_BYTES / -max-upload-bytes  request size cap (default 32 MiB)
//	READ_TIMEOUT / -read-timeout         HTTP read timeout (default none)
//	WRITE_TIMEOUT / -write-timeout       HTTP write timeout (default none)
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve-ltml: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	// renderHandler returns 405 for non-POST requests, so a single pattern suffices.
	mux.Handle("/render", newRenderHandler(cfg))

	srv := &http.Server{
		Addr:         cfg.Listen,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	log.Printf("serve-ltml: listening on %s (base-path=%s)", cfg.Listen, cfg.BasePath)
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "serve-ltml: %v\n", err)
		os.Exit(1)
	}
}
