// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"errors"
	"regexp"
)

const DefaultSpace = "std"

type TagFactory func() interface{}

var (
	reTag           = regexp.MustCompile(`^\w+$`)
	registeredTags  = make(map[string]TagFactory)
	errStdReserved  = errors.New("namespace '" + DefaultSpace + "' is reserved")
	errBadNamespace = errors.New("namespace restricted to letters, numbers and the underscore")
	errBadTag       = errors.New("namespace restricted to letters, numbers and the underscore")
)

func registerTag(namespace, tag string, f TagFactory) error {
	if !reTag.MatchString(namespace) {
		return errBadNamespace
	}
	if !reTag.MatchString(tag) {
		return errBadTag
	}
	registeredTags[namespace+":"+tag] = f
	return nil
}

func RegisterTag(namespace, tag string, f TagFactory) error {
	if namespace == DefaultSpace {
		return errStdReserved
	}
	return registerTag(namespace, tag, f)
}

func makeElement(namespace, tag string) interface{} {
	if f, ok := registeredTags[namespace+":"+tag]; ok {
		return f()
	}
	return nil
}
