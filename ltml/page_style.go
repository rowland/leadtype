package ltml

const (
	Portrait  = 0
	Landscape = 270
)

type PageSize [2]float64

type PageStyle struct {
	id          string
	size        string
	height      float64
	width       float64
	orientation int
}

func (ps *PageStyle) Clone() *PageStyle {
	if ps == nil {
		return nil
	}
	clone := *ps
	return &clone
}

func (ps *PageStyle) ID() string {
	return ps.id
}

func (ps *PageStyle) Height() float64 {
	return ps.height
}

func (ps *PageStyle) Orientation() int {
	return ps.orientation
}

func (ps *PageStyle) SetAttrs(attrs map[string]string) {
	if id, ok := attrs["id"]; ok {
		ps.id = id
	}
	var units Units = "pt"
	units.SetAttrs(attrs)
	prevOrientation := ps.orientation
	orientationChanged := false
	if orientation, ok := attrs["orientation"]; ok {
		switch orientation {
		case "portrait":
			ps.orientation = Portrait
			orientationChanged = ps.orientation != prevOrientation
		case "landscape":
			ps.orientation = Landscape
			orientationChanged = ps.orientation != prevOrientation
		}
	}
	hasSize := false
	if size, ok := attrs["size"]; ok {
		hasSize = true
		if sz, ok := lookupBuiltInPageSize(size); ok {
			if ps.orientation == Portrait {
				ps.width, ps.height = sz.Width, sz.Height
			} else {
				ps.width, ps.height = sz.Height, sz.Width
			}
		}
	}
	if height, ok := attrs["height"]; ok {
		ps.height = ParseMeasurement(height, units)
	}
	if width, ok := attrs["width"]; ok {
		ps.width = ParseMeasurement(width, units)
	}
	if orientationChanged && !hasSize {
		_, hasWidth := attrs["width"]
		_, hasHeight := attrs["height"]
		if !hasWidth && !hasHeight && ps.width != 0 && ps.height != 0 {
			ps.width, ps.height = ps.height, ps.width
		}
	}
}

func (ps *PageStyle) Width() float64 {
	return ps.width
}

func PageStyleFor(id string, scope HasScope) *PageStyle {
	if scope != nil {
		if ps, ok := scope.PageStyleFor(id); ok {
			return ps
		}
	}
	if spec, ok := lookupBuiltInPageSize(id); ok {
		return &PageStyle{id: spec.ID, width: spec.Width, height: spec.Height}
	}
	return nil
}

var _ HasAttrs = (*PageStyle)(nil)

func init() {
	registerTag(DefaultSpace, "pagestyle", func() any { return &PageStyle{} })
}
