// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"errors"
	"regexp"
)

const DefaultSpace = "std"

type TagFactory func() any

var (
	reTag           = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	registeredTags  = make(map[string]TagFactory)
	componentTags   = make(map[string]bool)
	errStdReserved  = errors.New("namespace '" + DefaultSpace + "' is reserved")
	errBadNamespace = errors.New("namespace restricted to letters, numbers, underscores, and hyphens")
	errBadTag       = errors.New("tag restricted to letters, numbers, underscores, and hyphens")
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

func registerTagExt(namespace, tag string, f TagFactory) error {
	if err := registerTag(namespace, tag, f); err != nil {
		return err
	}
	componentTags[namespace+":"+tag] = true
	return nil
}

func RegisterTagExt(namespace, tag string, f TagFactory) error {
	if namespace == DefaultSpace {
		return errStdReserved
	}
	return registerTagExt(namespace, tag, f)
}

func makeElement(namespace, tag string) any {
	if f, ok := registeredTags[namespace+":"+tag]; ok {
		return f()
	}
	return nil
}

func isComponentTag(namespace, tag string) bool {
	return componentTags[namespace+":"+tag]
}
