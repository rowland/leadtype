package ltml

type Position int8

const (
	Static = Position(iota)
	Relative
	Absolute
)

type positionedRadialCenter interface {
	usesPositionedRadialCenter(Position) bool
	resolvePositionedRadialBox(Position)
}

func LayoutAbsolute(container Container, style *LayoutStyle, writer Writer) (err error) {
	defer func() { err = wrapLayoutError("absolute", containerPath(container), err) }()
	if err := validateLayoutInputs(container, style); err != nil {
		return err
	}
	return layoutWidgetsWithPosition(writer, container.Widgets(), Absolute)
}

func LayoutRelative(container Container, style *LayoutStyle, writer Writer) (err error) {
	defer func() { err = wrapLayoutError("relative", containerPath(container), err) }()
	if err := validateLayoutInputs(container, style); err != nil {
		return err
	}
	return layoutWidgetsWithPosition(writer, container.Widgets(), Relative)
}

func layoutPositionedChildren(container Container, writer Writer) error {
	absolute, _ := printableWidgets(container, Absolute)
	if err := layoutWidgetsWithPosition(writer, absolute, Absolute); err != nil {
		return err
	}

	relative, _ := printableWidgets(container, Relative)
	return layoutWidgetsWithPosition(writer, relative, Relative)
}

func layoutWidgetsWithPosition(writer Writer, widgets []Widget, position Position) error {
	for _, widget := range widgets {
		if widget.Printed() {
			widget.SetVisible(false)
			continue
		}
		widget.SetVisible(true)
		widget.SetPosition(position)
		centered, usesCenter := widget.(positionedRadialCenter)
		centerPlacement := usesCenter && centered.usesPositionedRadialCenter(position)
		if !centerPlacement && !widget.LeftIsSet() && !widget.RightIsSet() {
			widget.SetLeft(0)
		}
		if !centerPlacement && !widget.TopIsSet() && !widget.BottomIsSet() {
			widget.SetTop(0)
		}
		if !widget.WidthIsSet() {
			width, err := widget.PreferredWidth(writer)
			if err != nil {
				return err
			}
			widget.ResolveWidth(width)
		}
		if centerPlacement && !widget.HeightIsSet() {
			height, err := widget.PreferredHeight(writer)
			if err != nil {
				return err
			}
			widget.ResolveHeight(height)
		}
		if centerPlacement {
			centered.resolvePositionedRadialBox(position)
		}
		if err := widget.LayoutWidget(writer); err != nil {
			return err
		}
		if !widget.HeightIsSet() {
			height, err := widget.PreferredHeight(writer)
			if err != nil {
				return err
			}
			widget.ResolveHeight(height)
		}
	}
	return nil
}

func printableWidgets(c Container, p Position) (widgets, remaining []Widget) {
	root := rootPageForContainer(c)
	physicalPageNo := 0
	if root != nil {
		if doc := root.document(); doc != nil {
			physicalPageNo = doc.CurrentPhysicalPageNo()
		}
	}
	flowPageIndex := 1
	if root != nil {
		flowPageIndex = root.flowPageIndex
	}
	parentRepeats := containerRepeatsForRender(c, root)
	for _, w := range c.Widgets() {
		if w.Position() == p && widgetDisplayForRender(w, parentRepeats, flowPageIndex, physicalPageNo) {
			widgets = append(widgets, w)
		} else {
			remaining = append(remaining, w)
		}
	}
	return
}

func containerRepeatsForRender(c Container, root *StdPage) bool {
	for current := c; current != nil; current = current.Container() {
		if root != nil && current == root {
			return false
		}
		if current.Display() != DisplayOnce {
			return true
		}
		if root != nil && current.Container() == nil {
			break
		}
	}
	return false
}
