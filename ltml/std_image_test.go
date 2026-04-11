package ltml

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

type imageTestWriter struct {
	labelTestWriter
	dimensions map[string][2]int
	dataBySrc  map[string][]byte
	calls      []imagePrintCall
}

type imagePrintCall struct {
	data   []byte
	x      float64
	y      float64
	width  *float64
	height *float64
}

func (w *imageTestWriter) ImageDimensions(data []byte) (width, height int, err error) {
	for src, blob := range w.dataBySrc {
		if string(blob) == string(data) {
			if dims, ok := w.dimensions[src]; ok {
				return dims[0], dims[1], nil
			}
		}
	}
	return 0, 0, nil
}

func (w *imageTestWriter) ImageDimensionsFromFile(filename string) (width, height int, err error) {
	if dims, ok := w.dimensions[filename]; ok {
		return dims[0], dims[1], nil
	}
	return 0, 0, nil
}

func (w *imageTestWriter) PrintImage(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	w.calls = append(w.calls, imagePrintCall{
		data:   data,
		x:      x,
		y:      y,
		width:  width,
		height: height,
	})
	return 0, 0, nil
}

func TestStdImage_PreferredWidthAndHeight_UseIntrinsicSize(t *testing.T) {
	img := &StdImage{src: "fixture.jpg", data: []byte("fixture.jpg")}
	w := &imageTestWriter{dimensions: map[string][2]int{"fixture.jpg": {144, 96}}, dataBySrc: map[string][]byte{"fixture.jpg": []byte("fixture.jpg")}}

	if got := img.PreferredWidth(w); got != 144 {
		t.Fatalf("PreferredWidth() = %v, want 144", got)
	}
	if got := img.PreferredHeight(w); got != 96 {
		t.Fatalf("PreferredHeight() = %v, want 96", got)
	}
}

func TestStdImage_PreferredHeight_InfersAspectRatioFromWidth(t *testing.T) {
	img := &StdImage{src: "fixture.jpg", data: []byte("fixture.jpg")}
	img.SetWidth(72)
	w := &imageTestWriter{dimensions: map[string][2]int{"fixture.jpg": {144, 96}}, dataBySrc: map[string][]byte{"fixture.jpg": []byte("fixture.jpg")}}

	if got := img.PreferredHeight(w); got != 48 {
		t.Fatalf("PreferredHeight() = %v, want 48", got)
	}
}

func TestStdImage_PreferredWidth_InfersAspectRatioFromHeight(t *testing.T) {
	img := &StdImage{src: "fixture.jpg", data: []byte("fixture.jpg")}
	img.SetHeight(48)
	w := &imageTestWriter{dimensions: map[string][2]int{"fixture.jpg": {144, 96}}, dataBySrc: map[string][]byte{"fixture.jpg": []byte("fixture.jpg")}}

	if got := img.PreferredWidth(w); got != 72 {
		t.Fatalf("PreferredWidth() = %v, want 72", got)
	}
}

func TestStdImage_DrawContent_UsesContentBoxDimensions(t *testing.T) {
	img := &StdImage{src: "fixture.jpg", data: []byte("fixture.jpg")}
	img.SetLeft(10)
	img.SetTop(20)
	img.SetWidth(120)
	img.SetHeight(60)
	img.padding.SetAll("6", "pt")
	w := &imageTestWriter{dimensions: map[string][2]int{"fixture.jpg": {144, 96}}, dataBySrc: map[string][]byte{"fixture.jpg": []byte("fixture.jpg")}}

	if err := img.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(w.calls))
	}
	call := w.calls[0]
	if got := string(call.data); got != "fixture.jpg" {
		t.Fatalf("data = %q, want fixture.jpg", got)
	}
	if call.x != 16 || call.y != 26 {
		t.Fatalf("x,y = %v,%v, want 16,26", call.x, call.y)
	}
	if call.width == nil || *call.width != 108 {
		t.Fatalf("width = %v, want 108", call.width)
	}
	if call.height == nil || *call.height != 48 {
		t.Fatalf("height = %v, want 48", call.height)
	}
}

func TestParse_ImageTag(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <image src="fixture.jpg" width="2in" />
  </page>
</ltml>`), WithAssetFS(testingMapFS("fixture.jpg", "image-data")))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
	if page == nil {
		t.Fatal("page is nil")
	}
	if len(page.children) != 1 {
		t.Fatalf("child count = %d, want 1", len(page.children))
	}
	img, ok := page.children[0].(*StdImage)
	if !ok {
		t.Fatalf("child type = %T, want *StdImage", page.children[0])
	}
	if img.src != "fixture.jpg" {
		t.Fatalf("src = %q, want %q", img.src, "fixture.jpg")
	}
	if got := string(img.data); got != "image-data" {
		t.Fatalf("data = %q, want image-data", got)
	}
}

func TestParse_ImageTag_SrcLoadsRelativeToParsedFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "fixture.jpg"), []byte("image-data"), 0o644); err != nil {
		t.Fatal(err)
	}
	inputPath := filepath.Join(dir, "doc.ltml")
	if err := os.WriteFile(inputPath, []byte(`
<ltml>
  <page>
    <image src="fixture.jpg" width="2in" />
  </page>
</ltml>`), 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := ParseFile(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	page := doc.Root().Page(0)
	img, ok := page.children[0].(*StdImage)
	if !ok {
		t.Fatalf("child type = %T, want *StdImage", page.children[0])
	}
	if got := string(img.data); got != "image-data" {
		t.Fatalf("data = %q, want image-data", got)
	}
}

func TestParse_ImageTag_NetworkSrcLoadsWhenEnabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("image-data"))
	}))
	defer srv.Close()

	doc, err := Parse([]byte(`
<ltml network-assets="true">
  <page>
    <image src="` + srv.URL + `" width="2in" />
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	page := doc.Root().Page(0)
	img, ok := page.children[0].(*StdImage)
	if !ok {
		t.Fatalf("child type = %T, want *StdImage", page.children[0])
	}
	if got := string(img.data); got != "image-data" {
		t.Fatalf("data = %q, want image-data", got)
	}
}
