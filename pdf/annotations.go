package pdf

import "io"

type null struct{}

func (n null) write(w io.Writer) {
	_, _ = w.Write([]byte("null "))
}

type linkAnnotation struct {
	dictionaryObject
	targetName string
}

func newLinkAnnotation(seq, gen int, rect rectangle) *linkAnnotation {
	a := new(linkAnnotation)
	a.dictionaryObject.init(seq, gen)
	a.dict["Type"] = name("Annot")
	a.dict["Subtype"] = name("Link")
	a.dict["Rect"] = &rect
	a.dict["Border"] = array{integer(0), integer(0), integer(0)}
	return a
}

func (a *linkAnnotation) setDestination(page *page, x, y float64) {
	delete(a.dict, "Dest")
	// Use an explicit GoTo action rather than a bare /Dest entry because it has
	// been more consistently honored across Preview, Chromium viewers, and PDF.js.
	a.dict["A"] = dictionary{
		"S": name("GoTo"),
		"D": destinationArray(page, x, y),
	}
	a.targetName = ""
}

func (a *linkAnnotation) setTarget(name string) {
	delete(a.dict, "A")
	delete(a.dict, "Dest")
	a.targetName = name
}

func (a *linkAnnotation) setURI(uri string) {
	delete(a.dict, "Dest")
	a.dict["A"] = dictionary{
		"S":   name("URI"),
		"URI": str(uri),
	}
	a.targetName = ""
}

func destinationArray(page *page, x, y float64) array {
	return array{
		&indirectObjectRef{page},
		name("XYZ"),
		number{x},
		number{y},
		null{},
	}
}
