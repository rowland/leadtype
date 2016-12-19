// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type HasAttrs interface {
	SetAttrs(attrs map[string]string)
}

type HasAttrsPrefix interface {
	SetAttrs(prefix string, attrs map[string]string)
}
