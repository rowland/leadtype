// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type StdSpan struct {
	AParent
	Identity
}

func (s *StdSpan) SetParent(value interface{}) error {
	switch value.(type) {
	case *StdSpan:
	case *StdParagraph:
	default:
		return fmt.Errorf("span must be child of p, pabel or another span.")
	}
	s.parent = value
	return nil
}

func init() {
	registerTag(DefaultSpace, "span", func() interface{} { return &StdSpan{} })
}
