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
	ary := Array{Name("name"), Integer(7)}
	ary.Write(&buf)

	expect(t, "[/name 7 ] ", buf.String())
}

func TestBoolean(t *testing.T) {
	var buf bytes.Buffer
	Boolean(true).Write(&buf)
	Boolean(false).Write(&buf)

	expect(t, "true false ", buf.String())
}

func TestDictionary(t *testing.T) {
	var buf bytes.Buffer
	d := Dictionary{"foo": String("bar"), "baz": Integer(7)}
	d.Write(&buf)

	expect(t, "<<\n/baz 7 \n/foo (bar) \n>>\n", buf.String())
}

func TestDictionaryObject(t *testing.T) {
	var buf bytes.Buffer
	d := &DictionaryObject{IndirectObject{1, 0}, Dictionary{"foo": String("bar"), "baz": Integer(7)}}
	d.Write(&buf)
	expect(t, "1 0 obj\n<<\n/baz 7 \n/foo (bar) \n>>\nendobj\n", buf.String())
}

func TestFreeXRefEntry(t *testing.T) {
	var buf bytes.Buffer
	e := &FreeXRefEntry{1, 0}
	e.Write(&buf)
	expect(t, "0000000001 00000 f\n", buf.String())
}

func TestHeader(t *testing.T) {
	var buf bytes.Buffer
	h := &Header{}
	h.Write(&buf)

	expect(t, "%PDF-1.3\n", buf.String())
}

func TestIndirectObject(t *testing.T) {
	var buf bytes.Buffer
	obj := &IndirectObject{1, 0}
	obj.WriteHeader(&buf)
	expect(t, "1 0 obj\n", buf.String())
	obj.WriteFooter(&buf)
	expect(t, "1 0 obj\nendobj\n", buf.String())
}

func TestIndirectObjectRef(t *testing.T) {
	var buf bytes.Buffer
	obj := &IndirectObject{1, 0}
	ind := &IndirectObjectRef{obj}
	ind.Write(&buf)
	expect(t, "1 0 R ", buf.String())
}

func TestInteger(t *testing.T) {
	var buf bytes.Buffer
	pi := Integer(7)
	pi.Write(&buf)

	expect(t, "7 ", buf.String())
}

func TestInUseXRefEntry(t *testing.T) {
	var buf bytes.Buffer
	e := &InUseXRefEntry{500, 0}
	e.Write(&buf)

	expect(t, "0000000500 00000 n\n", buf.String())
}

func TestName(t *testing.T) {
	var buf bytes.Buffer
	Name("name").Write(&buf)

	expect(t, "/name ", buf.String())
}

func TestNumber(t *testing.T) {
	var buf bytes.Buffer
	ni := int(7)
	ni32 := int32(8)
	ni64 := int64(9)
	f32 := float32(10.5)
	f64 := float64(11.5)
	Number{ni}.Write(&buf)
	Number{ni32}.Write(&buf)
	Number{ni64}.Write(&buf)
	Number{f32}.Write(&buf)
	Number{f64}.Write(&buf)

	expect(t, "7 8 9 10.5 11.5 ", buf.String())
}

func TestRectangle(t *testing.T) {
	var buf bytes.Buffer
	r := &Rectangle{1, 2, 3, 4}
	r.Write(&buf)

	expect(t, "[1 2 3 4 ] ", buf.String())
}

func TestString(t *testing.T) {
	var buf bytes.Buffer
	s := String("a\\b(cd)")
	expect(t, "a\\\\b\\(cd\\)", s.escape())
	s.Write(&buf)
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
	if tr.dict["Size"] != Integer(3) {
		t.Errorf("Size: Expected %d, got %d", 3, tr.dict["Size"])
	}
	tr.Write(&buf)
	expect(t, "trailer\n<<\n/Size 3 \n>>\nstartxref\n0\n%%EOF\n", buf.String())
}

func TestXRefSubSection(t *testing.T) {
	var buf bytes.Buffer
	ss := newXRefSubSection()
	ss.add(&InUseXRefEntry{0, 0})
	ss.add(&InUseXRefEntry{100, 1})
	ss.Write(&buf)
	expect(t, "0 3\n0000000000 65535 f\n0000000000 00000 n\n0000000100 00001 n\n", buf.String())
}

func TestXRefTable(t *testing.T) {
	var buf bytes.Buffer
	table := &XRefTable{}
	ss := newXRefSubSection()
	ss.add(&InUseXRefEntry{0, 0})
	ss.add(&InUseXRefEntry{100, 1})
	table.add(ss)
	table.Write(&buf)
	expect(t, "xref\n0 3\n0000000000 65535 f\n0000000000 00000 n\n0000000100 00001 n\n", buf.String())
}
