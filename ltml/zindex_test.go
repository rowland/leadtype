package ltml

import "testing"

type zIndexWidget struct {
	StdWidget
	name  string
	order *[]string
}

func (w *zIndexWidget) DrawContent(Writer) error {
	*w.order = append(*w.order, w.name)
	return nil
}

func TestStdWidget_SetAttrs_ZIndex(t *testing.T) {
	var widget StdWidget
	widget.SetAttrs(map[string]string{"z_index": "-2"})
	if widget.ZIndex() != -2 {
		t.Fatalf("expected z_index -2, got %d", widget.ZIndex())
	}
}

func TestStdContainer_DrawContent_OrdersChildrenByZIndex(t *testing.T) {
	var container StdContainer
	var order []string

	back := &zIndexWidget{name: "back", order: &order}
	middle := &zIndexWidget{name: "middle", order: &order}
	front := &zIndexWidget{name: "front", order: &order}
	tie1 := &zIndexWidget{name: "tie1", order: &order}
	tie2 := &zIndexWidget{name: "tie2", order: &order}

	back.SetAttrs(map[string]string{"z_index": "-1"})
	middle.SetAttrs(map[string]string{"z_index": "0"})
	front.SetAttrs(map[string]string{"z_index": "2"})
	tie1.SetAttrs(map[string]string{"z_index": "1"})
	tie2.SetAttrs(map[string]string{"z_index": "1"})

	container.AddChild(front)
	container.AddChild(tie1)
	container.AddChild(back)
	container.AddChild(middle)
	container.AddChild(tie2)

	if err := container.DrawContent(&labelTestWriter{}); err != nil {
		t.Fatalf("DrawContent returned error: %v", err)
	}

	want := []string{"back", "middle", "tie1", "tie2", "front"}
	if len(order) != len(want) {
		t.Fatalf("expected %d children printed, got %d (%v)", len(want), len(order), order)
	}
	for i, name := range want {
		if order[i] != name {
			t.Fatalf("expected print order %v, got %v", want, order)
		}
	}
}
