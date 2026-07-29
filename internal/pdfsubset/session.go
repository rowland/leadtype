// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

// Package pdfsubset provides short-lived resource sharing for PDF font
// subsetting without coupling font backends to the public font package.
package pdfsubset

type Session struct {
	resources map[any]resource
}

type resource struct {
	value any
	err   error
}

func NewSession() *Session {
	return &Session{resources: make(map[any]resource)}
}

// Load returns the resource previously stored for key, or calls loader once
// and caches both its value and error. Keys must be comparable.
func (session *Session) Load(key any, loader func() (any, error)) (any, error) {
	if session == nil {
		return loader()
	}
	if cached, ok := session.resources[key]; ok {
		return cached.value, cached.err
	}
	value, err := loader()
	session.resources[key] = resource{value: value, err: err}
	return value, err
}
