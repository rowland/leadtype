# LeadType Markup Language (LTML) Syntax Reference

LTML is an XML-based markup language for generating PDF documents. It provides
declarative control over layout, typography, and visual elements.

## Document Structure

Every LTML document begins with a root `<ltml>` element containing style
definitions, optional reusable `<canvas>` assets, and one or more `<page>`
elements.

```xml
<ltml units="in" margin="1">
  <!-- style definitions -->
  <canvas key="badge" width="120" height="60">
    <!-- reusable drawing -->
  </canvas>
  <page>
    <!-- content -->
  </page>
</ltml>
```

### Scope

`<ltml>`, `<canvas>`, and `<page>` establish a style scope. Style definitions
(`<font>`, `<pen>`, `<brush>`, `<para>`, `<bullet>`, `<layout>`), aliases
(`<define>`), and selector styles (`<style>`) placed inside a `<canvas>` are
visible only to that canvas capture. Definitions placed inside a `<page>` are
visible only to that page. Definitions placed directly inside `<ltml>` are
visible to all pages and canvases. A page or canvas can always reference
definitions from its parent `<ltml>` scope, but sibling scopes stay isolated.
The attributes on a scope owner may also reference resources declared inside
that same scope. This lets standalone pages and canvases use stable local names
without coordinating IDs with the rest of the document:

```xml
<ltml>
  <font id="body" name="Helvetica" size="12" />  <!-- shared by all pages -->

  <page font="body">
    <font id="body" name="Times New Roman" size="11" />  <!-- shadows the shared body -->
    <font id="title" name="Helvetica" size="24" weight="Bold" />  <!-- this page only -->
    <p font="title">Page One</p>
    <p>Body text.</p>
  </page>

  <page font="body">
    <!-- "title" is not visible here -->
    <p>Page Two uses the shared Helvetica body.</p>
  </page>
</ltml>
```

The same convention works at the document root, including for the document's
default page style:

```xml
<ltml font="body" style="book">
  <font id="body" name="Helvetica" size="11" />
  <pagestyle id="book" units="in" width="6" height="9" />
  <page>
    <p>Document content.</p>
  </page>
</ltml>
```

This is a narrow exception for the owner of a scope, not general declaration
hoisting. Put resource declarations before ordinary content that refers to
them. Dependencies between resource declarations also remain
declaration-ordered.

---

## Elements

### `<ltml>` — Document Root

The root element. Attributes set here apply as defaults to all pages. Direct
children may include style definitions, `<canvas>` definitions, and `<page>`
elements.

| Attribute | Description |
|-----------|-------------|
| `units`   | Default unit for measurements (`pt`, `in`, `cm`, `mm`, `dp`). Default: `pt`. |
| `margin`  | Page margin applied to all pages unless overridden. |
| `compress-pages` | If `true`, compress page content streams with `FlateDecode`. Default: `false`. |
| `compress-to-unicode` | If `true`, compress generated `ToUnicode` streams. Default: `false`. |
| `compress-embedded-fonts` | If `true`, compress embedded font subset streams. Default: `false`. |
| `ua` | If `true`, opt the whole document into tagged PDF output and accessibility structure generation. Default: `false`. |
| `lang` | BCP 47 natural-language tag written to the PDF catalog, such as `ar` or `zh-Hans`. |
| `svg-gradient-stop-opacity-mode` | Default SVG gradient stop-opacity rendering mode for pages. Use `compatibility` to collapse varying stop alpha to flat object opacity for broader PDF viewer compatibility. Default: `soft-mask`. |
| `svg-blend-mode` | Default SVG `mix-blend-mode` handling for pages. Use `ignore` to drop blend modes (matches legacy PDFlib output). Default: `respect` (matches WebKit/Chrome). |

#### Tagged PDF Accessibility

LTML tagged PDF output is opt-in at the document level:

```xml
<ltml ua="true">
  <page>
    <p>Hello <a uri="https://example.com">world</a></p>
  </page>
</ltml>
```

When `ua="true"` is set, LTML emits PDF structure for conservative defaults:

- `<p>` and `<label>` default to `P`
- `<a>` defaults to `Link`
- `<image>`, `<line>`, and shape primitives with `alt` default to `Figure`
- decorative graphics without `alt`, borders, backgrounds, debug grids, and shape chrome are emitted as PDF artifacts

Text widgets normally remain represented by their rendered text and per-font
`/ToUnicode` mappings. LTML does not add `/ActualText` replacement strings to
directly mappable text because replacement text can make an entire structure
element behave as one indivisible selection in PDF viewers. Arabic text retains
resolved `/ActualText` because shaping and bidirectional display order otherwise
prevent reliable recovery of logical reading order. Dynamic text such as
`<pageno>` is rendered directly and remains selectable.

Widget-level accessibility attributes are intentionally small:

- `role` overrides the computed PDF structure role for a participating widget
- `alt` supplies replacement text for participating graphics

LTML reserves `role="artifact"` as a special case. That value is
case-insensitive and suppresses the widget from tagged output instead of
creating a structure element.

If `ua` is absent or not `true`, LTML ignores widget accessibility attributes.

The `ltml.TestSamples` harness follows the same document-driven rule, so sample
PDFs remain untagged unless a sample opts in with `ua="true"` or the test
helper explicitly forces tagged output on its writer.

#### SVG Gradient Stop-Opacity Compatibility

SVG fills that use varying `stop-opacity` can require PDF soft masks, and some
renderers handle those less reliably than Chrome/PDFium. LTML exposes the PDF
writer compatibility switch on the root document as a default and on individual
pages as an override:

```xml
<ltml svg-gradient-stop-opacity-mode="compatibility">
  <page>
    <image src="hero.svg" width="4in" />
  </page>
  <page svg-gradient-stop-opacity-mode="soft-mask">
    <image src="detailed-art.svg" width="4in" />
  </page>
</ltml>
```

Accepted values:

- `soft-mask` — default high-fidelity rendering
- `compatibility` — collapse varying stop alpha to a single flat object opacity

Use `compatibility` when a document needs more consistent rendering in Preview
or PDF.js and the slight loss of intra-gradient transparency is acceptable. See
[docs/svg-rendering-compatibility.md](../docs/svg-rendering-compatibility.md)
for the PDF and Go API versions of the same setting.

#### SVG Blend Mode Compatibility

SVG `mix-blend-mode` declarations (e.g. `hard-light`) are honored by default.
Some legacy SVG→PDF pipelines silently drop blend modes; when artwork was
tuned against that flatter output, honoring the blend mode can produce
unexpectedly bright or saturated regions. Set `svg-blend-mode="ignore"` on the
root document to drop SVG blend modes by default, or on a single page to keep
the compatibility concern local:

```xml
<ltml svg-blend-mode="ignore">
  <page>
    <image src="hero.svg" width="4in" />
  </page>
  <page svg-blend-mode="respect">
    <image src="artwork.svg" width="4in" />
  </page>
</ltml>
```

Accepted values:

- `respect` — default; emit PDF blend-mode entries (matches WebKit/Chrome)
- `ignore` — silently drop blend modes (matches legacy PDFlib output)

The setting is independent of `svg-gradient-stop-opacity-mode` and they can be
combined.

---

### `<canvas>` — Reusable Drawing Asset

Defines a document-scoped reusable drawing with its own local coordinate
system. `<canvas>` is captured once at its natural size and later placed with
`<draw>` one or more times.

`<canvas>` must be a direct child of `<ltml>`. It is not rendered directly into
page flow.

```xml
<ltml units="pt">
  <canvas key="badge" width="160" height="80" layout="absolute">
    <circle left="16" top="16" width="48" height="48" fill.color="#cfe4ff" />
    <label left="76" top="24">Score 92</label>
  </canvas>

  <page>
    <draw key="badge" />
    <draw key="badge" width="80" />
  </page>
</ltml>
```

| Attribute | Description |
|-----------|-------------|
| `key` | Required document-wide asset key used by `<draw>`. Duplicate keys are rejected. |
| `width`, `height` | Required natural canvas size. LTML captures the canvas using this fixed coordinate system. |
| `layout` | Optional layout manager for child widgets. Default: `absolute`. Use `layout.*` for inline overrides. |
| `font` | Optional default font style for child text widgets. |
| `fill` / `fill.*` | Optional background brush for the canvas root box. |
| `border` / `border.*` | Optional border pen for the canvas root box. |
| `padding`, `padding-top`, `padding-right`, `padding-bottom`, `padding-left` | Optional inner spacing before child content begins. |

`<canvas>` supports ordinary LTML child widgets and local style definitions, but
not page-only behavior such as page creation, overflow retries, debug grids, or
page-flow continuation.

Inner page-dependent semantics are treated as visual-only during capture:

- inner links and destinations do not create PDF annotations
- inner `<pageno>` resolves to empty text
- inner `<index>` and `<index-entry>` widgets do not participate in document indexes
- child tagged-PDF structure is suppressed so accessibility stays on the outer `<draw>`

---

### `<page>` — Page

Defines a single page in the document. Pages must be direct children of `<ltml>`.

| Attribute     | Description |
|---------------|-------------|
| `units`       | Unit for measurements on this page. |
| `margin`      | Margin for all sides. |
| `margin-top`, `margin-right`, `margin-bottom`, `margin-left` | Per-side margins. |
| `style`       | Reference to a named `<page>` style. |
| `layout`      | Layout manager to use (`vbox`, `hbox`, `table`, `flow`, `absolute`, `relative`, `radial`, `radial-out`). Default: `vbox`. Use `layout.*` for inline overrides such as `layout.vpadding="9pt"`. |
| `dir`         | Layout direction: `ltr` (default) or `rtl`. Inherited by child containers. Invalid values fall back to `ltr`. |
| `grid`        | Optional debug grid. Use `true` for the default `0.25in` grid or supply a measurement such as `0.5in`. Add a comma-delimited count such as `0.25in,4` or `true,4` to draw every fourth line bolder. |
| `overflow`    | `true` or `false`. If `true`, allow the page to retry unprinted direct children on additional physical pages. Defaults to `true` for page `layout="flow"`, `layout="table"`, and `layout="vbox"`. |
| `svg-gradient-stop-opacity-mode` | Page-local SVG gradient stop-opacity rendering mode. Overrides the document default for SVG assets rendered on this page. |
| `svg-blend-mode` | Page-local SVG `mix-blend-mode` handling. Overrides the document default for SVG assets rendered on this page. |
| `font`        | Reference to a named `<font>` style. |
| `fill`        | Reference to a named `<brush>` style for the background. |
| `border`      | Reference to a named `<pen>` style for all borders. |

---

### `<draw>` — Canvas Placement

Places a named `<canvas>` asset into page or container layout. `<draw>` behaves
like an image-style placement widget.

```xml
<draw key="badge" width="96" alt="Score badge" />
```

| Attribute | Description |
|-----------|-------------|
| `key` | Required canvas key to place. |
| `width`, `height` | Optional explicit placement dimensions. If both are omitted, LTML uses the canvas natural size. If only one is supplied, LTML preserves aspect ratio. If both are supplied, LTML stretches to the exact box. |
| `max-width`, `max-height` | Optional maximum widget dimensions. For image-style widgets, caps preserve aspect ratio and choose whichever dimension dominates. |
| `margin`, `margin-top`, `margin-right`, `margin-bottom`, `margin-left` | Outer spacing around the widget box. |
| `padding`, `padding-top`, `padding-right`, `padding-bottom`, `padding-left` | Inner spacing inside the widget box. |
| `border` | Optional enclosing widget border, separate from the captured canvas content. |
| `fill` | Optional enclosing widget background, separate from the captured canvas content. |
| `rotate`, `origin-x`, `origin-y`, `shift-x`, `shift-y`, `align`, `display` | Same placement and transform attributes supported by other widgets. |
| `alt` | When `ua="true"`, opt the draw placement into tagged output and use this text as `/ActualText`. |
| `role` | Override the default tagged role when `ua="true"`. Draws with `alt` default to `Figure`. |

In v1, `<draw>` places only `<canvas>` assets even though the tag name is
generic.

---

### `<p>` — Paragraph

A block of text. Text content may include inline elements (`<span>`, `<b>`,
`<i>`, `<u>`, `<s>`).

```xml
<p font.weight="Bold" style.text-align="center">Hello, World!</p>
```

| Attribute          | Description |
|--------------------|-------------|
| `font`             | Reference to a named `<font>` style. |
| `font.name`        | Font family name (e.g., `Helvetica`, `Arial`). |
| `font.size`        | Font size as points (`12`) or page-root-relative rems (`1rem`, `0.875rem`). |
| `font.color`       | Font color (named color or hex). |
| `font.weight`      | Font weight (`Bold`, or empty for normal). |
| `font.style`       | Font style (`Italic`, `Oblique`, or empty for normal). |
| `font.underline`   | `true` or `false`. |
| `font.strikeout`   | `true` or `false`. |
| `font.underline-pen` | Reference to a named `<pen>` style used for underline stroke settings. |
| `font.strikeout-pen` | Reference to a named `<pen>` style used for strikeout stroke settings. |
| `font.underline-pos` | Underline position as an LTML measurement. |
| `font.strikeout-pos` | Strikeout position as an LTML measurement. |
| `font.line-height` | Line spacing multiplier (e.g., `1.5`). |
| `text-fill`, `text-fill.*` | Optional text brush. When present, LTML clips the paragraph text against the widget box fill brush instead of using flat `font.color`. Use the same brush attributes supported by `fill` and `fill.*`. |
| `style`            | Reference to a named `<para>` style. |
| `style.text-align` | Text alignment: logical `start`/`end`, physical `left`/`right`, `center`, or `justify`. Defaults to `start`. |
| `style.valign`     | Vertical alignment: `top`, `middle`, `bottom`, `baseline`. |
| `angle`            | Inside a sector, an unset or nonzero angle selects curved paragraph text. `angle="0"` selects horizontal, wedge-aware wrapping. Arbitrary straight paragraph angles are not supported. |
| `facing`           | Inside a sector, curved-text facing: `auto`, `upright`, or `upside-down`. |
| `bullet`           | Reference to one or more named `<bullet>` styles. Multiple names are whitespace-separated and each reserves its configured width before the paragraph text. |
| `width`, `height`  | Explicit dimensions. |
| `margin`, `margin-top`, `margin-right`, `margin-bottom`, `margin-left` | Outer spacing around the element. |
| `padding`, `padding-top`, `padding-right`, `padding-bottom`, `padding-left` | Inner spacing inside the element box. |
| `border`           | Reference to a named `<pen>` style. |
| `fill`             | Reference to a named `<brush>` style. |
| `rotate`           | Rotate the widget around its origin by the given degrees. |
| `origin-x`         | Rotation origin on the x axis: `start`, `center`, `end`, or a measurement. |
| `origin-y`         | Rotation origin on the y axis: `top`, `middle`, `bottom`, or a measurement. |
| `shift-x`          | Offset the widget horizontally after layout. Measurements use normal LTML units; percentages use the widget's resolved width. |
| `shift-y`          | Offset the widget vertically after layout. Measurements use normal LTML units; percentages use the widget's resolved height. |
| `align`            | Position within parent vbox: `top` (header), `bottom` (footer). |
| `display`          | Retry/visibility policy for repeated page rendering: `once` (default), `always`, `first`, `succeeding`, `even`, `odd`, `last`, or `none` to remove the widget from layout and rendering. |
| `overflow`         | `true` or `false`. Whether a direct page-child paragraph may continue across physical pages. Defaults to `true`. |
| `orphans`          | Minimum number of lines kept on the first continuation fragment. Defaults to `2`. |
| `widows`           | Minimum number of lines carried to the next continuation fragment. Defaults to `2`. |
| `colspan`, `rowspan` | Span multiple table cells (when inside a `table`). |
| `role` | Override the generated PDF structure type when `ua="true"`. Default tagged output uses `P`. |

When `ua="true"`, Arabic paragraphs emit logical-order `/ActualText` from their
fully resolved plain text. Other scripts rely on rendered text and `/ToUnicode`
so viewers can retain granular selection. Replacement text includes inline
links and `<pageno>` output, but not decorative bullet chrome or non-text
decoration.

---

### `<span>` — Inline Text Run

Applies font styling to a portion of text within a `<p>`. Must be a child of
`<p>`, `<label>`, or another `<span>`.

```xml
<p>Normal <span font.weight="Bold">bold</span> normal.</p>
```

Supports the same `font.*` attributes as `<p>`.

`<span>` does not define its own LTML accessibility attributes. In tagged
output, spans contribute text to the enclosing paragraph or label. Inline links
still emit `Link` structure elements. When Arabic replacement text is needed,
it remains on the enclosing paragraph or label rather than on the link itself.

---

### `<leader>` — Inline Leader Fill

Fills the remaining space between the text before it and the text after it
inside a paragraph or label. Leaders are useful for table-of-contents rows,
menus, forms, and other line items with a left label and right-side value.

```xml
<p width="100%">Introduction<leader />1</p>
<label width="3in">Subtotal<leader text="-" />$42.00</label>
```

| Attribute | Description |
|-----------|-------------|
| `text` | Pattern to repeat across the available space. Defaults to `.`. |
| `dot-spacing` | Preferred space between dots for the default `.` leader. Accepts LTML measurements and defaults to the current font's space width. LTML clamps the value to a usable range and may nudge the final character spacing so the leader still reaches the trailing text exactly. |
| `font` / `font.*` | Same inline font attributes supported by `<span>`. |

`<leader>` is valid only inside inline text containers: `<p>`, `<label>`, and
nested `<span>` content. A paragraph or label can contain at most one leader.
For wrapped paragraphs, LTML places the leader on the final rendered line after
wrapping the text before the leader. Labels remain single-line.

The default dot leader adds breathing room before and after the dots and uses
character spacing to fill the exact available gap. The breathing room is based
on the current leader font's `N` width, while default dot density is based on
the current font's space width unless `dot-spacing` is set. Custom `text`
patterns are repeated as whole units and may leave a small remainder if the
pattern width does not divide the available gap exactly.

---

### `<div>` — Container

A generic container for grouping and laying out child elements.

```xml
<div layout="hbox" padding="10pt">
  <p width="50%">Left column</p>
  <p width="50%">Right column</p>
</div>
```

Supports the same layout and styling attributes as `<p>`, plus:

| Attribute        | Description |
|------------------|-------------|
| `layout`         | Layout manager name (see [Layout Managers](#layout-managers)). Use `layout.*` for inline overrides. |
| `dir`            | Layout direction: `ltr` (default) or `rtl`. Inherited from parent container when not set. Reverses horizontal placement in `flow`, `vbox`, `hbox`, and `table` layouts. Invalid values fall back to `ltr`. |
| `cols`           | Number of columns. Required for row-major `table` layout unless `rows` is used instead. Optional for `radial` and `radial-out` when `angles` determines the angular slots. |
| `rows`           | Number of rows. Required for column-major `table` layout unless `cols` is used instead. In `radial`, rows are concentric tracks from outermost to innermost. In `radial-out`, row `0` is innermost and higher rows move outward. |
| `order`          | Grid fill order: `rows` (default) or `cols`. Used by `table`, `radial`, and `radial-out`. |
| `overflow`       | `true` or `false`. Whether a direct page-child `table` or `vbox` may continue across physical pages. Defaults to `true` for `layout="table"` and `layout="vbox"`. |
| `header-rows`    | Number of leading table rows that repeat on every fragment page. Defaults to `0`. |
| `footer-rows`    | Number of trailing table rows that repeat on every fragment page. Defaults to `0`. |
| `base-angle`     | Base angle in degrees for radial sector boundaries. Default: `0`. |
| `row-angle-offsets` | Optional comma-separated angular offsets added to `base-angle` by logical radial row. Missing row values default to `0`. |
| `angles`         | Comma-separated angular boundary bearings relative to `base-angle`. LTML normalizes, sorts, and deduplicates them before building sectors. |
| `sweep`          | Radial sector sweep direction: `ccw` (default) or `cw`. This changes how sectors span between boundaries without changing what the angle numbers mean. |
| `center-x`, `center-y` | Optional radial center coordinates in the container's content box. |
| `r`              | Optional outer radius for radial layout. Otherwise LTML infers it from the smaller content dimension. |
| `r0`             | Optional inner radius for radial layout. Preferred alias when paired with `r`. |
| `paragraph-style` | Default paragraph style for child `<p>` elements. |
| `bullets`        | Optional whitespace-separated list of bullet styles applied to direct child paragraphs in `layout="vbox"` containers. The predefined `unordered` style renders the default circle; the predefined `ordered` style renders the child paragraph's ordinal number. |
| `role` | Override the computed PDF structure type when `ua="true"`, for example `L` or `Table`. |

---

### `<sector>` — Radial Cell

`<sector>` is a container used inside a `layout="radial"` or `layout="radial-out"` parent. It behaves
like a radial table cell: LTML assigns it one wedge or annular-sector region
based on the parent grid, and the sector owns the special paint and layout
behavior for that region.

```xml
<div layout="radial" rows="2" cols="6">
  <sector colspan="2" fill="AliceBlue" border="solid">
    <label>Curved arc text</label>
  </sector>
  <sector rowspan="2" colspan="2">
    <p>Paragraphs curve along successive concentric arcs.</p>
    <p angle="0" position="relative">Horizontal wedge-aware overlay.</p>
  </sector>
</div>
```

Direct non-`<sector>` children of a radial container are wrapped in an
implicit sector automatically. Cell attributes (`colspan`, `rowspan`, `fill`,
borders, `padding`, `display`, and `z-index`) belong to that wrapper; ordinary widget
attributes remain on the source child. `units` applies to both. The transparent
wrapper does not copy the child's identity, classes, role, or alternative text.
For example, a `border` on a direct radial label borders its sector. To border
both, write an explicit sector containing a separately bordered label.

| Attribute | Description |
|-----------|-------------|
| `colspan`, `rowspan` | Span multiple radial slots, just like table cells. |
| `origin-x` | For positioned child widgets inside a sector, `start`, `center`, and `end` anchor to the sector start angle, midpoint angle, and end angle. |
| `origin-y` | For positioned child widgets inside a sector, `inner`, `middle`, and `outer` anchor to the inner radius, midpoint radius, and outer radius. |
| `border-outer`, `border-inner` | Physical aliases for the sector's `border-top` outer arc and `border-bottom` inner arc. |
| `border-start`, `border-end` | Physical aliases for the radial edges at the sector's start and end angles. These map to `border-left` and `border-right` according to the parent's sweep direction. |
| `padding`, `padding-top`, `padding-right`, `padding-bottom`, `padding-left` | Insets sector content. Top/bottom mean outer/inner; left/right mean start/end according to sweep. Start/end padding is a constant physical distance from each radial edge. |
| `layout.hpadding`, `layout.vpadding` | Horizontal item and vertical row gaps for the sector's shape-aware flow. |

An aggregate `border` supplies all four edges. Edge declarations override it:
physical sector aliases take precedence over mapped logical sides, which take
precedence over the aggregate border. `border-top`/`border-outer` paints the outer arc,
`border-bottom`/`border-inner` paints the inner arc, and the left/right borders
paint the two radial edges. For `sweep="ccw"`, left is start and right is end;
for `sweep="cw"`, right is start and left is end. A full-circle sector can
therefore use only `border-outer` and `border-inner` to avoid a radial seam.
The exact lowercase value `none` disables an aggregate or individual edge;
for example, `border="solid" border-inner="none"` leaves the other sector
edges intact.

An explicit sector contains widgets, not text. Nonblank text and inline-only
elements directly inside `<sector>` are errors; wrap them in `<label>`.
Sector attributes are not label defaults. Font, direction, and units retain
their normal container inheritance, while label angle, facing, alignment,
origins, and placement must be authored or styled on the label.

Static sector children participate in a compact shape-aware flow. LTML
preserves source order, chooses row breaks to fit the padded wedge, centers the
packed group, and honors `dir`. Paragraphs consume a full flow band. Curved
paragraphs and labels use arc-length footprints; straight labels and other
widgets use their ordinary boxes. A sector always uses this
special flow for static children even if another `layout` manager is parsed;
wrap children in a nested container to request a vbox, hbox, table, or ordinary
flow.

Set `position="relative"` or `position="absolute"` to remove a child from
sector flow. Existing positional side attributes still imply relative
positioning. Positioned children use sector references; omitted label origins
default to midpoint angle and radius. `origin-x="start|center|end"` pairs with
left/center/right glyph alignment unless `text-align` overrides it.

Labels without an effective `angle` follow their resolved circular arc with
automatic readable orientation. Any effective angle, including `angle="0"`,
renders the label as straight text at that absolute page angle. Straight
sector labels retain ordinary rectangular label boxes. Curved labels are text
overlays: width/height constraints, margin, padding, fill, border, text-fill,
and full-widget `rotate` remain parsed but are ignored while the label is
curved. They become active if an `angle` makes the label straight.
`fit="shrink"` fits curved text to the available sector arc down to the normal
6pt floor.

Paragraphs curve along concentric arcs by default. Each line wraps to the
available padded arc at its own radius; alignment, direction, and one resolved
facing apply consistently across the block. `style.text-align="justify"`
expands every line except the last. `angle="0"` selects horizontal
mode, where LTML computes each line width from the page-axis wedge chord.

All static paragraphs in one sector must use the same mode. To combine a
curved paragraph and an `angle="0"` paragraph, position one of them or place it
in another sector or nested container. Positioned curved paragraphs default
`origin-y` to `middle`; when `origin-x` is omitted it follows effective
start/center/end text alignment. Sector angle, facing, alignment, and origins
are never defaults for paragraph children.

In curved mode, paragraph dimensions, margins, padding, fill, borders,
`text-fill`, and generic `rotate` remain parsed but dormant. Bullets and
indentation are not drawn, leaders do not fill, and linked text does not create
a clickable annotation. Text, spans, font/color changes, shaping, dynamic
inline text, accessibility text, positioning, and offsets remain active. The
ordinary paragraph box and inline features remain available in horizontal
`angle="0"` mode.

---

### `<rect>` — Rectangle

Draws a rectangle with optional border and fill.

```xml
<rect width="100%" height="0.5in" border="solid" fill="LightYellow" corners="0.25" />
```

| Attribute  | Description |
|------------|-------------|
| `width`, `height` | Dimensions of the rectangle. |
| `border`   | Reference to a named `<pen>` style. |
| `fill`     | Reference to a named `<brush>` style. |
| `corners`  | Corner radius for rounded corners, in current units. Accepts 1, 2, 4, or 8 space-separated values; percentages such as `50%` resolve against the smaller box dimension. See corner value order below. |
| `margin`, `margin-top`, `margin-right`, `margin-bottom`, `margin-left` | Outer spacing around the rectangle widget. |
| `padding`, `padding-top`, `padding-right`, `padding-bottom`, `padding-left` | Inner spacing inside the rectangle widget. |

Corner values are interpreted as:

| Count | Meaning |
|-------|---------|
| 1 | Same x/y radius for all four corners. |
| 2 | First value applies to both top corners; second value applies to both bottom corners. |
| 4 | Top-left, top-right, bottom-right, bottom-left, each with the same x/y radius. |
| 8 | Top-left x, top-left y, top-right x, top-right y, bottom-right x, bottom-right y, bottom-left x, bottom-left y. |

For example, `corners="50%"` makes pill-shaped ends for a short rounded
rectangle, while `corners="8 8 0 0"` rounds only the top corners.

---

### Shape Widgets — `<circle>`, `<ellipse>`, `<polygon>`, `<star>`, `<arc>`, `<pie>`, `<arch>`

Draw shape primitives using the existing PDF shape APIs. These widgets are also
containers, so nested LTML content can still be laid out inside their content
boxes.

```xml
<circle width="1.5in" height="1.5in" border="solid" fill="LightBlue" />
<polygon width="1.5in" height="1.5in" sides="6" rotation="30" border="solid" fill="Gold" />
<pie width="1.5in" height="1.5in" start-angle="30" end-angle="150" border="solid" fill="Pink" />
```

Shared attributes:

| Attribute | Description |
|-----------|-------------|
| `width`, `height` | Layout box dimensions. Shapes are centered within their content boxes. |
| `border` | Reference to a named `<pen>` style for the shape outline. |
| `fill` | Reference to a named `<brush>` style for the interior. |
| `margin`, `margin-top`, `margin-right`, `margin-bottom`, `margin-left` | Outer spacing around the shape widget. |
| `padding`, `padding-top`, `padding-right`, `padding-bottom`, `padding-left` | Inner spacing inside the shape widget. |
| `reverse` | Reverse the path direction for shapes that support it. |
| `alt` | When `ua="true"`, opt the shape into tagged output and use this text as `/ActualText`. |
| `role` | Override the default tagged role when `ua="true"`. Shapes with `alt` default to `Figure`. |

Shape-specific attributes:

| Tag | Attributes |
|-----|------------|
| `<circle>` | `r` for an explicit radius. Otherwise radius is inferred from the content box. |
| `<ellipse>` | `rx`, `ry` for explicit radii. Otherwise radii are inferred from width and height. |
| `<polygon>` | `r`, `sides`, `rotation`. |
| `<star>` | `r` or `r1` for outer radius, `r2` for inner radius, `points`, `rotation`. |
| `<arc>` | `r`, `start-angle`, `end-angle`. |
| `<pie>` | `r`, `start-angle`, `end-angle`. |
| `<arch>` | `r1`, `r2`, `start-angle`, `end-angle`. |

---

### `<label>` — Simple Text Label

Draws a single text run using the current font and inline `<span>` styling.
Unlike `<p>`, it does not perform paragraph wrapping or bullet layout.

```xml
<label>Hello <span font.weight="Bold">world</span></label>
```

| Attribute | Description |
|-----------|-------------|
| `font` / `font.*` | Same font attributes supported by `<p>`. |
| `text-fill`, `text-fill.*` | Optional text brush. When present, LTML clips the label text against the widget box fill brush instead of using flat `font.color`. Use the same brush attributes supported by `fill` and `fill.*`. |
| `text-align` | Label text alignment: logical `start`/`end`, physical `left`/`right`, or `center`. Defaults to `start`. Affects the text anchor inside the label box. |
| `text-valign` | Label vertical text alignment: `top` (default), `middle`, or `bottom`. |
| `angle` | Rotate only the label text by the given degrees. Border/fill/background stay axis-aligned. |
| `facing` | Inside a sector, curved-text facing: `auto`, `upright`, or `upside-down`. |
| `fit="shrink"` | If `width` is set and the text is too wide, shrink the label text proportionally until it fits, down to a minimum of 6pt. |
| `width`, `height` | Optional explicit dimensions. |
| `margin`, `margin-top`, `margin-right`, `margin-bottom`, `margin-left` | Outer spacing around the label. |
| `padding`, `padding-top`, `padding-right`, `padding-bottom`, `padding-left` | Inner spacing inside the label box. |
| `border` | Reference to a named `<pen>` style. |
| `fill` | Reference to a named `<brush>` style. |
| `role` | Override the generated PDF structure type when `ua="true"`. Default tagged output uses `P`. |

The built-in `<br/>` alias expands to an empty `<label/>`, which behaves as a
line break in stacked layouts. Labels still do not wrap; `fit="shrink"` scales
the rendered text instead. Unlike the generic widget `rotate` attribute, label
`angle` rotates only the text paint operation.

When `ua="true"`, Arabic labels emit logical-order `/ActualText` from their
fully resolved plain text, including inline links and `<pageno>` output. Other
labels rely on rendered text and `/ToUnicode` for granular selection.

---

### `<pgbr>` — Overflow Page Break

`<pgbr/>` ends the current fragment of its direct overflow-enabled parent and
continues following content on the next physical page. It is a zero-footprint,
non-painting widget, so it does not depend on the amount of space remaining or
add layout padding.

```xml
<page layout="vbox" overflow="true">
  <p>First physical page</p>
  <pgbr/>
  <p>Next physical page</p>
</page>
```

The parent must be a continuing `vbox`, a `flow` page, or a row-major `table`.
Nested markers apply only to the container that directly owns them and do not
bubble through wrappers such as `hbox`. In a table, place `<pgbr/>` between
complete body rows; column-major tables, mid-row placement, and boundaries
crossed by a rowspan are invalid.

A marker at the start or end of a fragment is consumed without creating a
blank page. Consecutive markers collapse in the same way. When the direct
parent has no effective continuation path, including `overflow="false"`, the
marker is inert.

---

### Curved Text Roadmap

Curved text is available for labels inside `<sector>` elements. LTML still does
not have a general-purpose curved-text widget for arbitrary circular or
path-based text outside radial sectors.

The current direction is to add a dedicated curved-text element, likely
`<textpath>`, after the PDF API for circle and ellipse text has stabilized.

The planned LTML concepts mirror the PDF roadmap:

| Concept | Planned Values |
|---------|----------------|
| curve kind | `circle`, `ellipse`, later `path` |
| horizontal anchor | `left`, `center`, `right` |
| vertical anchor | `top`, `above`, `middle`, `baseline`, `below` |
| orientation | geometric or reader-friendly/upright |
| geometry | circle radius, ellipse `rx`/`ry`, later path control points |

SVG `textPath` may inform terminology, but LTML is not currently targeting full
SVG text-on-path compatibility.

---

### `<pre>` — Preformatted Text

Draws literal multiline text without paragraph reflow. This is intended for
code blocks and other preformatted content.

```xml
<pre font="fixed" border="solid" padding="6pt">
  if x &lt; 1 {
    return
  }
</pre>
```

| Attribute | Description |
|-----------|-------------|
| `src` | Optional path or URL to external preformatted text content. When both `src` and inline content are present, `src` wins. |
| `font` / `font.*` | Same font attributes supported by `<p>`. Defaults to the built-in `fixed` font style. |
| `width`, `height` | Optional explicit dimensions. |
| `margin`, `margin-top`, `margin-right`, `margin-bottom`, `margin-left` | Outer spacing around the block. |
| `padding`, `padding-top`, `padding-right`, `padding-bottom`, `padding-left` | Inner spacing inside the block. |
| `border` | Reference to a named `<pen>` style. |
| `fill` | Reference to a named `<brush>` style. |
| `role` | Override the generated PDF structure type when `ua="true"`. `<pre>` has no default tagged role. |

`<pre>` preserves internal spaces and line breaks, does not wrap lines, trims a
single surrounding newline from block content, removes common leading
indentation from non-blank lines, and expands tab characters to four spaces.
External `src` content is loaded lazily on first use using the same
source-resolution rules as component-backed tags such as `<svg>`:

- when an asset filesystem is attached, `src` must be a clean relative asset
  path and is read virtually from that asset filesystem
- when no asset filesystem is attached, relative file paths are resolved
  relative to the LTML document being parsed, while absolute paths remain
  absolute
- `http` and `https` URLs are supported only when document/network loading is
  explicitly enabled; network assets are fetched lazily into a temp file and
  cleaned up after rendering

When `<pre>` participates in tagged output through `role`, Arabic content uses
the same resolved preformatted text as logical-order `/ActualText`; other
scripts rely on rendered text and `/ToUnicode`.

---

### `<image>` — Image Placement

Places an image file into the document using the PDF image API.

```xml
<image src="../pdf/testdata/testimg.jpg" width="2in" />
```

| Attribute | Description |
|-----------|-------------|
| `src` | Path to the source image file. |
| `width`, `height` | Optional explicit dimensions. If only one is supplied, the other is inferred from the image aspect ratio. |
| `max-width`, `max-height` | Optional maximum widget dimensions. Caps preserve aspect ratio and choose whichever dimension dominates. |
| `margin`, `margin-top`, `margin-right`, `margin-bottom`, `margin-left` | Outer spacing around the widget box. |
| `padding`, `padding-top`, `padding-right`, `padding-bottom`, `padding-left` | Inner spacing inside the widget box. |
| `border` | Reference to a named `<pen>` style. |
| `fill` | Reference to a named `<brush>` style. |
| `alt` | When `ua="true"`, opt the image into tagged output and use this text as `/ActualText`. |
| `role` | Override the default tagged role when `ua="true"`. Images with `alt` default to `Figure`. |

The current implementation supports JPEG, PNG, and SVG files through the same
tag shape. Raster images are embedded as PDF image XObjects. SVG files are
parsed and rendered through the PDF vector drawing path so unsupported SVG
features are skipped with warnings to standard error rather than aborting the
whole document when the rest of the file can still render.

`<image>` loads its source lazily on first layout or render use, so its `src`
resolution is consistent with other `src`-loading tags such as `<svg>`:

- when an asset filesystem is attached, `src` must be a clean relative asset
  path such as `logo.png` or `assets/logo.png`; LTML keeps that path virtual
  and lets the writer read it through the configured asset filesystem
- when no asset filesystem is attached, relative file paths are resolved
  relative to the LTML document being parsed, while absolute paths remain
  absolute
- `http` and `https` URLs are supported only when document/network loading is
  explicitly enabled; network assets are fetched lazily into a temp file and
  cleaned up after rendering

Images without `alt` remain decorative and are not added to the document's
logical structure tree.

---

### `<svg>` — SVG Placement

Places SVG content through the dedicated SVG rendering path. The built-in
`<svg>` tag is a component-backed element, so `src` uses the same component
body-loading behavior as other components, including document asset FS and
optional network loading.

```xml
<svg src="logo.svg" width="2in" />
```

```xml
<svg width="2in">
  <svg xmlns="http://www.w3.org/2000/svg" width="120" height="60" viewBox="0 0 120 60">
    <rect width="120" height="60" fill="#ffffff"/>
    <circle cx="30" cy="30" r="18" fill="#88bbff"/>
  </svg>
</svg>
```

| Attribute | Description |
|-----------|-------------|
| `src` | Optional path or URL to an external SVG document. When both `src` and inline SVG body are present, `src` wins. |
| `style` | Optional CSS text injected as a `<style>` element inside the SVG root before rendering. Applies to both inline SVG bodies and external `src` SVGs. |
| `width`, `height` | Optional explicit dimensions. If only one is supplied, the other is inferred from the SVG aspect ratio. |
| `max-width`, `max-height` | Optional maximum widget dimensions. Caps preserve aspect ratio and choose whichever dimension dominates. |
| `margin`, `margin-top`, `margin-right`, `margin-bottom`, `margin-left` | Outer spacing around the widget box. |
| `padding`, `padding-top`, `padding-right`, `padding-bottom`, `padding-left` | Inner spacing inside the widget box. |
| `border` | Reference to a named `<pen>` style. |
| `fill` | Reference to a named `<brush>` style for the LTML widget box, not the SVG document root. |
| `alt` | When `ua="true"`, opt the SVG into tagged output and use this text as `/ActualText`. |
| `role` | Override the default tagged role when `ua="true"`. SVGs with `alt` default to `Figure`. |

LTML attributes affect page placement and widget styling only; they are not
forwarded into the SVG XML. Inline SVG bodies should contain a full nested SVG
document. External `src` content is loaded lazily into the component body on
first use, so local assets and network SVGs follow the same render path as
inline markup.

The `style` attribute is the exception: its value is injected into the SVG
document as a `<style><![CDATA[...]]></style>` child of the SVG root. This is
useful when reusing the same source SVG with different class-based colors or
stroke settings:

```xml
<svg src="badge.svg" width="1in"
     style=".accent { fill: #f6d44e; stroke: #222222; }" />
```

Because `<svg>` is component-backed, its `src` is loaded into the component
body lazily on first use, but its path and URL resolution rules match
`<image>`:

- when an asset filesystem is attached, `src` must be a clean relative asset
  path and is read virtually from that asset filesystem
- when no asset filesystem is attached, relative file paths are resolved
  relative to the LTML document being parsed, while absolute paths remain
  absolute
- `http` and `https` URLs are supported only when document/network loading is
  explicitly enabled; network assets are fetched lazily into a temp file and
  cleaned up after rendering

---

### `<line>` — Line Segment

Draws a line segment using the configured pen style.

```xml
<line width="100%" height="12pt" style="dashed" />
```

| Attribute | Description |
|-----------|-------------|
| `style` | Reference to a named `<pen>` style. Use `style.*` for inline overrides such as `style.width="2pt"` or `style.color="red"`. |
| `angle` | Line angle in degrees. `0` points right; positive angles rotate clockwise in page coordinates. |
| `length` | Optional explicit line length. |
| `width`, `height` | Optional layout box dimensions used to infer the line length and placement when `length` is omitted. |
| `margin`, `margin-top`, `margin-right`, `margin-bottom`, `margin-left` | Outer spacing around the widget box. |
| `padding`, `padding-top`, `padding-right`, `padding-bottom`, `padding-left` | Inner spacing inside the widget box. |
| `border` | Optional enclosing widget border, separate from the line stroke. Use `border.*` to override the enclosing border pen inline. |
| `alt` | When `ua="true"`, opt the line into tagged output and use this text as `/ActualText`. |
| `role` | Override the default tagged role when `ua="true"`. Lines with `alt` default to `Figure`. |

Horizontal, vertical, and diagonal lines are all represented with the same tag.
When `length` is omitted, the line is sized to fit within its content box.

---

### `<pageno>` — Inline Page Number

Renders the current document page number inline anywhere `<span>` is allowed.
It also supports page-counter control attributes for the current or following
rendered PDF page.

```xml
<p>Page <pageno /></p>
<p><pageno hidden="true" reset="1" /></p>
<p>Page <pageno start="10" font.weight="Bold" /></p>
```

| Attribute | Description |
|-----------|-------------|
| `font` / `font.*` | Same inline font attributes supported by `<span>`. |
| `start` | Set the visible number for the current rendered PDF page. |
| `reset` | Set the number that should begin on the next rendered PDF page. |
| `hidden` | If `true`, apply control semantics without rendering visible text. |

`<pageno>` is valid only where `<span>` is valid today: inside `<p>`,
`<label>`, and nested `<span>`. The current implementation renders decimal page
numbers only.

---

### `<index>` — Generated Index Or Table Of Contents

Renders collected `<index-entry>` rows after document layout resolves target
page numbers. The common table-of-contents form uses inline placeholders inside
one row template:

```xml
<index id="main_toc">
  <p width="100%"><index-title /><index-leader /><index-page /></p>
</index>

<label id="intro">Introduction</label>
<index-entry index="main_toc" target="intro">Introduction</index-entry>
```

`<index>` accepts zero or one block template child. The template may be a
`<p>`, `<label>`, or container. If omitted, LTML uses a full-width paragraph
template equivalent to:

```xml
<p width="100%"><index-title /><leader /><index-page /></p>
```

| Attribute | Description |
|-----------|-------------|
| `id` | Index identifier matched by each `<index-entry index="...">`. |
| `layout` / `layout.*` | Layout manager and inline layout overrides for the generated rows. |
| `font` / `font.*` | Default font style for generated row content. |
| `width`, `height`, margins, padding, border, fill, positioning attrs | Standard container/widget attributes. |

Generated index rows link to their target destination when rendered through an
LTML PDF writer that supports internal target links.

Inline placeholders for index row templates:

| Element | Description |
|---------|-------------|
| `<index-title />` | Entry label text from the matching `<index-entry>`. |
| `<index-page />` | Resolved destination page number. |
| `<index-leader />` | Built-in alias for `<leader />`, normally used between title and page. |

`<index-title>` and `<index-page>` support the same inline `font.*` attributes
as `<span>`. `<index-leader>` supports the same attributes as `<leader>`,
including `text`.

### `<index-entry>` — Hidden Index Entry

Contributes one entry to a named `<index>` without rendering visible content.

```xml
<index-entry index="main_toc" target="intro">Introduction</index-entry>
```

| Attribute | Description |
|-----------|-------------|
| `index` | Target index id. |
| `target` | Destination id to resolve and link to. |

The element body is the entry label text. Entries render in encounter order.
Any printed page or widget with an `id` can be a destination, and `<target
id="..."/>` can define an explicit zero-footprint destination.

`<index-entry>` does not require a matching `<index>` in the same document.
Entries whose index id is not rendered by any `<index>` widget are collected
during layout but produce no visible output. This lets shared page templates
carry TOC metadata without also requiring a table-of-contents page on every
render.

---

## Style Definitions

Style definitions are placed inside `<ltml>`, `<page>`, or `<canvas>` before
ordinary content that uses them. A scope owner's own attributes may reference
definitions placed at the beginning of that scope, as described under
[Scope](#scope).

### `<font>` — Font Style

```xml
<font id="title" name="Helvetica" size="24" color="navy" weight="Bold" />
```

```xml
<pen id="accent-line" color="tomato" width="1.5pt" pattern="dashed" cap="round_cap" />
<font
  id="accented"
  name="Helvetica"
  size="14"
  underline="true"
  font.underline-pen="accent-line"
  font.underline-pos="-1pt"
  strikeout="true"
  font.strikeout-pos="0.08in" />
```

| Attribute     | Description |
|---------------|-------------|
| `id`          | Name used to reference this style. |
| `name`        | Font family name. |
| `size`        | Font size as points (`12`) or page-root-relative rems (`1rem`, `0.875rem`). |
| `color`       | Text color. |
| `weight`      | `Bold`, or omit for normal. |
| `style`       | `Italic`, `Oblique`, or omit for normal. |
| `underline`   | `true` or `false`. |
| `strikeout`   | `true` or `false`. |
| `underline-pen` | Reference to a named `<pen>` style used for underline stroke settings. |
| `strikeout-pen` | Reference to a named `<pen>` style used for strikeout stroke settings. |
| `underline-pos` | Underline position as an LTML measurement. |
| `strikeout-pos` | Strikeout position as an LTML measurement. |
| `stroke-color` | Optional text stroke color for filled-and-stroked visible text. |
| `stroke-width` | Optional text stroke width as an LTML measurement (commonly in `pt`). |
| `line-height` | Line spacing multiplier. |

**Default font:** Helvetica 12pt.

**Built-in font style:** `fixed` — Courier New 12pt.

---

### `<pen>` — Pen Style (for borders and lines)

```xml
<pen id="rule" color="black" width="1pt" pattern="solid" />
<pen id="pill-border" kind="linear-gradient" width="5pt"
     x0="0%" y0="50%" x1="100%" y1="50%"
     stops="0:#ef5148,1:#4f93ad" cap="round_cap" />
```

| Attribute  | Description |
|------------|-------------|
| `id`       | Name used to reference this style. |
| `kind`     | Pen type: `solid` (default), `linear-gradient`, or `radial-gradient`. |
| `color`    | Line color for solid pens. |
| `width`    | Line width (with optional unit suffix, e.g. `2pt`, `0.5mm`). |
| `pattern`  | Line pattern: `solid`, `dashed`, `dotted`. |
| `cap`      | Line cap: `butt_cap`, `round_cap`, or `projecting_square_cap`. |
| `stops`    | Comma-separated gradient stops like `0:#112233,0.5:Gold,1:#445566`. |
| `x0`, `y0`, `x1`, `y1` | Gradient coordinates. Accept LTML measurements, and also accept percentages like `50%` resolved against the stroked box width or height. Used by linear gradients and the center points of radial gradients. |
| `r0`, `r1` | Radial gradient start and end radii. Accept LTML measurements, and also accept percentages resolved against the stroked box's smaller dimension. |

**Built-in pen styles:** `solid`, `dashed`, `dotted` — all black, hairline width.

You may also use a color name directly as a border value to create an
auto-generated solid pen for that color:

```xml
<rect border="red" />
```

The exact lowercase value `none` disables a standard widget border. Surrounding
whitespace is ignored, but differently cased values are ordinary pen names.
Individual sides override the aggregate border, so this draws three connected
edges and suppresses the top edge:

```xml
<label border="solid" border-top="none">Three-sided box</label>
```

Rounded `corners` remain active when individual sides are specified. Contiguous
edges with the same effective pen are drawn as one stroke, preserving their
corner joins. Adjacent edges with different pens meet at the midpoint of their
shared corner; the pens' endcap styles control that transition. An explicitly
disabled border remains disabled when only subattributes such as
`border.color` or `border-top.width` are applied later. Supply an explicit pen
value such as `border="solid"` or `border-top="dashed"` to re-enable that
property. Side declarations remain independent, so `border="none"
border-bottom="solid"` draws only the bottom edge.

---

### `<brush>` — Brush Style (for fill colors)

```xml
<brush id="highlight" color="yellow" />
```

| Attribute | Description |
|-----------|-------------|
| `id`      | Name used to reference this style. |
| `kind`    | Brush type: `solid` (default), `linear-gradient`, `radial-gradient`, `sweep-gradient`, or `image`. |
| `color`   | Fill color for solid brushes. |
| `stops`   | Comma-separated gradient stops like `0:#112233,0.5:Gold,1:#445566`. |
| `x0`, `y0`, `x1`, `y1` | Gradient coordinates. Accept LTML measurements, and also accept percentages like `50%` resolved against the painted box width or height. Used by linear gradients and the center points of radial gradients. |
| `r0`, `r1` | Radial gradient start and end radii. Accept LTML measurements, and also accept percentages resolved against the painted box's smaller dimension. |
| `steps` | Number of equal-angle subdivisions used for each adjacent stop interval of a `sweep-gradient`. Values less than or equal to zero default to `1`. |
| `src` | Image source for `kind="image"`. Supports the same asset resolution rules as `<image>`. |
| `fit` | Image brush sizing mode: `stretch`, `contain`, `cover`, or `tile`. |
| `anchor` | Image brush alignment inside the painted box: `center`, `top`, `bottom`, `left`, `right`, `top-left`, `top-right`, `bottom-left`, `bottom-right`. |
| `repeat` | Image repetition mode: `no-repeat` (default), `repeat`, `repeat-x`, or `repeat-y`. |
| `opacity` | Uniform opacity for gradient and image brushes. Accepts `0` to `1` values or percentages like `60%`. Default: `1`. |
| `tile-width`, `tile-height` | Explicit rendered tile size for `fit="tile"`. Accept LTML measurements, and also accept percentages like `50%` or `100%` resolved against the painted box. If only one side is specified, LTML preserves the source aspect ratio. |

`sweep-gradient` brushes are currently supported only as sector fills. The
sector supplies the center, inner and outer radii, and angular span; gradient
stops map from the sector's start angle (`0`) to its end angle (`1`), including
clockwise spans. PDF has no native sweep shading, so LTML approximates each
interval with clipped, chord-aligned linear gradients. Increase `steps` for a
smoother curved appearance at the cost of additional PDF shading operations.

The clipped segments overlap very slightly to prevent PDF rasterizers from
showing hairline cracks between them. At present, `opacity` is applied to each
segment rather than once to the completed sweep. A translucent sweep gradient
can therefore show thin, darker radial seams where adjacent segments overlap;
more `steps` produces more such boundaries. Omit `opacity` when those seams are
unacceptable. A future band-wide transparency implementation can remove this
limitation without reintroducing cracks.

```xml
<brush id="rainbow" kind="sweep-gradient" steps="12"
  stops="0:Red,0.25:Blue,0.5:Green,0.75:Gold,1:Red" />

<div layout="radial-out" rows="1" cols="1" r0="0.6in"
  width="2in" height="2in">
  <sector fill="rainbow"></sector>
</div>
```

Image brushes default to the source asset's intrinsic size when `fit="tile"` and
no explicit tile size is provided. That means very large source images may clip
instead of visibly repeating. In print-oriented documents, prefer an explicit
tile size for predictable output, and use percentage tile dimensions like
`tile-height="100%"` when you want high-resolution artwork to scale to the box
before repeating.

```xml
<brush id="metal" kind="image"
  src="../../docs/assets/metal-movable-type-banner.jpg"
  opacity="0.5" />

<vbox fill="metal"
  fill.fit="tile"
  fill.tile-height="100%"
  fill.repeat="repeat-x"
  fill.anchor="top-left">
  <p>This keeps each tile as tall as the widget box.</p>
</vbox>
```

You may also use a color name directly as a fill value:

```xml
<rect fill="LightBlue" />
```

Like other named styles, brushes can be referenced with `fill="brush-id"` and
overridden inline with `fill.*` attributes on a widget.

---

### `<para>` — Paragraph Style

```xml
<para id="body" text-align="justify" valign="top" bullet="bstar" />
```

| Attribute    | Description |
|--------------|-------------|
| `id`         | Name used to reference this style. |
| `text-align` | Logical `start`/`end`, physical `left`/`right`, `center`, or `justify`. Defaults to `start`. |
| `valign`     | `top`, `middle`, `bottom`, `baseline`. |
| `bullet`     | Reference to one or more named `<bullet>` styles. Multiple names are whitespace-separated and each reserves its configured width before the paragraph text. |

---

### `<bullet>` — Bullet Style

```xml
<font id="zapf" name="ZapfDingbats" size="12" />
<bullet id="bstar" font="zapf" text="&#x4E;" width="36" units="pt" />
```

| Attribute | Description |
|-----------|-------------|
| `id`      | Name used to reference this style. |
| `font`    | Reference to a named `<font>` style. |
| `text`    | The bullet character(s) to render for text bullets. |
| `format`  | Optional `fmt.Sprintf`-style integer format used when a container expands this bullet through `bullets`. For example, `format="%d."` renders `1.`, `2.`, etc. |
| `src`     | Asset path for image bullets. Rendered through LTML's normal `PrintImageFile` path. |
| `shape`   | Closed shape for shape bullets: `circle`, `ellipse`, `polygon`, `triangle`, `square`, or `star`. `triangle` and `square` are polygon aliases. |
| `width`   | Space reserved for the bullet slot. This remains the paragraph indent reservation for every bullet kind. |
| `r`       | Optional outer radius for `circle`, `polygon`, `triangle`, `square`, and `star` bullets. LTML uses it as the center-to-corner distance. |
| `rx`      | Optional horizontal radius for `shape="ellipse"`. |
| `ry`      | Optional vertical radius for `shape="ellipse"`. |
| `height`  | Optional minimum paragraph height reserved for the bullet. When present, LTML ensures the paragraph is at least this tall, so a tall bullet may span multiple wrapped text lines. |
| `pen`     | Optional `<pen>` style for shape bullet outlines. |
| `brush`   | Optional `<brush>` style for shape bullet fills, including gradients and image brushes. |
| `align-x` | Horizontal alignment of the painted bullet within the reserved slot: `start`, `center`, `end`. `start` flushes the shape to the slot's leading edge, `center` aligns shape center to slot center, and `end` flushes the shape to the slot's trailing edge. Defaults to `start` (`end` for RTL). |
| `align-y` | Vertical alignment of the painted bullet: `top`, `middle`, `baseline`. `top` flushes the shape to the top of the paragraph bullet slot. `middle` aligns the shape center with the middle of the paragraph's text block. `baseline` aligns the bottom of the painted shape to the first-line text baseline. Defaults to `top`. |
| `sides`   | Polygon side count for `shape="polygon"`. |
| `points`  | Star point count for `shape="star"`. |
| `rotation` | Optional rotation in degrees for ellipse, polygon, and star bullets. |
| `r0`      | Optional inner radius for `shape="star"`. |
| `units`   | Units for `width`, `height`, `r`, `rx`, `ry`, and `r0`. |

Examples:

```xml
<bullet id="dot" font="zapf" text="l" width="18pt" />
<bullet id="logo" src="../../pdf/testdata/test_scene.svg" width="18pt" height="14pt" />
<bullet id="tri" shape="triangle" width="18pt" r="6pt" brush="goldfill" />
<bullet id="oval" shape="ellipse" width="18pt" rx="6pt" ry="4pt" brush="skyfill" />
<bullet id="brand-star" shape="star" width="18pt" r="6pt" brush="goldfill" pen="solid" points="6" r0="4pt" rotation="15" />

Notes:
`triangle` maps to a 3-sided polygon with an upright default orientation.
`square` maps to a 4-sided polygon with an unrotated default orientation.
If `shape="triangle"` or `shape="square"`, LTML ignores any explicit `sides` value.
```

---

### `<layout>` — Layout Style

Pre-defines a named layout configuration that can be referenced by containers.

```xml
<layout id="tight" manager="vbox" padding="4" vpadding="6" />
```

| Attribute  | Description |
|------------|-------------|
| `id`       | Name used to reference this layout. |
| `manager`  | Layout algorithm (see [Layout Managers](#layout-managers)). |
| `padding`  | Space between children (both horizontal and vertical). |
| `hpadding` | Horizontal space between children. |
| `vpadding` | Vertical space between children. |
| `units`    | Units for padding values. |

---

### Page Size and Orientation

**Built-in page sizes** (set via `style` on `<page>`):

- ISO A: `A0` through `A10`
- ISO B: `B0` through `B10`
- ISO C: `C0` through `C10`
- ISO raw: `RA0` through `RA4`, `SRA0` through `SRA4`
- North American: `halfletter`, `statement`, `letter`, `legal`,
  `juniorlegal`, `tabloid`, `ledger`, `governmentletter`,
  `governmentlegal`, `executive`, `folio`, `quarto`
- ANSI: `ansia` through `ansie`
- ARCH: `archa`, `archb`, `archc`, `archd`, `arche`, `arche1`, `arche2`,
  `arche3`

Built-in page-size lookup is case-insensitive, so `A4` and `a4` resolve to the
same built-in size. `letter` remains the default page size.

The page-style override attribute `style.orientation` (`portrait` or
`landscape`) swaps width and height when combined with `style`. Explicit
`width` and `height` values can also be set directly on `<page>` to use a
one-off custom page size.

```xml
<page style="A4" style.orientation="landscape" units="cm" margin="2">
  <!-- content -->
</page>
```

#### Custom named page sizes (`<pagestyle>`)

Use `<pagestyle>` to define a reusable named page size at document scope:

```xml
<pagestyle id="book" units="in" width="6" height="9"/>
<page style="book" margin="0.75in">
  <!-- content -->
</page>
```

For a standalone page, the page may instead reference a page style declared at
the beginning of its own scope:

```xml
<page style="book" margin="0.75in">
  <pagestyle id="book" units="in" width="6" height="9"/>
  <!-- content -->
</page>
```

| Attribute     | Description |
|---------------|-------------|
| `id`          | Name used in the `style` attribute of `<page>`. Required. |
| `width`       | Page width. |
| `height`      | Page height. |
| `units`       | Unit for `width` and `height`: `pt` (default), `in`, `cm`, `mm`, `dp`. |
| `orientation` | `portrait` (default) or `landscape`. |

At document scope, the name is available to all subsequent `<page>` elements.
At page scope, it is isolated to that page and may safely reuse a name from
another page.

#### Avery extension package

Avery label-stock support is provided by the optional
`github.com/rowland/leadtype/avery` package rather than the core `std`
namespace. Import it for registration, then use the `avery` XML namespace:

```go
import _ "github.com/rowland/leadtype/avery"
```

```xml
<ltml xmlns:avery="avery">
  <page style="letter" margin="0">
    <avery:labelsheet stock="5160" show-metrics="true" show-outline="true">
      <avery:label border="thin" padding="6pt">
        <p font.weight="Bold">Jamie Smith</p>
        <p>123 Main St</p>
        <p>Seattle, WA 98101</p>
      </avery:label>
      <avery:label border="thin" padding="6pt">
        <p font.weight="Bold">Morgan Lee</p>
        <p>987 Market Ave</p>
        <p>Portland, OR 97205</p>
      </avery:label>
    </avery:labelsheet>
  </page>
</ltml>
```

The extension currently includes this built-in catalog of Avery-compatible US
Letter label stocks:

| Canonical ID | Accepted Aliases | Layout |
|--------------|------------------|--------|
| `avery5160` | `5160`, `8160` | 3 × 10 labels, 1" × 2-5/8" |
| `avery5161` | `5161`, `8161` | 2 × 10 labels, 1" × 4" |
| `avery5162` | `5162`, `8162` | 2 × 7 labels, 1-1/3" × 4" |
| `avery5163` | `5163`, `8163` | 2 × 5 labels, 2" × 4" |
| `avery5164` | `5164`, `8164` | 2 × 3 labels, 3-1/3" × 4" |
| `avery5167` | `5167`, `8167` | 4 × 20 labels, 1/2" × 1-3/4" |
| `avery5366` | `5366` | 2 × 15 labels, 2/3" × 3-7/16" |
| `avery5395` | `5395` | 2 × 4 badges, 2-1/3" × 3-3/8" |

`<avery:labelsheet>` attributes:

| Attribute | Description |
|-----------|-------------|
| `stock` | Built-in stock id or alias. Required. |
| `order` | Fill order for child labels: `rows` (default) or `cols`. |
| `show-metrics` | When `true`, prints a compact appendix caption with the canonical stock id, label size, and grid count. |
| `show-outline` | When `true`, draws the full stock grid as an overlay, which is useful for samples and calibration. |
| `border`, `font.*`, `font`, positioning attrs | Standard widget styling/positioning attrs still apply. |

`<avery:labelsheet>` behaves like a stock-backed label table:

- Each direct `<avery:label>` child occupies one label slot.
- Slot width, height, row count, column count, and gutter spacing come from the selected stock.
- Children are filled across rows by default, or down columns when `order="cols"`.
- Printing fails with a clear error if the number of child labels exceeds the stock capacity.

`<avery:label>` is the author-facing label cell tag. It behaves like a small
container, so standard widget styling such as `padding`, `border`, `fill`, and
nested content like `<p>` and `<label>` can be used inside each label.

---

## Aliases (`<define>`)

Create a shorthand tag that expands to another tag with preset attributes.

```xml
<define id="td" tag="p" border="solid" padding="3pt" />
```

Now `<td>` is equivalent to `<p border="solid" padding="3pt">`.

| Attribute | Description |
|-----------|-------------|
| `id`      | The new tag name. |
| `tag`     | The underlying tag to expand to. |
| *(any)*   | Default attribute values for the expanded tag. |

### Built-in Aliases

| Alias   | Expands to | Default Attributes |
|---------|------------|--------------------|
| `<h>`   | `<p>`      | `font.weight="Bold"`, `style.text-align="center"`, `width="100%"` |
| `<h1>`  | `<label>`  | `font.weight="Bold"`, `font.size="2rem"`, `role="H1"` |
| `<h2>`  | `<label>`  | `font.weight="Bold"`, `font.size="1.75rem"`, `role="H2"` |
| `<h3>`  | `<label>`  | `font.weight="Bold"`, `font.size="1.5rem"`, `role="H3"` |
| `<h4>`  | `<label>`  | `font.weight="Bold"`, `font.size="1.25rem"`, `role="H4"` |
| `<h5>`  | `<label>`  | `font.weight="Bold"`, `font.size="1.125rem"`, `role="H5"` |
| `<h6>`  | `<label>`  | `font.weight="Bold"`, `font.size="1rem"`, `role="H6"` |
| `<b>`   | `<span>`   | `font.weight="Bold"` |
| `<i>`   | `<span>`   | `font.style="Italic"` |
| `<u>`   | `<span>`   | `font.underline="true"` |
| `<s>`   | `<span>`   | `font.strikeout="true"` |
| `<index-leader>` | `<leader>` | *(empty; accepts the same inline attributes as `<leader>`)* |
| `<hbox>` | `<div>`   | `layout="hbox"` |
| `<vbox>` | `<div>`   | `layout="vbox"` |
| `<ul>` | `<div>` | `layout="vbox"`, `bullets="unordered"` |
| `<ol>` | `<div>` | `layout="vbox"`, `bullets="ordered"` |
| `<table>` | `<div>` | `layout="table"` |
| `<disc>` | `<div>` | `layout="radial"` |
| `<th>`  | `<p>`      | `role="TH"`, `font.weight="Bold"` |
| `<td>`  | `<p>`      | `role="TD"` |
| `<layer>` | `<div>` | `position="relative"`, `width="100%"`, `height="100%"` |
| `<br>`  | `<label>`  | *(empty line break)* |

---

## Style (CSS-like Selectors)

Apply attributes to elements based on tag name, id, and class selectors.

```xml
<style tier="4">
  p { font.size: 14; }
  p.intro { font.weight: Bold; style.text-align: justify; }
  div#footer { margin-top: 20; }
</style>
```

```xml
<style src="styles.ltml.css" />
```

| Attribute | Description |
|-----------|-------------|
| `src` | Optional path or URL to an external LTML stylesheet. When both `src` and inline rules are present, `src` wins. |
| `tier` | Optional cascade tier override. Higher tiers override lower tiers before specificity and source order are considered. |

Style blocks use CSS-style selector syntax:

| Pattern         | Matches |
|-----------------|---------|
| `p`             | All `<p>` elements |
| `.classname`    | Elements with `class="classname"` |
| `p.classname`   | `<p>` elements with `class="classname"` |
| `p#myid`        | `<p>` elements with `id="myid"` |
| `div p`         | `<p>` elements anywhere inside a `<div>` |
| `div > p`       | `<p>` elements that are direct children of a `<div>` |
| `p, span`       | All `<p>` and `<span>` elements |

Selector names accept letters, digits, underscores, and hyphens in tags, ids,
and classes, so selectors such as `my-widget`, `#hero-panel`, and `.demo-card`
are valid.

LTML supports these pseudo-classes:

| Pseudo-class | Matches |
|--------------|---------|
| `:dir(ltr)` | Widgets whose effective inherited layout direction is left-to-right. |
| `:dir(rtl)` | Widgets whose effective inherited layout direction is right-to-left. |
| `:first-child` | The first direct child widget of a container. |
| `:last-child` | The last direct child widget of a container. |
| `:first-row` | Widgets anchored in row `0` of a `layout="table"` container. |
| `:last-row` | Widgets anchored in the last row of a `layout="table"` container. |
| `:first-col` | Widgets anchored in column `0` of a `layout="table"` container. |
| `:last-col` | Widgets anchored in the last column of a `layout="table"` container. |
| `:row-even`, `:row-odd` | Widgets in even/odd zero-based table rows. |
| `:col-even`, `:col-odd` | Widgets in even/odd zero-based table columns. |
| `:row-N` | Widgets anchored in zero-based table row `N`, for example `:row-2`. |
| `:col-N` | Widgets anchored in zero-based table column `N`, for example `:col-1`. |

Notes:

- Direction pseudo-classes use the effective `dir`, including inheritance and
  local container overrides; they do not require a `dir` attribute on the
  matched widget. Only the exact `:dir(ltr)` and `:dir(rtl)` spellings are
  supported.
- Row and column pseudo-classes apply only to direct children of a
  `layout="table"` container.
- Row and column numbering is zero-based.
- For `rowspan` and `colspan`, LTML matches row/column pseudo-classes using the
  widget's anchor cell rather than every covered cell.

CSS-style `/* ... */` comments are ignored inside `<style>`. Rules inside
`<!-- XML comments -->` are also parsed, allowing selectors to be commented out
with nested comment delimiters.

External `src` content is loaded when the `<style>` block is parsed, then kept
in memory on that rule set so later selector matching does not reread the
source asset. File paths and URLs use the same asset-resolution rules as other
text-backed LTML sources:

- when an asset filesystem is attached, `src` must be a clean relative asset
  path and is read virtually from that asset filesystem
- when no asset filesystem is attached, relative file paths are resolved
  relative to the LTML document being parsed, while absolute paths remain
  absolute
- `http` and `https` URLs are supported only when document/network loading is
  explicitly enabled

`<style>` accepts an optional `tier` attribute. Higher tiers override lower
tiers before specificity and source order are considered. Default tiers are:

- document-scope `<ltml><style>`: `0`
- page-scope `<page><style>`: `1`

Attribute priority (lowest to highest):
1. Default attributes from an alias (`<define>`)
2. Attributes from matching rules, ordered by tier, then specificity, then declaration order
3. Direct XML attributes on the element

---

## Layout Managers

Set via the `layout` attribute on any container element or via `<layout id="..." manager="...">`.

| Name       | Description |
|------------|-------------|
| `vbox`     | Stack children vertically (default). |
| `hbox`     | Arrange children side by side horizontally. |
| `table`    | Grid layout. Requires `cols` (for row order) or `rows` (for column order). |
| `flow`     | Wrap children left-to-right, top-to-bottom like inline text. |
| `absolute` | Children are positioned absolutely; no automatic layout. |
| `relative` | Children use relative positioning. |
| `radial`   | Grid layout on concentric tracks and angular sectors. Requires `rows`, `cols`, or explicit `angles` so LTML can derive the radial grid. |
| `radial-out` | Radial grid layout that starts at the center and grows outward. Requires `rows`, `cols`, or explicit `angles` so LTML can derive the radial grid. |

All layout managers except `absolute` and `relative` honor the `dir` attribute.
When `dir="rtl"` is set on a container (or inherited from a parent), horizontal
placement is mirrored so that content flows from the right edge.
This mirrors layout placement. Paragraph and label `text-align` defaults to the
logical `start` edge, which is left in LTR and right in RTL. Explicit `start` and
`end` values follow the effective inherited `dir`; physical `left` and `right`
values are not mirrored.

| `text-align` | `dir="ltr"` | `dir="rtl"` |
|--------------|-------------|-------------|
| omitted / `start` | left | right |
| `end` | right | left |
| `left` | left | left |
| `right` | right | right |
| `center` | center | center |
| `justify` | justify | justify |

Direction-aware alignment does not automatically change paragraph shaping,
glyph ordering, or bidi behavior inside text widgets. Sector `text-align` keeps
its separate angular-anchor semantics.

### VBox Details

- Children stack top to bottom.
- Width defaults to content width of the container.
- Use `align="top"` to pin an element to the top (header behavior).
- Use `align="bottom"` to pin an element to the bottom (footer behavior).
- Use `align-self="start"`, `align-self="center"`, or `align-self="end"` to control horizontal placement within the vbox row.
- In `dir="rtl"`, `align-self="start"` means right and `align-self="end"` means left.
- In `dir="rtl"`, children are flush against the right edge plus padding instead of the left.
- When at least one child uses `height="auto"` and a height-constrained vbox
  fragment has true surplus height beyond the specified, percent, and preferred
  heights of the children on that fragment plus `layout.vpadding`, omitted
  heights keep their preferred heights and the `auto` children split the
  leftover height evenly.
- In natural-height vboxes, and in constrained vboxes without surplus,
  `height="auto"` behaves the same as an omitted height.
- When a vbox splits across pages, `height="auto"` is evaluated separately for
  each fragment page based only on the children present on that fragment.

### HBox Details

- Children are laid out left to right.
- Use `align="left"` to pin to the left side.
- Use `align="right"` to pin to the right side.
- Use `align-self="start"`, `align-self="center"`, or `align-self="end"` to control vertical placement within the hbox track.
- In `hbox`, `align-self="start"` means top and `align-self="end"` means bottom.
- Unaligned children share remaining width equally unless `width` is specified.
- When at least one child uses `width="auto"` and the hbox has true surplus
  width beyond the preferred widths of the remaining unsized children plus
  hpadding, omitted widths keep their preferred widths and the `auto` children
  split the leftover space evenly.
- In constrained hboxes, and in layout managers without their own `auto`
  sizing policy, `width="auto"` behaves the same as an omitted width.
- In `dir="rtl"`, stacking order reverses: children flow right to left, `align="left"` pins to the right side, and `align="right"` pins to the left side.

### Table Details

- Set `cols` for row-major order (`order="rows"`, the default).
- Set `rows` for column-major order (`order="cols"`).
- Use `colspan` and `rowspan` attributes on cells to span multiple slots.
- Column widths can be fixed (`width="120pt"`), percentage (`width="40%"`), or
  omitted. Omitted columns keep the historical table behavior: they share the
  remaining width equally.
- When at least one single-column cell uses `width="auto"` and the table has
  true surplus width beyond the preferred widths of omitted and auto columns,
  omitted columns keep their preferred widths and auto columns split the
  remaining width evenly. In constrained tables where omitted preferred widths
  can still fit, omitted columns keep those preferred widths and auto columns
  split the remaining width. Only when omitted preferred widths cannot fit does
  the table fall back to equal sharing.
- Cells with `colspan > 1` receive the resolved width of their spanned columns
  but do not drive auto column sizing.
- When at least one row contains a `height="auto"` cell and a height-constrained
  table has true surplus height beyond fixed and preferred row heights, omitted
  rows keep their preferred heights and auto rows split the remaining height
  evenly.
- In natural-height tables, and in constrained tables without surplus,
  `height="auto"` behaves the same as an omitted height.
- When a direct page-child table splits across pages, auto row height is
  evaluated separately for each fragment page.
- In `dir="rtl"`, columns are placed right to left (column 0 at the right edge).

### Flow Details

- Children are placed left to right, wrapping to the next row when the container
  width is exceeded.
- In `dir="rtl"`, children are placed right to left, wrapping back to the right edge.

### Radial Details

- Use `rows` to specify concentric tracks and `cols` to specify angular slots.
- If `angles` is present, LTML treats the values as boundary bearings, then
  normalizes, sorts, deduplicates, and closes the circle automatically.
- At least one of `rows` or `cols` must be specified unless `angles` supplies
  the columns and the missing dimension can be derived from the children.
- `sweep="ccw"` (default) advances grid columns counterclockwise;
  `sweep="cw"` advances them clockwise. In either direction, a colspan keeps
  the first column's starting boundary and extends through subsequent columns.
- `order="rows"` fills sectors around the circle before moving inward.
- `order="cols"` fills sectors inward before advancing to the next angular slot.
- Row `0` is the outermost track; higher row numbers move inward.
- `base-angle` rotates the whole radial grid.
- `row-angle-offsets` rotates individual rows relative to `base-angle`, for
  example `row-angle-offsets="22.5,0,0,0,22.5"`. Missing values default to
  zero; trailing values beyond the rows populated by the dataset are ignored.
- A sector may span multiple rows only when those rows have equivalent angular
  offsets. Differently offset rows cannot share the current annular-sector
  geometry.
- A single distinct `angles` value means one full-circle sector.
- `center-x`, `center-y`, `r`, and `r0` override inferred geometry.
- Explicit sectors contain widgets; wrap sector text in `<label>`.
- Direct non-`<sector>` children are wrapped in implicit sectors automatically.
- Positioned children inside a sector may use `origin-x="start|center|end"` and
  `origin-y="inner|middle|outer"` to anchor to radial reference points.
- Static sector children participate in shape-aware source-order flow.
- Non-static children are overlays; omitted origins use midpoint angle and radius.
- Sector paragraphs curve by default; use `angle="0"` for horizontal wedge-aware wrapping.

### Radial-Out Details

- Use `rows` to specify concentric tracks and `cols` to specify angular slots.
- If `angles` is present, LTML treats the values as boundary bearings, then
  normalizes, sorts, deduplicates, and closes the circle automatically.
- At least one of `rows` or `cols` must be specified unless `angles` supplies
  the columns and the missing dimension can be derived from the children.
- `sweep="ccw"` (default) advances grid columns counterclockwise;
  `sweep="cw"` advances them clockwise. In either direction, a colspan keeps
  the first column's starting boundary and extends through subsequent columns.
- `order="rows"` fills sectors around the circle before moving outward.
- `order="cols"` fills sectors outward before advancing to the next angular slot.
- Row `0` is the innermost track; higher row numbers move outward.
- `base-angle` rotates the whole radial grid.
- `row-angle-offsets` rotates individual logical rows relative to `base-angle`.
  Missing values default to zero; trailing values beyond the rows populated by
  the dataset are ignored.
- A sector may span multiple rows only when those rows have equivalent angular
  offsets.
- A single distinct `angles` value means one full-circle sector.
- `center-x`, `center-y`, `r`, and `r0` override inferred geometry.
- Explicit sectors contain widgets; wrap sector text in `<label>`.
- Direct non-`<sector>` children are wrapped in implicit sectors automatically.
- Positioned children inside a sector may use `origin-x="start|center|end"` and
  `origin-y="inner|middle|outer"` to anchor to radial reference points.
- Static sector children participate in shape-aware source-order flow.
- Non-static children are overlays; omitted origins use midpoint angle and radius.
- Sector paragraphs curve by default; use `angle="0"` for horizontal wedge-aware wrapping.

Example:

```xml
<div layout="radial" rows="3" angles="0,45,90,135,225,315,360" base-angle="-90" width="4in" height="4in">
  <sector colspan="2">
    <label>Curved title</label>
    <label angle="90" position="relative" origin-x="end" origin-y="outer">12</label>
  </sector>
  <p colspan="2">This paragraph curves inside its implicit sector.</p>
  <p colspan="2" angle="0">This paragraph uses horizontal wedge-aware wrapping.</p>
  <sector><label angle="0">Horizontal</label></sector>
</div>
```

### Positioning Details

- Widgets default to `position="static"` and participate in their parent layout.
- `position="absolute"` uses page-space coordinates.
- `position="relative"` offsets the widget from its parent container.
- If `top`, `right`, `bottom`, or `left` is present and `position` is omitted,
  LTML follows legacy ERML behavior and treats the widget as `relative`.
- In `layout="absolute"`, all children are treated as absolute-positioned and
  default missing horizontal/vertical anchors to `left="0"` and `top="0"`.
- In `layout="relative"`, all children are treated as relative-positioned and
  default missing horizontal/vertical anchors to `left="0"` and `top="0"`.
- For text widgets such as `<label>` and `<p>`, `top` anchors the widget box.
  Visible glyphs begin lower at the text ascent/baseline, so text can look
  lower than boxes or shapes with the same `top` value.
- `shift-x` and `shift-y` apply after layout and are especially useful for
  nudging layout-managed widgets. Percent shifts are relative to the shifted
  widget's own resolved width or height.
- `rotate` wraps the widget's normal background/content/border rendering.
- `origin-x` defaults to the widget's left edge; `origin-y` defaults to the
  widget's top edge.
- `z-index="N"` controls paint order among siblings in the same container.
  Lower values paint first, higher values paint later and appear on top.
  Equal `z-index` values preserve source order.

#### Relative Layout

Use `layout="relative"` on a container when you want a local coordinate system
for overlays, callouts, badges, image annotations, or other composed in-page
regions.

- All children are treated as `position="relative"`, even if they do not set
  `position` explicitly.
- Child coordinates are measured from the container's own box, not from the
  page origin.
- Missing horizontal/vertical anchors default to `left="0"` and `top="0"`.
- Width and height can still come from preferred size, percentages, or paired
  sides such as `left` + `right`.
- Negative `right` and `bottom` values anchor inward from the far edge of the
  container.
- Relative-positioned children do not consume space in normal `vbox`, `hbox`,
  or `flow` placement.
- The same behavior also works at page scope when you intentionally want the
  page itself to act as the container.

Example:

```xml
<div left="1in" top="2in" width="4in" height="2in" border="thin" layout="relative">
  <label left="0.2in" top="0.2in">Top-left note</label>
  <rect right="-0.25in" bottom="-0.25in" width="1in" height="0.5in" fill="Gold" />
</div>
```

For a fuller example, see `ltml/samples/test_037_relative_layout.ltml`.

### Page Flow Details

Page flow has two related parts: widget visibility (`display`) and page
continuation (`overflow`).

#### Visibility and Retry

- `display` defaults to `once`.
- `display="none"` (or CSS `display: none`) removes a widget from layout and
  rendering, and can override a repeating display mode assigned by another rule.
- Page `overflow` defaults to `true` for page `layout="flow"`,
  `layout="table"`, and `layout="vbox"`, and allows LTML to retry unprinted
  direct children on later physical pages.
- `overflow` controls page continuation; it does not establish a clipping
  boundary. A height-constrained `flow`, `table`, or `vbox` paints all of its
  children when it has no effective continuation path, even when they extend
  beyond the container.
- When page overflow is disabled, `flow`, `table`, and `vbox` pages paint all
  of their children on the current physical page.
- Direct page-child paragraph, table, and vbox `overflow` defaults to `true`
  and allows the child to be fragmented across physical pages.
- Use `overflow="false"` on a direct page child to keep that child whole.
- A direct `<pgbr/>` ends the current `vbox`, `flow` page, or row-major `table`
  fragment. Leading, consecutive, and trailing markers do not create blank pages.
- `display="even"` and `display="odd"` follow the physical PDF page sequence,
  not `<pageno>` display values.
- `display="last"` renders only on the final physical page generated by an
  overflowing `<page>`. Last-page widgets participate in final-page layout, so
  an aligned footer can carry displaced body content forward rather than overlap it.

#### Continuing Direct Page Children

Only direct page children are tracked as page-flow items. An enabled vbox may
propagate continuation for descendant vboxes, tables, and paragraphs through
an unbroken chain of enabled vboxes. A table can continue its own whole rows
through that chain, but a table cell does not propagate continuation for a
container inside the cell. An `hbox`, nested `flow`, table cell, disabled vbox,
or other non-propagating container breaks the path; vertical layouts below the
break paint their children visibly instead of silently hiding them.

- Direct page-child paragraphs continue by wrapped lines.
- Direct page-child tables continue by whole rows.
- Direct page-child `vbox` containers continue with the next stacked children
  on each physical page.
- Direct page-child `flow` containers do not continue across pages. A page with
  `layout="flow"` can still retry its own direct children.
- Paragraph continuation uses `orphans="2"` and `widows="2"` by default.
- Table continuation is available for `layout="table"` containers.
- VBox continuation is available for `layout="vbox"` containers.
- Nested `<pgbr/>` markers must be direct children of the continuing vbox or
  row-major table they control; markers do not propagate through other layouts.

#### Repeated Chrome on Fragment Pages

- Table `header-rows` and `footer-rows` repeat on every fragment page.
- In `vbox`, direct children with `align="top"` and `align="bottom"` define
  the fragment header/footer bands.
- Those aligned children repeat on fragment pages only when their `display`
  mode is itself repeating, such as `always`, `odd`, or `even`.

#### Layout-Specific Notes

- `hbox` preserves source order for `align="left"` and `align="right"` groups,
  but `hbox` itself does not participate in overflow retries.
- `flow` pages participate in overflow retries; `flow` containers do not
  fragment as direct page children.

---

## Measurement Values

Measurements can be expressed in several forms:

| Form          | Example     | Description |
|---------------|-------------|-------------|
| Bare number   | `72`        | In the current `units` of the element or container. |
| With unit     | `1in`, `2.5cm`, `210mm`, `200dp`, `14pt` | Explicit unit overrides the current `units`. |
| Percentage    | `50%`       | Percentage of the container's content width or height. |
| Relative      | `+10`, `-5pt`, `+0.25in` | Offset from the container's content dimension. |
| Auto          | `auto`      | Automatic layout-managed size. Supported by `hbox` width, `vbox` height, table column width, and table row height; elsewhere it behaves like omitting the dimension. |

`max-width` and `max-height` accept bare numbers, unit-suffixed measurements,
percentages, and relative values. `auto` or an omitted max attribute means
there is no cap.

**Supported units:** `pt` (points, 1/72 inch), `in` (72 pt per inch), `mm`
(72/25.4 pt per millimeter), `cm` (10 mm, so 720/25.4 pt per centimeter), `dp`
(thousandths of an inch: 1000 `dp` = 1 `in`, i.e. 0.072 pt per `dp`).

Implementation note: LTML stores and computes measurements internally in
points. Alternate units are only used in markup attributes. `ParseMeasurement`
converts each specified value to points using:

1. the explicit unit suffix, if present, or
2. the element's current default `units` value, if the measurement is bare.

Measurements are valid wherever `width`, `height`, `margin`, `padding`,
`corners`, positional attributes (`top`, `right`, `bottom`, `left`), and
similar dimension values are accepted.

`rem` is reserved for font sizing only. Use it with `font.size` and `<font
size="...">`; LTML does not accept `rem` for general geometric measurements.

---

## Colors

Colors can be specified by name wherever a color attribute is accepted. Standard
CSS color names are supported (e.g., `red`, `blue`, `LightYellow`, `navy`).
Hexadecimal color notation (e.g., `#ff0000`) is also accepted.

---

## Examples

### Hello, World

```xml
<ltml>
  <page units="in" margin="1">
    <p>Hello, World!</p>
  </page>
</ltml>
```

### Document with Styles and Rich Text

```xml
<ltml units="in" margin="1">
  <font id="title" name="Helvetica" size="24" weight="Bold" color="navy" />
  <font id="body" name="Helvetica" size="12" />
  <pen id="rule" color="black" width="0.5pt" pattern="solid" />

  <page>
    <h>My Document</h>
    <rect height="2pt" border="rule" width="100%" />
    <p font="body">
      This paragraph has <b>bold</b>, <i>italic</i>,
      and <u>underlined</u> text.
    </p>
  </page>
</ltml>
```

### Two-Column Layout

```xml
<ltml>
  <page units="in" margin="1">
    <hbox padding="12pt">
      <p width="50%" style.text-align="justify">Left column content.</p>
      <p width="50%" style.text-align="justify">Right column content.</p>
    </hbox>
  </page>
</ltml>
```

### Table

```xml
<ltml>
  <define id="td" tag="p" border="solid" padding="4pt" />
  <page margin="1in">
    <table order="rows" cols="3" padding="4pt">
      <td>Name</td>
      <td>Age</td>
      <td>City</td>
      <td>Alice</td>
      <td>30</td>
      <td>Boston</td>
      <td colspan="2">Bob and Carol</td>
      <td>Chicago</td>
    </table>
  </page>
</ltml>
```

### Bulleted List

```xml
<ltml units="in" margin="1">
  <bullet id="brand-star" shape="star" width="24pt" r="8pt" brush="Gold" pen="solid" points="5" r0="4pt" />
  <bullet id="ordered-mark" format="%d." width="24pt" />
  <bullet id="badge" src="badge.svg" width="18pt" height="12pt" />
  <page>
    <ul>
      <p>First unordered item</p>
      <p>Second unordered item</p>
    </ul>

    <ul bullets="brand-star">
      <p>Custom unordered marker</p>
      <p>Reuses a named bullet style</p>
    </ul>

    <ol bullets="badge ordered-mark">
      <p>First ordered item</p>
      <p>Second ordered item</p>
      <p>Third ordered item</p>
    </ol>
  </page>
</ltml>
```

### Styles and Classes

```xml
<ltml units="in" margin="1">
  <style>
    p.heading { font.size: 18; font.weight: Bold; style.text-align: center; }
    p.body    { font.size: 11; style.text-align: justify; }
  </style>
  <page>
    <p class="heading">Section Title</p>
    <p class="body">Body text goes here.</p>
  </page>
</ltml>
```
