// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"encoding/xml"
	"strings"
)

func mapFromXmlAttrs(attrs []xml.Attr) map[string]string {
	result := make(map[string]string)
	for _, attr := range attrs {
		result[attr.Name.Local] = attr.Value
	}
	return result
}

func MapHasKeyPrefix(attrs map[string]string, prefix string) bool {
	for k, _ := range attrs {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

// type stringSlice []string

// func (ss stringSlice) index(value string) int {
// 	for i, s := range ss {
// 		if s == value {
// 			return i
// 		}
// 	}
// 	return -1
// }

// func (ss stringSlice) has(value string) bool {
// 	return ss.index(value) >= 0
// }
