// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
)

type Identity struct {
	Tag     string
	ID      string
	Classes []string
}

func (i *Identity) SetIentifiers(attrs map[string]string) {
	i.Tag = attrs["tag"]
	i.ID = attrs["id"]
	if class, ok := attrs["class"]; ok {
		i.Classes = strings.Split(class, " ")
	}
}

func (i *Identity) String() string {
	return fmt.Sprintf("tag=%s id=%s class=%s", i.Tag, i.ID, strings.Join(i.Classes, " "))
}
