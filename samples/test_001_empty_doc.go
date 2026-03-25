// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import "github.com/rowland/leadtype/pdf"

func init() {
	registerSample("test_001_empty_doc", "create an empty PDF document", runTest001EmptyDoc)
}

func runTest001EmptyDoc() (string, error) {
	return writeDoc("test_001_empty_doc.pdf", func(doc *pdf.DocWriter) error {
		return nil
	})
}
