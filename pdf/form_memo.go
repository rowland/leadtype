// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rowland/leadtype/options"
)

type cachedForm struct {
	form   *pdfForm
	name   string
	width  float64
	height float64
}

func memoFormOptionsKey(opts options.Options, units string) string {
	if len(opts) == 0 && units == "" {
		return ""
	}
	keys := make([]string, 0, len(opts)+1)
	for key := range opts {
		switch key {
		case "page_width", "page_height", "crop_width", "crop_height", "rotate":
			continue
		case "units":
			continue
		default:
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	var b strings.Builder
	if units != "" {
		fmt.Fprintf(&b, "units=%s;", units)
	}
	for _, key := range keys {
		fmt.Fprintf(&b, "%s=%s;", key, memoFormOptionValue(opts[key]))
	}
	return b.String()
}

func memoFormOptionValue(value any) string {
	switch value := value.(type) {
	case nil:
		return "<nil>"
	case string:
		return value
	case bool:
		if value {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", value)
	case float64:
		return g(value)
	default:
		return fmt.Sprintf("%v", value)
	}
}
