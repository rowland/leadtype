# Extending the LeadType Markup Language

This document explains the internal architecture of the `ltml` package and how
to add new elements, style types, and layout managers to the language.

## Architecture Overview

LTML documents are processed in two phases:

1. **Parsing** — An XML decoder walks the document tree, creating Go objects for
   each element and wiring them together via parent/container/scope relationships.
2. **Rendering** — The document tree is traversed top-down, with each element
   delegating to its children and calling into the `Writer` interface to produce
   PDF output.

### Core Types

```
Doc                      — top-level parser; holds a stack and scope stack
  └─ StdDocument         — root <ltml> element; embeds StdPage
       └─ StdPage        — <page>; has a Scope for style lookup
            └─ StdContainer — <div>, <table>, etc.; manages children
                 └─ StdWidget — base for all visual elements
```

**Key interfaces** (all in `package ltml`):

| Interface         | Purpose |
|-------------------|---------|
| `Widget`          | Any visual element: geometry, spacing, printing. |
| `Container`       | A `Widget` that holds child widgets. |
| `Printer`         | The four rendering hooks (`BeforePrint`, `PaintBackground`, `DrawBorder`, `DrawContent`). |
| `Styler`          | A named, cloneable style object that can be applied to a `Writer`. |
| `HasScope`        | An element that owns a style/layout/alias namespace. |
| `WantsScope`      | An element that needs a reference to the nearest enclosing `HasScope`. |
| `WantsContainer`  | An element that needs a reference to its parent `Container`. |
| `HasParent`       | An element that needs a reference to its parent element. |
| `HasAttrs`        | An element that accepts a `map[string]string` of attributes. |
| `HasAttrsPrefix`  | Like `HasAttrs` but receives a string prefix (used by style objects). |
| `HasText`         | An element that receives character data from the XML stream. |
| `HasComment`      | An element that receives XML comment data. |
| `Identifier`      | An element that stores `id`, `class`, and `tag` for selector matching. |
| `HasPath`         | An element that can report its selector path (used by rule matching). |

---

## Parsing Flow

`Doc.ParseReader` drives a standard `xml.Decoder` loop. For each token:

- **`xml.StartElement`** → `Doc.startElement`
  1. Checks the scope for an alias matching the tag name.
  2. Calls `makeElement(namespace, tag)` to construct an instance via the tag
     factory registry.
  3. If the element implements `WantsScope`, injects the current scope.
  4. If the element implements `WantsContainer` / `HasParent`, links it to the
     current container/parent on the stack.
  5. Merges attributes in priority order: alias defaults → matching rule
     attributes → direct XML attributes.
  6. If the element is a `Styler`, registers it in the current scope.
  7. If the element is a `*LayoutStyle`, `*Alias`, or `*Rules`, registers it
     in the current scope.
  8. Pushes the element onto the element stack (and scope stack if `HasScope`).

- **`xml.EndElement`** → pops the element (and scope) stack.
- **`xml.CharData`** → calls `AddText` if the top element implements `HasText`.
- **`xml.Comment`** → calls `AddComment` if the top element implements
  `HasComment`.

---

## Rendering Flow

`Doc.Print(w Writer)` iterates `StdDocument` objects and calls `Print` on each.

For every widget the global `Print` helper is used:

```go
func Print(widget Widget, writer Writer) error {
    widget.BeforePrint(writer)    // page setup, layout computation
    widget.PaintBackground(writer) // fill rectangle
    widget.DrawBorder(writer)      // draw borders
    widget.DrawContent(writer)     // recurse into children / draw text
    return nil
}
```

`StdContainer.DrawContent` iterates its children and calls `Print` on each,
producing a depth-first traversal.

Layout is triggered by `StdPage.BeforePrint` (or `StdContainer.LayoutWidget`),
which calls `LayoutContainer → LayoutStyle.Layout → LayoutFunc`.

---

## Tag Factory Registry

Every element type is registered with a name and a factory function:

```go
// Public API — use any namespace except "std"
func RegisterTag(namespace, tag string, f TagFactory) error

// Internal use only
func registerTag(namespace, tag string, f TagFactory) error
```

- `namespace` and `tag` must match `^\w+$`.
- The `"std"` namespace is reserved for built-in elements.
- Tags are looked up as `"namespace:tag"` when the XML decoder resolves names.
- The default XML namespace is `"std"` (set via `dec.DefaultSpace`), so
  unqualified tags like `<p>` resolve to `"std:p"`.

All built-in registrations use `registerTag` inside `init()` functions:

```go
func init() {
    registerTag(DefaultSpace, "p", func() interface{} { return &StdParagraph{} })
}
```

---

## Adding a Custom Element

### 1. Implement the element type

Embed `StdWidget` (or `StdContainer` if the element holds children) to inherit
the standard geometry, border, fill, and font handling:

```go
package mypkg

import "github.com/rowland/leadtype/ltml"

type MyWidget struct {
    ltml.StdWidget
    label string
}

// Implement HasAttrs to receive XML attributes.
func (w *MyWidget) SetAttrs(attrs map[string]string) {
    w.StdWidget.SetAttrs(attrs) // handle width, height, margin, font, border, fill, …
    if v, ok := attrs["label"]; ok {
        w.label = v
    }
}

// Override DrawContent to render custom content.
func (w *MyWidget) DrawContent(writer ltml.Writer) error {
    // use writer methods: MoveTo, Print, SetFont, Rectangle2, …
    return nil
}
```

### 2. Register the tag

```go
func init() {
    ltml.RegisterTag("mypkg", "mywidget", func() interface{} {
        return &MyWidget{}
    })
}
```

### 3. Use in LTML

```xml
<ltml xmlns:mypkg="mypkg">
  <page>
    <mypkg:mywidget label="Hello" width="200pt" height="50pt" />
  </page>
</ltml>
```

---

## Adding a Custom Style

Styles are objects that implement `Styler`:

```go
type Styler interface {
    ID() string
    Apply(Writer)
}
```

They are stored in the scope and looked up by `id`. To participate in attribute
parsing, implement `HasAttrsPrefix`:

```go
type HasAttrsPrefix interface {
    SetAttrs(prefix string, attrs map[string]string)
}
```

The `prefix` parameter allows the same `SetAttrs` method to handle both
definition-time attributes (prefix `""`) and inline overrides (prefix `"mystyle."`).

### Example: a custom gradient style

```go
type GradientStyle struct {
    id         string
    startColor colors.Color
    endColor   colors.Color
}

func (gs *GradientStyle) ID() string { return gs.id }

func (gs *GradientStyle) Apply(w ltml.Writer) {
    // apply gradient via writer methods
}

func (gs *GradientStyle) SetAttrs(prefix string, attrs map[string]string) {
    if id, ok := attrs[prefix+"id"]; ok {
        gs.id = id
    }
    if c, ok := attrs[prefix+"start"]; ok {
        gs.startColor = ltml.NamedColor(c)
    }
    if c, ok := attrs[prefix+"end"]; ok {
        gs.endColor = ltml.NamedColor(c)
    }
}

func init() {
    ltml.RegisterTag("mypkg", "gradient", func() interface{} {
        return &GradientStyle{}
    })
}
```

Usage:

```xml
<ltml xmlns:mypkg="mypkg">
  <mypkg:gradient id="sunset" start="orange" end="purple" />
  <page>
    <rect width="100%" height="1in" fill="sunset" />
  </page>
</ltml>
```

**Note:** For a custom style to be auto-applied by `fill=` or `border=`
references, the built-in `BrushStyleFor` / `PenStyleFor` helpers look up only
`BrushStyle` / `PenStyle` objects. Custom style types must be applied
explicitly by elements that understand them.

---

## Adding a Custom Layout Manager

A layout manager is a function with the signature:

```go
type LayoutFunc func(container Container, style *LayoutStyle, writer Writer)
```

Register it by name:

```go
func init() {
    ltml.RegisterLayoutManager("mygrid", MyGridLayout)
}

func MyGridLayout(c ltml.Container, style *ltml.LayoutStyle, w ltml.Writer) {
    widgets := c.Widgets() // all direct children
    contentLeft := ltml.ContentLeft(c)
    contentTop  := ltml.ContentTop(c)
    contentW    := ltml.ContentWidth(c)

    for i, widget := range widgets {
        widget.SetLeft(contentLeft + float64(i%3) * contentW/3)
        widget.SetTop(contentTop  + float64(i/3) * 100)
        widget.SetWidth(contentW / 3)
        widget.LayoutWidget(w) // recurse into containers
    }
}
```

Use in LTML:

```xml
<ltml xmlns:mypkg="mypkg">
  <page>
    <div layout="mygrid">
      <p>Cell 1</p>
      <p>Cell 2</p>
      <p>Cell 3</p>
    </div>
  </page>
</ltml>
```

### Layout Helper Functions

These are exported from `widget.go` / `container.go` for use in layout
functions:

| Function               | Returns |
|------------------------|---------|
| `ContentLeft(w)`       | Left edge of content area (after margin + padding). |
| `ContentTop(w)`        | Top edge of content area. |
| `ContentRight(w)`      | Right edge of content area. |
| `ContentBottom(w)`     | Bottom edge of content area. |
| `ContentWidth(w)`      | Width of content area. |
| `ContentHeight(w)`     | Height of content area. |
| `NonContentWidth(w)`   | Total horizontal margin + padding. |
| `NonContentHeight(w)`  | Total vertical margin + padding. |
| `MaxContentHeight(c)`  | Maximum content height available (accounts for unbounded containers). |

---

## The Scope System

Each `HasScope` element (currently `StdDocument` and `StdPage`) owns a `Scope`
that stores named styles, layouts, page styles, aliases, and rules. Scopes form
a parent chain; lookups walk from the innermost outward.

```
defaultScope (global)
  └─ StdDocument.Scope
       └─ StdPage.Scope
```

**Default scope contents:**

| Key       | Type        | Value |
|-----------|-------------|-------|
| `solid`   | `*PenStyle` | Black, hairline, solid |
| `dotted`  | `*PenStyle` | Black, hairline, dotted |
| `dashed`  | `*PenStyle` | Black, hairline, dashed |
| `fixed`   | `*FontStyle`| Courier New 12pt |
| `letter`, `legal`, `A4`, `B5`, `C5` | `*PageStyle` | Standard page sizes |
| `vbox`, `hbox`, `table`, `flow`, `absolute`, `relative` | `*LayoutStyle` | Default layouts |
| `h`, `b`, `i`, `u`, `s`, `hbox`, `vbox`, `table`, `layer`, `br` | `*Alias` | Built-in tag aliases |

`HasScope` interface (implemented by `Scope`):

```go
type HasScope interface {
    AddAlias(*Alias) error
    AddLayout(*LayoutStyle) error
    AddPageStyle(*PageStyle) error
    AddRules(*Rules) error
    AddStyle(Styler) error
    AliasFor(name string) (*Alias, bool)
    EachRuleFor(path string, func(*Rule))
    LayoutFor(id string) (*LayoutStyle, bool)
    PageStyleFor(id string) (*PageStyle, bool)
    SetParentScope(HasScope)
    StyleFor(id string) (Styler, bool)
}
```

---

## The Writer Interface

`Writer` (defined in `writer.go`) is the abstraction over the PDF backend. All
drawing operations go through it. Key methods used by built-in elements:

```go
type Writer interface {
    // Text
    SetFont(name string, size float64, opts options.Options)
    SetStrikeout(bool)
    SetUnderline(bool)
    SetLineSpacing(float64)
    FontColor() colors.Color
    FontSize() float64
    Strikeout() bool
    Underline() bool
    LineSpacing() float64
    Fonts() []*font.Font

    // Drawing
    MoveTo(x, y float64)
    LineTo(x, y float64)
    Rectangle2(x, y, w, h float64, stroke, fill bool, corners Corners, ...) error
    SetFillColor(colors.Color)
    SetLineColor(colors.Color)
    SetLineWidth(float64)
    SetLineDashPattern(string)

    // Text output
    Print(s string)
    PrintParagraph(lines []*rich_text.RichText, opts options.Options) error
    Loc() (x, y float64)

    // Page control
    NewPage()
}
```

The production implementation is `ltpdf.DocWriter` in `ltml/ltpdf/doc_writer.go`.

---

## Capability Interfaces Checklist

When implementing a new element, implement only the interfaces it needs:

| Interface        | Implement when… |
|------------------|-----------------|
| `HasAttrs`       | Element reads flat `map[string]string` attributes. |
| `HasAttrsPrefix` | Element is a style object that supports prefixed attribute merging. |
| `HasText`        | Element accepts character data (text nodes). |
| `HasComment`     | Element accepts XML comment nodes. |
| `WantsScope`     | Element needs to look up styles or layouts at parse time. |
| `WantsContainer` | Element must be registered with its parent container. |
| `HasParent`      | Element needs a reference to the parent element (not just the container). |
| `HasScope`       | Element defines a new scope for styles/layouts (creates a scope boundary). |
| `Identifier`     | Element has `id`, `class`, `tag` for CSS selector matching. |
| `HasPath`        | Element can report its selector path for rule matching (use if `Identifier` is implemented). |
| `Styler`         | Element is a reusable style object to be stored in the scope. |
| `Widget`         | Element is a visual element that participates in layout and rendering. |
| `Container`      | Element holds and manages child widgets. |

Embed `StdWidget` to satisfy `Widget` automatically; embed `StdContainer` to
satisfy both `Widget` and `Container`.

---

## Attribute Priority

Attributes are applied to an element in this order (each overrides the previous):

1. **Alias defaults** — `Attrs` map from the matching `*Alias`, if the tag name
   resolved through an alias.
2. **Rule attributes** — for each `Rule` whose selector matches the element's
   path, attributes are applied in document order (outer scope before inner).
3. **Direct XML attributes** — the actual attributes written on the tag in the
   document.

This mirrors CSS specificity: inline styles win over rules, which win over
defaults.
