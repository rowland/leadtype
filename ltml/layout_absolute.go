package ltml

type Position int8

const (
	Static = Position(iota)
	Relative
	Absolute
)

func LayoutAbsolute(container Container, style *LayoutStyle, writer Writer) {
	layoutWidgetsWithPosition(writer, container.Widgets(), Absolute)
}

func LayoutRelative(container Container, style *LayoutStyle, writer Writer) {
	layoutWidgetsWithPosition(writer, container.Widgets(), Relative)
}

func layoutPositionedChildren(container Container, writer Writer) {
	absolute, _ := printableWidgets(container, Absolute)
	layoutWidgetsWithPosition(writer, absolute, Absolute)

	relative, _ := printableWidgets(container, Relative)
	layoutWidgetsWithPosition(writer, relative, Relative)
}

func layoutWidgetsWithPosition(writer Writer, widgets []Widget, position Position) {
	for _, widget := range widgets {
		if widget.Printed() {
			widget.SetVisible(false)
			continue
		}
		widget.SetVisible(true)
		widget.SetPosition(position)
		if !widget.LeftIsSet() && !widget.RightIsSet() {
			widget.SetLeft(0)
		}
		if !widget.TopIsSet() && !widget.BottomIsSet() {
			widget.SetTop(0)
		}
		if !widget.WidthIsSet() {
			widget.ResolveWidth(widget.PreferredWidth(writer))
		}
		widget.LayoutWidget(writer)
		if !widget.HeightIsSet() {
			widget.ResolveHeight(widget.PreferredHeight(writer))
		}
	}
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
