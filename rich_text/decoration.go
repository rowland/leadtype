package rich_text

import "github.com/rowland/leadtype/colors"

// DecorationOverrides carries optional render-time decoration overrides for a
// text run. It is intended to be treated as immutable and shared across many
// leaf pieces.
type DecorationOverrides struct {
	Underline  DecorationLineOverrides
	Strikeout  DecorationLineOverrides
	TextStroke DecorationStrokeOverrides
}

type DecorationLineOverrides struct {
	Color       colors.Color
	Width       float64
	Pattern     string
	CapStyle    string
	Position    float64
	HasColor    bool
	HasWidth    bool
	HasPattern  bool
	HasPosition bool
}

type DecorationStrokeOverrides struct {
	Color    colors.Color
	Width    float64
	HasColor bool
	HasWidth bool
}

func (line DecorationLineOverrides) Equal(other DecorationLineOverrides) bool {
	return line.Color == other.Color &&
		line.Width == other.Width &&
		line.Pattern == other.Pattern &&
		line.CapStyle == other.CapStyle &&
		line.Position == other.Position &&
		line.HasColor == other.HasColor &&
		line.HasWidth == other.HasWidth &&
		line.HasPattern == other.HasPattern &&
		line.HasPosition == other.HasPosition
}

func (overrides *DecorationOverrides) Equal(other *DecorationOverrides) bool {
	if overrides == nil || other == nil {
		return overrides == other
	}
	return overrides.Underline.Equal(other.Underline) &&
		overrides.Strikeout.Equal(other.Strikeout) &&
		overrides.TextStroke == other.TextStroke
}
