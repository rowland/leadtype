package ltml

import (
	"fmt"
	"os"
)

var debugEnabled = os.Getenv("LEADTYPE_DEBUG") != ""

func debugf(format string, args ...interface{}) {
	if !debugEnabled {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}
