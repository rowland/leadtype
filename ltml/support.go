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

func MapHasAnyKey(attrs map[string]string, keys ...string) bool {
	for _, key := range keys {
		if _, ok := attrs[key]; ok {
			return true
		}
	}
	return false
}

func filterMapAttrs(prefix string, attrs map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range attrs {
		if strings.HasPrefix(k, prefix) {
			result[strings.TrimPrefix(k, prefix)] = v
		}
	}
	return result
}

func addUnits(attrs map[string]string, units Units) map[string]string {
	if _, ok := attrs["units"]; ok {
		return attrs
	}
	attrs["units"] = string(units)
	return attrs
}
