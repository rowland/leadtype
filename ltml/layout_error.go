package ltml

import "fmt"

// LayoutError reports a failure while validating, measuring, or laying out a
// container. Err contains the underlying failure.
type LayoutError struct {
	Manager string
	Path    string
	Err     error
}

func (err *LayoutError) Error() string {
	if err == nil {
		return "<nil>"
	}
	prefix := "layout"
	if err.Manager != "" {
		prefix += " " + err.Manager
	}
	if err.Path != "" {
		prefix += " at " + err.Path
	}
	if err.Err == nil {
		return prefix
	}
	return fmt.Sprintf("%s: %v", prefix, err.Err)
}

func (err *LayoutError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

func wrapLayoutError(manager, path string, err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*LayoutError); ok {
		return err
	}
	return &LayoutError{Manager: manager, Path: path, Err: err}
}

func layoutManagerName(style *LayoutStyle, fallback string) string {
	if style != nil && style.manager != "" {
		return style.manager
	}
	return fallback
}

func validateLayoutInputs(container Container, style *LayoutStyle) error {
	if container == nil {
		return fmt.Errorf("container is nil")
	}
	if style == nil {
		return fmt.Errorf("layout style is nil")
	}
	for i, child := range container.Widgets() {
		if child == nil {
			return fmt.Errorf("child %d is nil", i)
		}
	}
	return nil
}
