// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package serveltml

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func Main(stderr io.Writer, registerWidgets func() error) int {
	if stderr == nil {
		stderr = os.Stderr
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
