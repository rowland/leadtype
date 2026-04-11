// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type RawBody interface {
	SetBody(string)
}

type Component interface {
	Container
	RawBody
}

type WantsDoc interface {
	SetDoc(*Doc)
}
