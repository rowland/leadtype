package ltml

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/rowland/leadtype/pdf"
)

func applyWriterAccessibility(w Writer, root *StdDocument) {
	if root == nil || !root.ua {
		return
	}
	w.EnableTaggedPDF(true)
}

func withAccessibilityTag(w Writer, tag string, opts pdf.AccessibilityOptions, fn func() error) error {
	if tag != "" {
		var err error
		wrapped := func() {
			if fn != nil {
				err = fn()
			}
		}
		if tagErr := w.WithAccessibilityTag(tag, opts, wrapped); tagErr != nil {
			return tagErr
		}
		return err
	}
	if fn != nil {
		return fn()
	}
	return nil
}

func withAccessibilityArtifact(w Writer, fn func() error) error {
	var err error
	wrapped := func() {
		if fn != nil {
			err = fn()
		}
	}
	if tagErr := w.WithAccessibilityArtifact(wrapped); tagErr != nil {
		return tagErr
	}
	return err
}

func writerTaggedPDFEnabled(w Writer) bool {
	return w.TaggedPDFEnabled()
}

var accessibilityRoleRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)

func normalizeAccessibilityRole(role string) string {
	role = strings.TrimSpace(role)
	if role == "" {
		return ""
	}
	if !accessibilityRoleRe.MatchString(role) {
		return ""
	}
	return role
}

func roleIsArtifact(role string) bool {
	return strings.ToLower(strings.TrimSpace(role)) == "artifact"
}

func (w *StdWidget) resolvedAccessibilityRole(defaultRole string) string {
	if w.role != "" && !roleIsArtifact(w.role) {
		if role := normalizeAccessibilityRole(w.role); role != "" {
			return role
		}
	}
	return normalizeAccessibilityRole(defaultRole)
}

// withWidgetRoleAccessibility associates rendered text with a structure
// element. Most text relies on its per-font ToUnicode mapping: adding
// replacement text would make the whole structure element an atomic selection
// unit in PDF viewers. Arabic is the exception because shaping and bidi display
// order can prevent the glyph stream from recovering logical reading order, so
// Arabic widgets retain their resolved text as ActualText.
func withWidgetRoleAccessibility(w Writer, widget *StdWidget, defaultRole, resolvedText string, fn func() error) error {
	if widget == nil {
		if fn != nil {
			return fn()
		}
		return nil
	}
	if !writerTaggedPDFEnabled(w) {
		if fn != nil {
			return fn()
		}
		return nil
	}
	if roleIsArtifact(widget.role) {
		return withAccessibilityArtifact(w, fn)
	}
	role := widget.resolvedAccessibilityRole(defaultRole)
	if role == "" {
		if fn != nil {
			return fn()
		}
		return nil
	}
	return withAccessibilityTag(w, role, pdf.AccessibilityOptions{
		ActualText: accessibilityReplacementText(resolvedText),
		ID:         widget.AccessibilityLogicalID(),
	}, fn)
}

func accessibilityReplacementText(text string) string {
	for _, r := range text {
		if unicode.Is(unicode.Arabic, r) {
			return text
		}
	}
	return ""
}

func withGraphicAccessibility(w Writer, widget *StdWidget, defaultRole string, fn func() error) error {
	if widget == nil {
		if fn != nil {
			return fn()
		}
		return nil
	}
	if !writerTaggedPDFEnabled(w) {
		if fn != nil {
			return fn()
		}
		return nil
	}
	if roleIsArtifact(widget.role) {
		return withAccessibilityArtifact(w, fn)
	}
	if widget.AccessibilityAlt() == "" {
		return withAccessibilityArtifact(w, fn)
	}
	role := widget.resolvedAccessibilityRole(defaultRole)
	if role == "" {
		if fn != nil {
			return fn()
		}
		return nil
	}
	return withAccessibilityTag(w, role, pdf.AccessibilityOptions{
		ActualText: widget.AccessibilityAlt(),
		ID:         widget.AccessibilityLogicalID(),
	}, fn)
}
