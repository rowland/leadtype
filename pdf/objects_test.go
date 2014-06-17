// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"testing"
)

func stringFromWriter(w writer) string {
	var buf bytes.Buffer
	w.write(&buf)
	return buf.String()
}

func TestArray(t *testing.T) {
	var buf bytes.Buffer
	ary := array{name("name"), integer(7)}
	ary.write(&buf)

	expectS(t, "[/name 7 ] ", buf.String())
}

func TestBody(t *testing.T) {
	ioInt := &indirectObject{1, 0, integer(7)}
	ioStr := &indirectObject{2, 0, str("Hello")}
	var b body
	b.add(ioInt)
	b.add(ioStr)
	ss := newXRefSubSection()
	var buf bytes.Buffer
	b.write(&buf, ss)
	buf.Reset()
	ss.write(&buf)
	expectS(t, "0 3\n0000000000 65535 f\n0000000000 00000 n\n0000000018 00000 n\n", buf.String())
}

func TestBoolean(t *testing.T) {
	var buf bytes.Buffer
	boolean(true).write(&buf)
	boolean(false).write(&buf)

	expectS(t, "true false ", buf.String())
}

func TestCatalog(t *testing.T) {
	ps := newPages(1, 0)
	o := newOutlines(2, 0)
	c := newCatalog(3, 0, "UseNone", ps, o)
	var buf bytes.Buffer
	c.write(&buf)
	expectS(t, "3 0 obj\n<<\n/Outlines 2 0 R \n/PageMode /UseNone \n/Pages 1 0 R \n/Type /Catalog \n>>\nendobj\n", buf.String())
}

func TestDictionary(t *testing.T) {
	var buf bytes.Buffer
	d := dictionary{"foo": str("bar"), "baz": integer(7)}
	d.write(&buf)

	expectS(t, "<<\n/baz 7 \n/foo (bar) \n>>\n", buf.String())
}

func TestDictionaryObject(t *testing.T) {
	var buf bytes.Buffer
	d := &dictionaryObject{indirectObject{1, 0, nil}, dictionary{"foo": str("bar"), "baz": integer(7)}}
	d.write(&buf)
	expectS(t, "1 0 obj\n<<\n/baz 7 \n/foo (bar) \n>>\nendobj\n", buf.String())
}

func TestFile(t *testing.T) {
	f := newFile()
	expectS(t, "%PDF-1.3\n", stringFromWriter(&f.header))
	var buf bytes.Buffer
	var ss xRefSubSection
	f.body.write(&buf, &ss)
	expectS(t, "", buf.String())
	expectS(t, "trailer\n<<\n>>\nstartxref\n0\n%%EOF\n", stringFromWriter(&f.trailer))
	expectS(t, "%PDF-1.3\nxref\n0 1\n0000000000 65535 f\ntrailer\n<<\n/Size 1 \n>>\nstartxref\n9\n%%EOF\n", stringFromWriter(f))
}

func TestFontDescriptor(t *testing.T) {
	fd := newFontDescriptor(100, 0,
		"ArialMT", "Arial",
		32,
		[]int{-665, -325, 2029, 1006},
		0, 0, 0,
		9.1,
		723, 525, 905, -212, 33, 2000, 0)
	expected := "100 0 obj\n" +
		"<<\n" +
		"/Ascent 905 \n" +
		"/AvgWidth 0 \n" +
		"/CapHeight 723 \n" +
		"/Descent -212 \n" +
		"/Flags 32 \n" +
		"/FontBBox [-665 -325 2029 1006 ] \n" +
		"/FontFamily (Arial) \n" +
		"/FontName /ArialMT \n" +
		"/ItalicAngle 9.1 \n" +
		"/Leading 33 \n" +
		"/MaxWidth 2000 \n" +
		"/MissingWidth 0 \n" +
		"/StemH 0 \n" +
		"/StemV 0 \n" +
		"/Type /FontDescriptor \n" +
		"/XHeight 525 \n" +
		">>\n" +
		"endobj\n"
	var buf bytes.Buffer
	fd.write(&buf)
	expectS(t, expected, buf.String())
}

func TestFreeXRefEntry(t *testing.T) {
	var buf bytes.Buffer
	e := &freeXRefEntry{1, 0, nil}
	e.write(&buf)
	expectS(t, "0000000001 00000 f\n", buf.String())
}

func TestHeader(t *testing.T) {
	var buf bytes.Buffer
	h := &header{}
	h.write(&buf)

	expectS(t, "%PDF-1.3\n", buf.String())
}

func TestIndirectObject(t *testing.T) {
	var buf bytes.Buffer
	obj := &indirectObject{1, 0, nil}
	obj.writeHeader(&buf)
	expectS(t, "1 0 obj\n", buf.String())
	buf.Reset()
	obj.writeFooter(&buf)
	expectS(t, "endobj\n", buf.String())
	buf.Reset()
	obj.write(&buf)
	expectS(t, "1 0 obj\nendobj\n", buf.String())
}

func TestIndirectObjectRef(t *testing.T) {
	var buf bytes.Buffer
	obj := &indirectObject{1, 0, nil}
	ind := &indirectObjectRef{obj}
	ind.write(&buf)
	expectS(t, "1 0 R ", buf.String())
}

func TestInteger(t *testing.T) {
	var buf bytes.Buffer
	pi := integer(7)
	pi.write(&buf)

	expectS(t, "7 ", buf.String())
}

func TestInUseXRefEntry(t *testing.T) {
	var buf bytes.Buffer
	e := &inUseXRefEntry{500, 0}
	e.write(&buf)

	expectS(t, "0000000500 00000 n\n", buf.String())
}

func TestName(t *testing.T) {
	var buf bytes.Buffer
	name("name").write(&buf)

	expectS(t, "/name ", buf.String())
}

func nameShouldEqual(t *testing.T, expected string, actual writer) {
	if name(expected) != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestNameArray(t *testing.T) {
	a := nameArray("PDF", "Text", "ImageB", "ImageC")
	nameShouldEqual(t, "PDF", a[0])
	nameShouldEqual(t, "Text", a[1])
	nameShouldEqual(t, "ImageB", a[2])
	nameShouldEqual(t, "ImageC", a[3])
}

func TestNumber(t *testing.T) {
	var buf bytes.Buffer
	ni := int(7)
	ni32 := int32(8)
	ni64 := int64(9)
	f32 := float32(10.5)
	f64 := float64(11.5)
	number{ni}.write(&buf)
	number{ni32}.write(&buf)
	number{ni64}.write(&buf)
	number{f32}.write(&buf)
	number{f64}.write(&buf)

	expectS(t, "7 8 9 10.5 11.5 ", buf.String())
}

func TestOutlines(t *testing.T) {
	o := newOutlines(1, 0)

	if o.dict["Type"] != name("Outlines") {
		t.Error("outlines not initialized properly")
	}
	var buf bytes.Buffer
	o.write(&buf)
	expectS(t, "1 0 obj\n<<\n/Count 0 \n/Type /Outlines \n>>\nendobj\n", buf.String())
}

func TestPage(t *testing.T) {
	ps := newPages(1, 0)
	p := newPage(2, 0, ps)
	s := newStream(3, 0, []byte("test"))
	p.add(s)

	// initialization
	if p.dict["Type"] != name("Page") {
		t.Error("page not initialized properly")
	}

	// writeBody
	var buf bytes.Buffer
	p.writeBody(&buf)
	expectS(t, "<<\n/Contents 3 0 R \n/Length 4 \n/Parent 1 0 R \n/Type /Page \n>>\n", buf.String())

	// Contents
	expectS(t, stringFromWriter(&indirectObjectRef{s}), stringFromWriter(p.dict["Contents"]))
	p.add(s)
	p.writeBody(&buf)
	buf.Reset()
	p.dict["Contents"].write(&buf)
	expectS(t, "[3 0 R 3 0 R ] ", buf.String())

	// setThumb
	p.setThumb(s)
	expectS(t, stringFromWriter(&indirectObjectRef{s}), stringFromWriter(p.dict["Thumb"]))
	// setAnnots
	// TODO
	// setBeads
	// TODO
}

func TestPageBase(t *testing.T) {
	obj := &indirectObject{1, 0, nil}
	base := newPageBase(2, 0, obj)
	var buf bytes.Buffer
	base.dict["Parent"].write(&buf)
	expectS(t, "1 0 R ", buf.String())

	r := rectangle{1, 2, 3, 4}
	base.setMediaBox(r)
	buf.Reset()
	base.dict["MediaBox"].write(&buf)
	expectS(t, "[1 2 3 4 ] ", buf.String())

	resources := newResources(3, 0)
	base.setResources(resources)
	buf.Reset()
	base.dict["Resources"].write(&buf)
	expectS(t, "3 0 R ", buf.String())

	base.setCropBox(r)
	buf.Reset()
	base.dict["CropBox"].write(&buf)
	expectS(t, "[1 2 3 4 ] ", buf.String())

	base.setRotate(90)
	if base.dict["Rotate"] != integer(90) {
		t.Errorf("Error setting rotate")
	}

	base.setDuration(5)
	buf.Reset()
	base.dict["Dur"].write(&buf)
	expectS(t, "5 ", buf.String())

	base.setHidden(true)
	buf.Reset()
	base.dict["Hid"].write(&buf)
	expectS(t, "true ", buf.String())

	// TODO: setTransition & setAdditionalActions
}

func TestPages(t *testing.T) {
	ps := newPages(1, 0)
	p := newPage(2, 0, ps)
	if ps.dict["Type"] != name("Pages") {
		t.Error("pages not initialized properly")
	}
	ps.add(p)
	var buf bytes.Buffer
	ps.write(&buf)
	expectS(t, "1 0 obj\n<<\n/Count 1 \n/Kids [2 0 R ] \n/Type /Pages \n>>\nendobj\n", buf.String())
}

func TestReal(t *testing.T) {
	var buf bytes.Buffer
	r := real(9.1)
	r.write(&buf)

	expectS(t, "9.1 ", buf.String())
}

func TestRectangle(t *testing.T) {
	var buf bytes.Buffer
	r := &rectangle{1, 2, 3, 4}
	r.write(&buf)

	expectS(t, "[1 2 3 4 ] ", buf.String())
}

func TestResources_ProcSet(t *testing.T) {
	r := newResources(1, 0)
	a := nameArray("PDF", "Text", "ImageB", "ImageC")
	r.setProcSet(a)

	if r.dict["ProcSet"] == nil {
		t.Error("setProcSet: failed")
	}
}

func TestResources_Font(t *testing.T) {
	var buf bytes.Buffer
	r := newResources(1, 0)
	obj := &indirectObject{2, 0, nil}
	ref := &indirectObjectRef{obj}
	r.setFont("F1", ref)
	r.write(&buf)

	expectS(t, "1 0 obj\n<<\n/Font <<\n/F1 2 0 R \n>>\n\n>>\nendobj\n", buf.String())
}

func TestResources_XObject(t *testing.T) {
	var buf bytes.Buffer
	r := newResources(1, 0)
	obj := &indirectObject{2, 0, nil}
	ref := &indirectObjectRef{obj}
	r.setXObject("Im1", ref)
	r.write(&buf)

	expectS(t, "1 0 obj\n<<\n/XObject <<\n/Im1 2 0 R \n>>\n\n>>\nendobj\n", buf.String())
}

func TestStr(t *testing.T) {
	var buf bytes.Buffer
	s := str("a\\b(cd)")
	expectS(t, "a\\\\b\\(cd\\)", string(s.escape()))
	s.write(&buf)
	expectS(t, "(a\\\\b\\(cd\\)) ", buf.String())
}

func TestStream(t *testing.T) {
	s := newStream(1, 0, []byte("test"))
	s.setFilter("bogus")

	if s.dict["Filter"] != name("bogus") {
		t.Error("setFilter: failed")
	}
	if s.dict["Length"] != integer(4) {
		t.Errorf("Length: expected %v, got %v", integer(4), s.dict["Length"])
	}
	var buf bytes.Buffer
	s.writeBody(&buf)
	expectS(t, "<<\n/Filter /bogus \n/Length 4 \n>>\nstream\ntestendstream\n", buf.String())
	buf.Reset()
	s.write(&buf)
	expectS(t, "1 0 obj\n<<\n/Filter /bogus \n/Length 4 \n>>\nstream\ntestendstream\nendobj\n", buf.String())
}

func TestTrailer(t *testing.T) {
	var buf bytes.Buffer
	tr := newTrailer()

	if tr.xrefTableSize() != 0 {
		t.Errorf("xrefTableSize: Expected, %d, got %d", 0, tr.xrefTableSize())
	}

	tr.setXrefTableSize(3)

	if tr.xrefTableStart != 0 {
		t.Errorf("xrefTableStart: Expected %d, got %d", 0, tr.xrefTableStart)
	}
	if tr.xrefTableSize() != 3 {
		t.Errorf("xrefTableSize: Expected %d, got %d", 3, tr.xrefTableSize())
	}
	if tr.dict["Size"] != integer(3) {
		t.Errorf("Size: Expected %d, got %d", 3, tr.dict["Size"])
	}
	tr.write(&buf)
	expectS(t, "trailer\n<<\n/Size 3 \n>>\nstartxref\n0\n%%EOF\n", buf.String())
}

func TestXRefSubSection(t *testing.T) {
	var buf bytes.Buffer
	ss := newXRefSubSection()
	ss.add(&inUseXRefEntry{0, 0})
	ss.add(&inUseXRefEntry{100, 1})
	ss.write(&buf)
	expectS(t, "0 3\n0000000000 65535 f\n0000000000 00000 n\n0000000100 00001 n\n", buf.String())
}

func TestXRefTable(t *testing.T) {
	var buf bytes.Buffer
	table := &xRefTable{}
	ss := newXRefSubSection()
	ss.add(&inUseXRefEntry{0, 0})
	ss.add(&inUseXRefEntry{100, 1})
	table.add(ss)
	table.write(&buf)
	expectS(t, "xref\n0 3\n0000000000 65535 f\n0000000000 00000 n\n0000000100 00001 n\n", buf.String())
}
