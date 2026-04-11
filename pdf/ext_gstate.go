package pdf

type extGState struct {
	dictionaryObject
}

func newExtGState(seq, gen int) *extGState {
	gs := &extGState{}
	gs.dictionaryObject.init(seq, gen)
	gs.dict["Type"] = name("ExtGState")
	return gs
}

func (gs *extGState) setFillAlpha(alpha float64) {
	gs.dict["ca"] = real(alpha)
}

func (gs *extGState) setStrokeAlpha(alpha float64) {
	gs.dict["CA"] = real(alpha)
}

func (gs *extGState) setBlendMode(blendMode string) {
	gs.dict["BM"] = name(blendMode)
}

func (gs *extGState) setSoftMask(form seqGen, subtype string, backdrop []float64) {
	mask := dictionary{
		"S": name(subtype),
		"G": &indirectObjectRef{form},
	}
	if len(backdrop) > 0 {
		mask["BC"] = newRealArray(backdrop...)
	}
	gs.dict["SMask"] = mask
}
