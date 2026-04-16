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
	fileCalls  []imageFilePrintCall
	byteCalls  []imageBytePrintCall
}

type imageFilePrintCall struct {
	filename string
	x        float64
	y        float64
	width    *float64
	height   *float64
}

type imageBytePrintCall struct {
	data   []byte
	x      float64
	y      float64
	width  *float64
	height *float64
}

type cleanupTrackingImageWriter struct {
	imageTestWriter
	cleanups []func()
}

func (w *cleanupTrackingImageWriter) RegisterWriteToCleanup(fn func()) {
	w.cleanups = append(w.cleanups, fn)
}

func (w *imageTestWriter) ImageDimensions(data []byte) (width, height int, err error) {
	return 0, 0, nil
}

func (w *imageTestWriter) ImageDimensionsFromFile(filename string) (width, height int, err error) {
	if dims, ok := w.dimensions[filename]; ok {
		return dims[0], dims[1], nil
	}
	return 0, 0, nil
}

func (w *imageTestWriter) PrintImage(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	w.byteCalls = append(w.byteCalls, imageBytePrintCall{
		data:   data,
		x:      x,
		y:      y,
		width:  width,
		height: height,
	})
	return 0, 0, nil
}

func (w *imageTestWriter) PrintImageFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	w.fileCalls = append(w.fileCalls, imageFilePrintCall{
		filename: filename,
		x:        x,
		y:        y,
		width:    width,
		height:   height,
	})
	return 0, 0, nil
}

func (w *imageTestWriter) PaintImageFile(filename string, x, y, width, height, opacity float64) error {
	_, _, err := w.PrintImageFile(filename, x, y, &width, &height)
	return err
}

func TestStdImage_PreferredWidthAndHeight_UseIntrinsicSize(t *testing.T) {
	img := &StdImage{src: "fixture.jpg"}
	img.SetDoc(newDocWithOptions(WithAssetFS(testingMapFS("fixture.jpg", "image-data"))))
	w := &imageTestWriter{dimensions: map[string][2]int{"fixture.jpg": {144, 96}}}

	if got := img.PreferredWidth(w); got != 144 {
		t.Fatalf("PreferredWidth() = %v, want 144", got)
	}
	if got := img.PreferredHeight(w); got != 96 {
		t.Fatalf("PreferredHeight() = %v, want 96", got)
	}
}

func TestStdImage_PreferredHeight_InfersAspectRatioFromWidth(t *testing.T) {
	img := &StdImage{src: "fixture.jpg"}
	img.SetDoc(newDocWithOptions(WithAssetFS(testingMapFS("fixture.jpg", "image-data"))))
	img.SetWidth(72)
	w := &imageTestWriter{dimensions: map[string][2]int{"fixture.jpg": {144, 96}}}

	if got := img.PreferredHeight(w); got != 48 {
		t.Fatalf("PreferredHeight() = %v, want 48", got)
	}
}

func TestStdImage_PreferredWidth_InfersAspectRatioFromHeight(t *testing.T) {
	img := &StdImage{src: "fixture.jpg"}
	img.SetDoc(newDocWithOptions(WithAssetFS(testingMapFS("fixture.jpg", "image-data"))))
	img.SetHeight(48)
	w := &imageTestWriter{dimensions: map[string][2]int{"fixture.jpg": {144, 96}}}

	if got := img.PreferredWidth(w); got != 72 {
		t.Fatalf("PreferredWidth() = %v, want 72", got)
	}
}

func TestStdImage_DrawContent_UsesContentBoxDimensions(t *testing.T) {
	img := &StdImage{src: "fixture.jpg"}
	img.SetDoc(newDocWithOptions(WithAssetFS(testingMapFS("fixture.jpg", "image-data"))))
	img.SetLeft(10)
	img.SetTop(20)
	img.SetWidth(120)
	img.SetHeight(60)
	img.padding.SetAll("6", "pt")
	w := &imageTestWriter{}

	if err := img.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.fileCalls) != 1 {
		t.Fatalf("file call count = %d, want 1", len(w.fileCalls))
	}
	if len(w.byteCalls) != 0 {
		t.Fatalf("byte call count = %d, want 0", len(w.byteCalls))
	}
	call := w.fileCalls[0]
	if call.filename != "fixture.jpg" {
		t.Fatalf("filename = %q, want fixture.jpg", call.filename)
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

func TestParse_ImageTag_UsesAssetFSReferenceWithoutLoadingBytes(t *testing.T) {
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
	img, ok := page.children[0].(*StdImage)
	if !ok {
		t.Fatalf("child type = %T, want *StdImage", page.children[0])
	}
	ref, err := img.assetSource()
	if err != nil {
		t.Fatal(err)
	}
	if ref.kind != assetSourceAssetFS || ref.identifier != "fixture.jpg" {
		t.Fatalf("ref = %#v, want assetFS fixture.jpg", ref)
	}
}

func TestParse_ImageTag_SrcLoadsRelativeToParsedFile(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture.jpg")
	if err := os.WriteFile(fixture, []byte("image-data"), 0o644); err != nil {
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
	ref, err := img.assetSource()
	if err != nil {
		t.Fatal(err)
	}
	if ref.kind != assetSourceFile || ref.identifier != fixture {
		t.Fatalf("ref = %#v, want file %q", ref, fixture)
	}
}

func TestParse_ImageTag_NetworkSrcLoadsToTempFileWhenEnabled(t *testing.T) {
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
	ref1, err := img.assetSource()
	if err != nil {
		t.Fatal(err)
	}
	ref2, err := img.assetSource()
	if err != nil {
		t.Fatal(err)
	}
	if ref1.kind != assetSourceNetworkTempFile || ref1.identifier == "" {
		t.Fatalf("ref1 = %#v, want network temp file", ref1)
	}
	if ref1.identifier != ref2.identifier {
		t.Fatalf("temp file should be reused, got %q then %q", ref1.identifier, ref2.identifier)
	}
	if _, err := os.Stat(ref1.identifier); err != nil {
		t.Fatalf("temp file missing before cleanup: %v", err)
	}
	doc.cleanupAssetSources()
	if _, err := os.Stat(ref1.identifier); !os.IsNotExist(err) {
		t.Fatalf("temp file should be removed after cleanup, stat err=%v", err)
	}
}

func TestDocPrint_RegistersCleanupForNetworkImageTempFiles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("image-data"))
	}))
	defer srv.Close()

	doc, err := Parse([]byte(`
<ltml network-assets="true">
  <page>
    <image src="` + srv.URL + `" width="2in" height="1in" />
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	img, ok := doc.Page(0).Widgets()[0].(*StdImage)
	if !ok {
		t.Fatalf("first child = %T, want *StdImage", doc.Page(0).Widgets()[0])
	}
	ref, err := img.assetSource()
	if err != nil {
		t.Fatal(err)
	}
	writer := &cleanupTrackingImageWriter{
		imageTestWriter: imageTestWriter{dimensions: map[string][2]int{ref.identifier: {144, 72}}},
	}
	if err := doc.Print(writer); err != nil {
		t.Fatal(err)
	}
	if len(writer.cleanups) != 1 {
		t.Fatalf("cleanup count = %d, want 1", len(writer.cleanups))
	}
	if _, err := os.Stat(ref.identifier); err != nil {
		t.Fatalf("temp file missing before registered cleanup: %v", err)
	}
	writer.cleanups[0]()
	if _, err := os.Stat(ref.identifier); !os.IsNotExist(err) {
		t.Fatalf("temp file should be removed after registered cleanup, stat err=%v", err)
	}
}
