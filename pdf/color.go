package pdf

type Color int32

func (this Color) RGB() (r, g, b uint8) {
	b = uint8(this & 0xFF)
	g = uint8((this >> 8) & 0xFF)
	r = uint8((this >> 16) & 0xFF)
	return
}

func (this Color) RGB64() (r, g, b float64) {
	ri, gi, bi := this.RGB()
	r, g, b = float64(ri)/255, float64(gi)/255, float64(bi)/255
	return
}
