package ltml

import "testing"

func TestTextStyleResolvedTextAlign(t *testing.T) {
	tests := []struct {
		name  string
		value string
		set   bool
		ltr   HAlign
		rtl   HAlign
	}{
		{name: "omitted", ltr: HAlignLeft, rtl: HAlignRight},
		{name: "start", value: "start", set: true, ltr: HAlignLeft, rtl: HAlignRight},
		{name: "end", value: "end", set: true, ltr: HAlignRight, rtl: HAlignLeft},
		{name: "left", value: "left", set: true, ltr: HAlignLeft, rtl: HAlignLeft},
		{name: "right", value: "right", set: true, ltr: HAlignRight, rtl: HAlignRight},
		{name: "center", value: "center", set: true, ltr: HAlignCenter, rtl: HAlignCenter},
		{name: "justify", value: "justify", set: true, ltr: HAlignJustify, rtl: HAlignJustify},
		{name: "invalid", value: "invalid", set: true, ltr: HAlignLeft, rtl: HAlignLeft},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := &TextStyle{}
			if tt.set {
				style.SetAttrs(map[string]string{"text-align": tt.value})
			}
			ltr := &StdContainer{}
			rtl := &StdContainer{}
			rtl.SetAttrs(map[string]string{"dir": "rtl"})

			if got := style.ResolvedTextAlign(ltr); got != tt.ltr {
				t.Fatalf("LTR alignment = %s, want %s", got, tt.ltr)
			}
			if got := style.ResolvedTextAlign(rtl); got != tt.rtl {
				t.Fatalf("RTL alignment = %s, want %s", got, tt.rtl)
			}
		})
	}
}

func TestParagraphStyleLogicalAlignmentResolvesPerUsingContainer(t *testing.T) {
	style := &ParagraphStyle{}
	style.SetAttrs(map[string]string{"text-align": "start"})

	ltrParent := &StdContainer{}
	rtlParent := &StdContainer{}
	rtlParent.SetAttrs(map[string]string{"dir": "rtl"})

	ltrParagraph := &StdParagraph{}
	ltrParagraph.paragraphStyle = style
	if err := ltrParagraph.SetContainer(ltrParent); err != nil {
		t.Fatal(err)
	}
	rtlParagraph := &StdParagraph{}
	rtlParagraph.paragraphStyle = style
	if err := rtlParagraph.SetContainer(rtlParent); err != nil {
		t.Fatal(err)
	}

	if got := style.ResolvedTextAlign(ltrParagraph); got != HAlignLeft {
		t.Fatalf("style under LTR container = %s, want left", got)
	}
	if got := style.ResolvedTextAlign(rtlParagraph); got != HAlignRight {
		t.Fatalf("style under RTL container = %s, want right", got)
	}
}

func TestParagraphLogicalAlignmentUsesExplicitChildDirection(t *testing.T) {
	parent := &StdContainer{}
	parent.SetAttrs(map[string]string{"dir": "rtl"})
	p := &StdParagraph{}
	if err := p.SetContainer(parent); err != nil {
		t.Fatal(err)
	}
	p.SetAttrs(map[string]string{
		"dir":              "ltr",
		"style.text-align": "start",
	})

	if got := p.ParagraphStyle().ResolvedTextAlign(p); got != HAlignLeft {
		t.Fatalf("explicit LTR child alignment = %s, want left", got)
	}
}

func TestParagraphLogicalAlignmentOptionsArePhysical(t *testing.T) {
	parent := &StdContainer{}
	parent.SetAttrs(map[string]string{"dir": "rtl"})

	for _, tt := range []struct {
		value string
		want  string
	}{
		{value: "start", want: "right"},
		{value: "end", want: "left"},
		{value: "left", want: "left"},
		{value: "right", want: "right"},
	} {
		t.Run(tt.value, func(t *testing.T) {
			p := &StdParagraph{}
			if err := p.SetContainer(parent); err != nil {
				t.Fatal(err)
			}
			p.SetAttrs(map[string]string{"style.text-align": tt.value})
			if got := paragraphTextFillOptions(p).StringDefault("text-align", ""); got != tt.want {
				t.Fatalf("writer text-align = %q, want %q", got, tt.want)
			}
		})
	}
}
