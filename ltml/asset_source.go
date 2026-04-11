package ltml

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type assetSourceKind int

const (
	assetSourceUnknown assetSourceKind = iota
	assetSourceAssetFS
	assetSourceFile
	assetSourceNetworkTempFile
)

type assetSourceRef struct {
	kind       assetSourceKind
	identifier string
}

type assetSourceEntry struct {
	ref       assetSourceRef
	attempted bool
	err       error
}

type assetSourceManager struct {
	mu      sync.Mutex
	entries map[string]*assetSourceEntry
}

func newAssetSourceManager() *assetSourceManager {
	m := &assetSourceManager{
		entries: make(map[string]*assetSourceEntry),
	}
	runtime.SetFinalizer(m, (*assetSourceManager).cleanup)
	return m
}

func (m *assetSourceManager) resolve(doc *Doc, container Container, src string) (assetSourceRef, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return assetSourceRef{}, nil
	}

	m.mu.Lock()
	if entry, ok := m.entries[src]; ok && entry.attempted {
		ref, err := entry.ref, entry.err
		m.mu.Unlock()
		return ref, err
	}
	m.mu.Unlock()

	ref, err := resolveAssetSourceRef(doc, container, src)

	m.mu.Lock()
	entry, ok := m.entries[src]
	if !ok {
		entry = &assetSourceEntry{}
		m.entries[src] = entry
	}
	entry.ref = ref
	entry.err = err
	entry.attempted = true
	m.mu.Unlock()

	return ref, err
}

func (m *assetSourceManager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, entry := range m.entries {
		if entry == nil || !entry.attempted || entry.ref.kind != assetSourceNetworkTempFile || entry.ref.identifier == "" {
			continue
		}
		_ = os.Remove(entry.ref.identifier)
		entry.ref = assetSourceRef{}
		entry.attempted = false
		entry.err = nil
	}
	runtime.SetFinalizer(m, nil)
}

func resolveAssetSourceRef(doc *Doc, container Container, src string) (assetSourceRef, error) {
	if assetURL, err := url.Parse(src); err == nil && assetURL.Scheme != "" {
		return resolveURLAssetSourceRef(container, src, assetURL)
	}
	return resolveFileAssetSourceRef(doc, src)
}

func resolveFileAssetSourceRef(doc *Doc, src string) (assetSourceRef, error) {
	if doc != nil && doc.assetFS != nil {
		if !validAssetPath(src) {
			return assetSourceRef{}, fmt.Errorf("invalid asset path %q", src)
		}
		info, err := fs.Stat(doc.assetFS, src)
		if err != nil {
			return assetSourceRef{}, err
		}
		if info.IsDir() {
			return assetSourceRef{}, fmt.Errorf("asset path %q is a directory", src)
		}
		return assetSourceRef{kind: assetSourceAssetFS, identifier: src}, nil
	}
	if doc != nil && doc.sourceDir != "" && !filepath.IsAbs(src) {
		src = filepath.Join(doc.sourceDir, src)
	}
	info, err := os.Stat(src)
	if err != nil {
		return assetSourceRef{}, err
	}
	if info.IsDir() {
		return assetSourceRef{}, fmt.Errorf("asset path %q is a directory", src)
	}
	return assetSourceRef{kind: assetSourceFile, identifier: src}, nil
}

func resolveURLAssetSourceRef(container Container, src string, assetURL *url.URL) (assetSourceRef, error) {
	if assetURL == nil || !assetURL.IsAbs() {
		return assetSourceRef{}, fmt.Errorf("component src URL must be absolute: %q", src)
	}
	switch assetURL.Scheme {
	case "http", "https":
	default:
		return assetSourceRef{}, fmt.Errorf("unsupported component src URL scheme %q", assetURL.Scheme)
	}
	if networkAssetsDisabled.Load() {
		return assetSourceRef{}, fmt.Errorf("network assets are globally disabled")
	}
	stdDoc := (*StdDocument)(nil)
	if container != nil {
		stdDoc = documentForContainer(container)
	}
	if stdDoc == nil || !stdDoc.NetworkAssetsEnabled() {
		return assetSourceRef{}, fmt.Errorf("network assets are disabled for this document")
	}

	resp, err := http.Get(assetURL.String())
	if err != nil {
		return assetSourceRef{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return assetSourceRef{}, fmt.Errorf("fetching %q returned %s", assetURL.String(), resp.Status)
	}

	tmp, err := os.CreateTemp("", "ltml-asset-*")
	if err != nil {
		return assetSourceRef{}, err
	}
	filename := tmp.Name()
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		_ = os.Remove(filename)
		return assetSourceRef{}, err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(filename)
		return assetSourceRef{}, err
	}
	return assetSourceRef{kind: assetSourceNetworkTempFile, identifier: filename}, nil
}

func readAssetSource(doc *Doc, ref assetSourceRef) ([]byte, error) {
	switch ref.kind {
	case assetSourceAssetFS:
		if doc == nil || doc.assetFS == nil {
			return nil, fmt.Errorf("asset filesystem is not configured")
		}
		return fs.ReadFile(doc.assetFS, ref.identifier)
	case assetSourceFile, assetSourceNetworkTempFile:
		return os.ReadFile(ref.identifier)
	default:
		return nil, fmt.Errorf("asset source is not resolved")
	}
}
