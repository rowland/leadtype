package ltml

import (
	"fmt"
	"os"
)

var debugEnabled = os.Getenv("LEADTYPE_DEBUG") != ""

func debugf(format string, args ...any) {
	if !debugEnabled {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}
