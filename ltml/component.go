// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type Component interface {
	Container
	SetBody(string)
}

type WantsDoc interface {
	SetDoc(*Doc)
}
