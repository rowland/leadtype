package ltml

const (
	Portrait  = 0
	Landscape = 270
)

type PageSize [2]float64

var PageSizes = map[string]PageSize{
	"letter": {612, 792},
	"legal":  {612, 1008},
	"A4":     {595, 842},
	"B5":     {499, 708},
	"C5":     {459, 649},
}

type PageStyle struct {
	id          string
	size        string
	height      float64
	width       float64
	orientation int
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
	if orientation, ok := attrs["orientation"]; ok {
		switch orientation {
		case "portrait":
			ps.orientation = Portrait
		case "landscape":
			ps.orientation = Landscape
		}
	}
	if size, ok := attrs["size"]; ok {
		if sz, ok := PageSizes[size]; ok {
			if ps.orientation == Portrait {
				ps.width, ps.height = sz[0], sz[1]
			} else {
				ps.width, ps.height = sz[1], sz[0]
			}
		}
	}
	if height, ok := attrs["height"]; ok {
		ps.height = ParseMeasurement(height, "pt")
	}
	if width, ok := attrs["width"]; ok {
		ps.width = ParseMeasurement(width, "pt")
	}
}

func (ps *PageStyle) Width() float64 {
	return ps.width
}

func PageStyleFor(id string, scope HasScope) *PageStyle {
	ps, _ := scope.PageStyle(id)
	return ps
}

var _ HasAttrs = (*PageStyle)(nil)

var defaultPageStyles = map[string]*PageStyle{}

func init() {
	for id, sz := range PageSizes {
		defaultPageStyles[id] = &PageStyle{id: id, width: sz[0], height: sz[1]}
	}
	registerTag(DefaultSpace, "page", func() interface{} { return &PageStyle{} })
}
