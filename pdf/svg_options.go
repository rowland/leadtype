package pdf

import "github.com/rowland/leadtype/options"

// SVGGradientStopOpacityMode controls how SVG gradient stop opacity is emitted
// into PDF.
type SVGGradientStopOpacityMode string

const (
	svgGradientStopOpacityModeOption = "svg_gradient_stop_opacity_mode"

	// SVGGradientStopOpacityModeSoftMask preserves varying stop opacity with a
	// PDF soft mask.
	SVGGradientStopOpacityModeSoftMask SVGGradientStopOpacityMode = "soft-mask"
	// SVGGradientStopOpacityModeCompatibility collapses varying stop opacity to
	// flat object opacity for broader viewer compatibility.
	SVGGradientStopOpacityModeCompatibility SVGGradientStopOpacityMode = "compatibility"
)

// SVGBlendMode controls whether SVG mix-blend-mode declarations are emitted.
type SVGBlendMode string

const (
	svgBlendModeOption = "svg_blend_mode"

	// SVGBlendModeRespect emits supported PDF blend modes for SVG
	// mix-blend-mode declarations.
	SVGBlendModeRespect SVGBlendMode = "respect"
	// SVGBlendModeIgnore drops SVG mix-blend-mode declarations.
	SVGBlendModeIgnore SVGBlendMode = "ignore"
)

// ParseSVGGradientStopOpacityMode parses the canonical LTML attribute values.
func ParseSVGGradientStopOpacityMode(mode string) (SVGGradientStopOpacityMode, bool) {
	switch mode {
	case string(SVGGradientStopOpacityModeSoftMask):
		return SVGGradientStopOpacityModeSoftMask, true
	case string(SVGGradientStopOpacityModeCompatibility):
		return SVGGradientStopOpacityModeCompatibility, true
	default:
		return SVGGradientStopOpacityModeSoftMask, false
	}
}

func (mode SVGGradientStopOpacityMode) String() string {
	return string(mode)
}

func svgGradientStopOpacityMode(opts options.Options) SVGGradientStopOpacityMode {
	if opts == nil {
		return SVGGradientStopOpacityModeSoftMask
	}
	switch mode := opts[svgGradientStopOpacityModeOption].(type) {
	case SVGGradientStopOpacityMode:
		switch mode {
		case SVGGradientStopOpacityModeCompatibility:
			return mode
		default:
			return SVGGradientStopOpacityModeSoftMask
		}
	case string:
		if parsed, ok := ParseSVGGradientStopOpacityMode(mode); ok {
			return parsed
		}
		return SVGGradientStopOpacityModeSoftMask
	default:
		return SVGGradientStopOpacityModeSoftMask
	}
}

// ParseSVGBlendMode parses the canonical LTML attribute values.
func ParseSVGBlendMode(mode string) (SVGBlendMode, bool) {
	switch mode {
	case string(SVGBlendModeRespect):
		return SVGBlendModeRespect, true
	case string(SVGBlendModeIgnore):
		return SVGBlendModeIgnore, true
	default:
		return SVGBlendModeRespect, false
	}
}

func (mode SVGBlendMode) String() string {
	return string(mode)
}

func svgBlendMode(opts options.Options) SVGBlendMode {
	if opts == nil {
		return SVGBlendModeRespect
	}
	switch mode := opts[svgBlendModeOption].(type) {
	case SVGBlendMode:
		switch mode {
		case SVGBlendModeIgnore:
			return mode
		default:
			return SVGBlendModeRespect
		}
	case string:
		if parsed, ok := ParseSVGBlendMode(mode); ok {
			return parsed
		}
		return SVGBlendModeRespect
	default:
		return SVGBlendModeRespect
	}
}
