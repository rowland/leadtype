package ltml

type StdTarget struct {
	StdWidget
}

func (t *StdTarget) PreferredHeight(Writer) float64 {
	return 0
}

func (t *StdTarget) PreferredWidth(Writer) float64 {
	return 0
}

func (t *StdTarget) SetAttrs(attrs map[string]string) {
	t.StdWidget.SetAttrs(attrs)
	t.SetWidth(0)
	t.SetHeight(0)
}

func (t *StdTarget) ZeroFootprint() bool {
	return true
}

func init() {
	registerTag(DefaultSpace, "target", func() any { return &StdTarget{} })
}

var _ HasAttrs = (*StdTarget)(nil)
var _ Identifier = (*StdTarget)(nil)
var _ WantsContainer = (*StdTarget)(nil)
var _ ZeroFootprint = (*StdTarget)(nil)
