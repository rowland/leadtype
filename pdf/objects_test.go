package pdf

import (
	"testing"
	"bytes"
)

func expect(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected |%s|, got |%s|", expected, actual)
	}
}

func TestArray(t *testing.T) {
	var buf bytes.Buffer
	ary := array{name("name"), integer(7)}
	ary.write(&buf)

	expect(t, "[/name 7 ] ", buf.String())
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
	expect(t, "0 3\n0000000000 65535 f\n0000000000 00000 n\n0000000018 00000 n\n", buf.String())
}

func TestBoolean(t *testing.T) {
	var buf bytes.Buffer
	boolean(true).write(&buf)
	boolean(false).write(&buf)

	expect(t, "true false ", buf.String())
}

func TestDictionary(t *testing.T) {
	var buf bytes.Buffer
	d := dictionary{"foo": str("bar"), "baz": integer(7)}
	d.write(&buf)

	expect(t, "<<\n/baz 7 \n/foo (bar) \n>>\n", buf.String())
}

func TestDictionaryObject(t *testing.T) {
	var buf bytes.Buffer
	d := &dictionaryObject{indirectObject{1, 0, nil}, dictionary{"foo": str("bar"), "baz": integer(7)}}
	d.write(&buf)
	expect(t, "1 0 obj\n<<\n/baz 7 \n/foo (bar) \n>>\nendobj\n", buf.String())
}

func TestFreeXRefEntry(t *testing.T) {
	var buf bytes.Buffer
	e := &freeXRefEntry{1, 0, nil}
	e.write(&buf)
	expect(t, "0000000001 00000 f\n", buf.String())
}

func TestHeader(t *testing.T) {
	var buf bytes.Buffer
	h := &header{}
	h.write(&buf)

	expect(t, "%PDF-1.3\n", buf.String())
}

func TestIndirectObject(t *testing.T) {
	var buf bytes.Buffer
	obj := &indirectObject{1, 0, nil}
	obj.writeHeader(&buf)
	expect(t, "1 0 obj\n", buf.String())
	buf.Reset()
	obj.writeFooter(&buf)
	expect(t, "endobj\n", buf.String())
	buf.Reset()
	obj.write(&buf)
	expect(t, "1 0 obj\nendobj\n", buf.String())
}

func TestIndirectObjectRef(t *testing.T) {
	var buf bytes.Buffer
	obj := &indirectObject{1, 0, nil}
	ind := &indirectObjectRef{obj}
	ind.write(&buf)
	expect(t, "1 0 R ", buf.String())
}

func TestInteger(t *testing.T) {
	var buf bytes.Buffer
	pi := integer(7)
	pi.write(&buf)

	expect(t, "7 ", buf.String())
}

func TestInUseXRefEntry(t *testing.T) {
	var buf bytes.Buffer
	e := &inUseXRefEntry{500, 0}
	e.write(&buf)

	expect(t, "0000000500 00000 n\n", buf.String())
}

func TestName(t *testing.T) {
	var buf bytes.Buffer
	name("name").write(&buf)

	expect(t, "/name ", buf.String())
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

	expect(t, "7 8 9 10.5 11.5 ", buf.String())
}

func TestRectangle(t *testing.T) {
	var buf bytes.Buffer
	r := &rectangle{1, 2, 3, 4}
	r.write(&buf)

	expect(t, "[1 2 3 4 ] ", buf.String())
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

	expect(t, "1 0 obj\n<<\n/Font <<\n/F1 2 0 R \n>>\n\n>>\nendobj\n", buf.String())
}

func TestResources_XObject(t *testing.T) {
	var buf bytes.Buffer
	r := newResources(1, 0)
	obj := &indirectObject{2, 0, nil}
	ref := &indirectObjectRef{obj}
	r.setXObject("Im1", ref)
	r.write(&buf)

	expect(t, "1 0 obj\n<<\n/XObject <<\n/Im1 2 0 R \n>>\n\n>>\nendobj\n", buf.String())
}

func TestStr(t *testing.T) {
	var buf bytes.Buffer
	s := str("a\\b(cd)")
	expect(t, "a\\\\b\\(cd\\)", s.escape())
	s.write(&buf)
	expect(t, "(a\\\\b\\(cd\\)) ", buf.String())
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
	expect(t, "trailer\n<<\n/Size 3 \n>>\nstartxref\n0\n%%EOF\n", buf.String())
}

func TestXRefSubSection(t *testing.T) {
	var buf bytes.Buffer
	ss := newXRefSubSection()
	ss.add(&inUseXRefEntry{0, 0})
	ss.add(&inUseXRefEntry{100, 1})
	ss.write(&buf)
	expect(t, "0 3\n0000000000 65535 f\n0000000000 00000 n\n0000000100 00001 n\n", buf.String())
}

func TestXRefTable(t *testing.T) {
	var buf bytes.Buffer
	table := &xRefTable{}
	ss := newXRefSubSection()
	ss.add(&inUseXRefEntry{0, 0})
	ss.add(&inUseXRefEntry{100, 1})
	table.add(ss)
	table.write(&buf)
	expect(t, "xref\n0 3\n0000000000 65535 f\n0000000000 00000 n\n0000000100 00001 n\n", buf.String())
}
