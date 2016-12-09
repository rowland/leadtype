// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type Rectangle struct {
	StdContainer
}

func init() {
	registerTag(DefaultSpace, "rect", func() interface{} { return &Rectangle{} })
}
