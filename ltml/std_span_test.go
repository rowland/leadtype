package ltml

import (
	"strings"
	"testing"
)

func TestParse_RejectsSpanDirectChildOfDiv(t *testing.T) {
	_, err := Parse([]byte(`
<ltml>
  <page>
    <div><span>broken</span></div>
  </page>
</ltml>`))
	if err == nil {
		t.Fatal("expected invalid span parent error")
	}
	if !strings.Contains(err.Error(), "span must be child of p, label, sector, a or another span") {
		t.Fatalf("Parse error = %q, want invalid span parent error", err)
	}
}
