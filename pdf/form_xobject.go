package pdf

type pdfForm struct {
	stream
}

func newPDFForm(seq, gen int, data []byte) *pdfForm {
	form := &pdfForm{}
	form.stream.init(seq, gen, data)
	form.dict["Type"] = name("XObject")
	form.dict["Subtype"] = name("Form")
	form.dict["FormType"] = integer(1)
	return form
}

func (f *pdfForm) setBBox(width, height float64) {
	f.dict["BBox"] = array{real(0), real(0), real(width), real(height)}
}

func (f *pdfForm) setResources(r *resources) {
	f.dict["Resources"] = &indirectObjectRef{r}
}

type renderedForm struct {
	form      *pdfForm
	resources *resources
}

func (r *resources) clone(seq, gen int) *resources {
	clone := newResources(seq, gen)
	if procSet, ok := r.dict["ProcSet"]; ok {
		clone.dict["ProcSet"] = procSet
	}
	clone.fonts = cloneDictionary(r.fonts)
	clone.dict["Font"] = clone.fonts
	if len(r.xObjects) > 0 {
		clone.xObjects = cloneDictionary(r.xObjects)
		clone.dict["XObject"] = clone.xObjects
	}
	if len(r.shadings) > 0 {
		clone.shadings = cloneDictionary(r.shadings)
		clone.dict["Shading"] = clone.shadings
	}
	if len(r.patterns) > 0 {
		clone.patterns = cloneDictionary(r.patterns)
		clone.dict["Pattern"] = clone.patterns
	}
	if len(r.extGStates) > 0 {
		clone.extGStates = cloneDictionary(r.extGStates)
		clone.dict["ExtGState"] = clone.extGStates
	}
	return clone
}

func cloneDictionary(src dictionary) dictionary {
	dst := make(dictionary, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}
