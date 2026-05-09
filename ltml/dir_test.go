package ltml

import "testing"

func TestParseDir(t *testing.T) {
	tests := []struct {
		input string
		want  Dir
	}{
		{"ltr", DirLTR},
		{"rtl", DirRTL},
		{"", DirLTR},
		{"RTL", DirLTR}, // case-sensitive
		{"invalid", DirLTR},
	}
	for _, tt := range tests {
		if got := ParseDir(tt.input); got != tt.want {
			t.Errorf("ParseDir(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestDir_String(t *testing.T) {
	if got := DirLTR.String(); got != "ltr" {
		t.Errorf("DirLTR.String() = %q, want %q", got, "ltr")
	}
	if got := DirRTL.String(); got != "rtl" {
		t.Errorf("DirRTL.String() = %q, want %q", got, "rtl")
	}
}

func TestDir_Inheritance(t *testing.T) {
	// Root container defaults to LTR
	root := &StdContainer{}
	if got := root.Dir(); got != DirLTR {
		t.Fatalf("root Dir() = %v, want DirLTR", got)
	}

	// Child inherits from parent
	root.dirExplicit = true
	root.dir = DirRTL
	child := &StdContainer{}
	child.SetContainer(root)
	if got := child.Dir(); got != DirRTL {
		t.Fatalf("child Dir() = %v, want DirRTL (inherited)", got)
	}

	// Explicit LTR overrides inherited RTL
	child.dirExplicit = true
	child.dir = DirLTR
	if got := child.Dir(); got != DirLTR {
		t.Fatalf("child Dir() = %v, want DirLTR (explicit override)", got)
	}

	// Grandchild inherits from child (LTR), not root (RTL)
	grandchild := &StdContainer{}
	grandchild.SetContainer(child)
	if got := grandchild.Dir(); got != DirLTR {
		t.Fatalf("grandchild Dir() = %v, want DirLTR (inherited from child)", got)
	}
}

func TestDir_SetAttrs(t *testing.T) {
	c := &StdContainer{}
	c.SetAttrs(map[string]string{"dir": "rtl"})
	if got := c.Dir(); got != DirRTL {
		t.Fatalf("Dir() after SetAttrs = %v, want DirRTL", got)
	}

	c2 := &StdContainer{}
	c2.SetAttrs(map[string]string{"dir": "ltr"})
	if got := c2.Dir(); got != DirLTR {
		t.Fatalf("Dir() after SetAttrs = %v, want DirLTR", got)
	}
	if !c2.dirExplicit {
		t.Fatal("dirExplicit should be true after SetAttrs")
	}
}

// Helper to make an RTL container with known geometry.
func rtlContainer(left, top, width, height float64) *StdContainer {
	c := positionedContainer(left, top, width, height)
	c.dirExplicit = true
	c.dir = DirRTL
	return c
}

func TestLayoutFlow_RTL(t *testing.T) {
	// Container: left=0, top=0, width=200, height=100, no padding/margin.
	ltrC := positionedContainer(0, 0, 200, 100)
	rtlC := rtlContainer(0, 0, 200, 100)

	style := &LayoutStyle{}

	// Three widgets of width 50, height 20.
	makeWidgets := func(c *StdContainer) []*positionedTestWidget {
		var ws []*positionedTestWidget
		for i := 0; i < 3; i++ {
			w := &positionedTestWidget{preferredWidth: 50, preferredHeight: 20}
			w.SetWidth(50)
			w.SetContainer(c)
			c.AddChild(w)
			ws = append(ws, w)
		}
		return ws
	}

	ltrW := makeWidgets(ltrC)
	LayoutFlow(ltrC, style, nil)

	// LTR: widgets at x=0, 50, 100
	if got := ltrW[0].Left(); got != 0 {
		t.Errorf("LTR flow widget 0 left = %v, want 0", got)
	}
	if got := ltrW[1].Left(); got != 50 {
		t.Errorf("LTR flow widget 1 left = %v, want 50", got)
	}
	if got := ltrW[2].Left(); got != 100 {
		t.Errorf("LTR flow widget 2 left = %v, want 100", got)
	}

	rtlW := makeWidgets(rtlC)
	LayoutFlow(rtlC, style, nil)

	// RTL: widgets at x=150, 100, 50 (right edge at 200, 150, 100)
	if got := rtlW[0].Left(); got != 150 {
		t.Errorf("RTL flow widget 0 left = %v, want 150", got)
	}
	if got := rtlW[1].Left(); got != 100 {
		t.Errorf("RTL flow widget 1 left = %v, want 100", got)
	}
	if got := rtlW[2].Left(); got != 50 {
		t.Errorf("RTL flow widget 2 left = %v, want 50", got)
	}
}

func TestLayoutFlow_RTL_Wrapping(t *testing.T) {
	// Container width=100, three widgets of width 40. First two fit, third wraps.
	c := rtlContainer(0, 0, 100, 200)
	style := &LayoutStyle{}

	var ws []*positionedTestWidget
	for i := 0; i < 3; i++ {
		w := &positionedTestWidget{preferredWidth: 40, preferredHeight: 20}
		w.SetWidth(40)
		w.SetContainer(c)
		c.AddChild(w)
		ws = append(ws, w)
	}

	LayoutFlow(c, style, nil)

	// RTL: widget 0 at x=60, widget 1 at x=20, widget 2 wraps to x=60
	if got := ws[0].Left(); got != 60 {
		t.Errorf("RTL wrap widget 0 left = %v, want 60", got)
	}
	if got := ws[1].Left(); got != 20 {
		t.Errorf("RTL wrap widget 1 left = %v, want 20", got)
	}
	if got := ws[2].Left(); got != 60 {
		t.Errorf("RTL wrap widget 2 left = %v, want 60", got)
	}
	if got := ws[2].Top(); got != 20 {
		t.Errorf("RTL wrap widget 2 top = %v, want 20", got)
	}
}

func TestLayoutVBox_RTL(t *testing.T) {
	// Container: width=300, child width=100
	ltrC := positionedContainer(0, 0, 300, 200)
	rtlC := rtlContainer(0, 0, 300, 200)

	style := &LayoutStyle{}

	makeLtrWidget := func(c *StdContainer) *positionedTestWidget {
		w := &positionedTestWidget{preferredWidth: 100, preferredHeight: 30}
		w.SetWidth(100)
		w.SetContainer(c)
		c.AddChild(w)
		return w
	}

	ltrW := makeLtrWidget(ltrC)
	LayoutVBox(ltrC, style, nil)
	// LTR: flush left at x=0
	if got := ltrW.Left(); got != 0 {
		t.Errorf("LTR vbox widget left = %v, want 0", got)
	}

	rtlW := makeLtrWidget(rtlC)
	LayoutVBox(rtlC, style, nil)
	// RTL: flush right, so left = 300 - 100 = 200
	if got := rtlW.Left(); got != 200 {
		t.Errorf("RTL vbox widget left = %v, want 200", got)
	}
}

func TestLayoutVBox_AlignSelfEndMirrorsInRTL(t *testing.T) {
	c := rtlContainer(0, 0, 300, 200)
	style := &LayoutStyle{}

	w := &positionedTestWidget{preferredWidth: 100, preferredHeight: 30}
	w.SetWidth(100)
	w.SetAttrs(map[string]string{"align-self": "end"})
	w.SetContainer(c)
	c.AddChild(w)

	LayoutVBox(c, style, nil)

	if got := w.Left(); got != 0 {
		t.Errorf("rtl vbox end-aligned child left = %v, want 0", got)
	}
}

func TestLayoutHBox_RTL(t *testing.T) {
	c := rtlContainer(0, 0, 300, 100)
	style := &LayoutStyle{}

	// Two widgets, 80 wide each, no alignment set (unaligned).
	var ws []*positionedTestWidget
	for i := 0; i < 2; i++ {
		w := &positionedTestWidget{preferredWidth: 80, preferredHeight: 30}
		w.SetWidth(80)
		w.SetContainer(c)
		c.AddChild(w)
		ws = append(ws, w)
	}

	LayoutHBox(c, style, nil)

	// RTL unaligned: first widget right-edge at 300, so left=220; second at left=140
	if got := ws[0].Left(); got != 220 {
		t.Errorf("RTL hbox widget 0 left = %v, want 220", got)
	}
	if got := ws[1].Left(); got != 140 {
		t.Errorf("RTL hbox widget 1 left = %v, want 140", got)
	}
}

func TestLayoutHBox_RTL_AlignedPanels(t *testing.T) {
	c := rtlContainer(0, 0, 300, 100)
	style := &LayoutStyle{}

	lpanel := &positionedTestWidget{preferredWidth: 60, preferredHeight: 30}
	lpanel.SetWidth(60)
	lpanel.align = AlignLeft
	lpanel.SetContainer(c)
	c.AddChild(lpanel)

	rpanel := &positionedTestWidget{preferredWidth: 60, preferredHeight: 30}
	rpanel.SetWidth(60)
	rpanel.align = AlignRight
	rpanel.SetContainer(c)
	c.AddChild(rpanel)

	LayoutHBox(c, style, nil)

	// RTL: AlignLeft panel goes to the right edge: left = 300 - 60 = 240
	if got := lpanel.Left(); got != 240 {
		t.Errorf("RTL hbox AlignLeft panel left = %v, want 240", got)
	}
	// RTL: AlignRight panel goes to the left edge: left = 0
	if got := rpanel.Left(); got != 0 {
		t.Errorf("RTL hbox AlignRight panel left = %v, want 0", got)
	}
}

func TestLayoutTable_RTL(t *testing.T) {
	c := rtlContainer(0, 0, 300, 200)
	c.cols = 3
	c.order = TableOrderRows
	c.layout = &LayoutStyle{manager: "table"}
	style := &LayoutStyle{}

	var ws []*positionedTestWidget
	for i := 0; i < 3; i++ {
		w := &positionedTestWidget{preferredWidth: 100, preferredHeight: 30}
		w.SetWidth(100)
		w.SetContainer(c)
		c.AddChild(w)
		ws = append(ws, w)
	}

	LayoutTable(c, style, nil)

	// RTL: col 0 at right, col 1 in middle, col 2 at left
	// ContentRight = 300, each column 100 wide, no hpadding
	if got := ws[0].Left(); got != 200 {
		t.Errorf("RTL table col 0 left = %v, want 200", got)
	}
	if got := ws[1].Left(); got != 100 {
		t.Errorf("RTL table col 1 left = %v, want 100", got)
	}
	if got := ws[2].Left(); got != 0 {
		t.Errorf("RTL table col 2 left = %v, want 0", got)
	}
}

func TestLayoutAbsolute_RTL_NoChange(t *testing.T) {
	c := rtlContainer(0, 0, 200, 200)

	w := &positionedTestWidget{preferredWidth: 50, preferredHeight: 50}
	w.SetContainer(c)
	w.position = Absolute
	w.SetLeft(10)
	w.SetTop(10)
	c.AddChild(w)

	style := &LayoutStyle{}
	LayoutAbsolute(c, style, nil)

	// Absolute layout should not be affected by dir
	if got := w.Left(); got != 10 {
		t.Errorf("RTL absolute widget left = %v, want 10 (unchanged)", got)
	}
}
