package pdf

import (
	"strings"

	"github.com/rowland/leadtype/options"
)

const (
	svgGradientStopOpacityModeOption   = "svg_gradient_stop_opacity_mode"
	svgGradientStopOpacityModeSoftMask = "soft-mask"
	svgGradientStopOpacityModeFlat     = "flat"
)

func normalizeSVGGradientStopOpacityMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "default", "soft-mask", "softmask":
		return svgGradientStopOpacityModeSoftMask
	case "flat", "compatibility", "compatible":
		return svgGradientStopOpacityModeFlat
	default:
		return svgGradientStopOpacityModeSoftMask
	}
}

func svgGradientStopOpacityMode(opts options.Options) string {
	if opts == nil {
		return svgGradientStopOpacityModeSoftMask
	}
	return normalizeSVGGradientStopOpacityMode(opts.StringDefault(svgGradientStopOpacityModeOption, svgGradientStopOpacityModeSoftMask))
}
