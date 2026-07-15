package ltml

import "fmt"

func validateDocumentLayouts(root *StdDocument) error {
	if root == nil {
		return nil
	}
	pageStyle := root.pageStyle
	if pageStyle == nil {
		pageStyle = PageStyleFor("letter", root.scope)
	}
	if pageStyle == nil || pageStyle.Width() <= 0 || pageStyle.Height() <= 0 {
		return wrapLayoutError("", root.Path(), fmt.Errorf("default page style is missing or invalid"))
	}

	var validationErr error
	walkWidgets(root, func(widget Widget) bool {
		container, ok := widget.(Container)
		if !ok {
			return true
		}
		validationErr = validateContainerLayout(container)
		return validationErr == nil
	})
	return validationErr
}

func validateContainerLayout(container Container) error {
	if container == nil {
		return wrapLayoutError("", "", fmt.Errorf("container is nil"))
	}
	style := container.LayoutStyle()
	if style == nil {
		return wrapLayoutError("", container.Path(), fmt.Errorf("layout style is nil"))
	}
	if err := validateLayoutInputs(container, style); err != nil {
		return wrapLayoutError(style.manager, container.Path(), err)
	}
	if _, err := LayoutManagerFor(style.manager); err != nil {
		return wrapLayoutError(style.manager, container.Path(), err)
	}

	var err error
	switch style.manager {
	case "table":
		var info *tableGridInfo
		info, err = tableGridFor(container)
		if err == nil {
			_, _, err = tableBodyRange(container, info.grid.Rows())
		}
	case "radial", "radial-out":
		err = validateRadialStructure(container)
	}
	return wrapLayoutError(style.manager, container.Path(), err)
}

func validateRadialStructure(container Container) error {
	var cells []Widget
	for _, child := range container.Widgets() {
		if _, ok := child.(radialCell); !ok {
			return fmt.Errorf("radial layout child %T is not a sector", child)
		}
		cells = append(cells, child)
	}
	if len(cells) == 0 {
		return nil
	}
	rowsHint, colsHint := container.Rows(), container.Cols()
	if base, ok := container.(*StdContainer); ok && len(base.Angles()) > 0 {
		colsFromAngles := radialExplicitSectorCount(base.Angles())
		if colsHint > 0 && colsHint != colsFromAngles {
			return fmt.Errorf("radial cols=%d conflicts with %d normalized angle spans", colsHint, colsFromAngles)
		}
		colsHint = colsFromAngles
	}
	_, _, cols, err := deriveRadialGrid(cells, container.Order(), rowsHint, colsHint)
	if err != nil {
		return err
	}
	_, err = radialAngleSpans(container, cols)
	return err
}
