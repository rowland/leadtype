// Copyright 2016, 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"sort"
	"strings"
)

type Identity struct {
	Tag     string
	ID      string
	Classes []string
}

func (i *Identity) SelectorTag() string {
	result := i.Tag
	if i.ID != "" {
		result += "#" + i.ID
	}
	if len(i.Classes) > 0 {
		result += "." + strings.Join(i.Classes, ".")
	}
	return result
}

func (i *Identity) SetIentifiers(attrs map[string]string) {
	i.Tag = attrs["tag"]
	i.ID = attrs["id"]
	if class, ok := attrs["class"]; ok {
		i.Classes = strings.Split(class, " ")
		sort.Strings(i.Classes)
	}
}

func (i *Identity) String() string {
	return fmt.Sprintf("tag=%s id=%s class=%s", i.Tag, i.ID, strings.Join(i.Classes, " "))
}
