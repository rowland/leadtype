# LeadType Markup Language (LTML) Syntax Reference

LTML is an XML-based markup language for generating PDF documents. It provides
declarative control over layout, typography, and visual elements.

## Document Structure

Every LTML document begins with a root `<ltml>` element containing style
definitions and one or more `<page>` elements.

```xml
<ltml units="in" margin="1">
  <!-- style definitions -->
  <page>
    <!-- content -->
  </page>
</ltml>
```

### Scope

Both `<ltml>` and `<page>` establish a style scope. Style definitions
(`<font>`, `<pen>`, `<brush>`, `<para>`, `<bullet>`, `<layout>`), aliases
(`<define>`), and selector styles (`<style>`) placed inside a
`<page>` are visible only to that page. Definitions placed directly inside
`<ltml>` are visible to all pages. A page can always reference definitions from
its parent `<ltml>` scope, but other pages cannot see definitions made inside a
sibling page.

```xml
<ltml>
  <font id="body" name="Helvetica" size="12" />  <!-- shared by all pages -->

  <page>
    <font id="title" name="Helvetica" size="24" weight="Bold" />  <!-- this page only -->
    <p font="title">Page One</p>
    <p font="body">Body text.</p>
  </page>

  <page>
    <!-- "title" is not visible here -->
    <p font="body">Page Two</p>
  </page>
</ltml>
```

---

## Elements

### `<ltml>` — Document Root

The root element. Attributes set here apply as defaults to all pages.

| Attribute | Description |
|-----------|-------------|
| `units`   | Default unit for measurements (`pt`, `in`, `cm`). Default: `pt`. |
| `margin`  | Page margin applied to all pages unless overridden. |
| `compress-pages` | If `true`, compress page content streams with `FlateDecode`. Default: `false`. |
| `compress-to-unicode` | If `true`, compress generated `ToUnicode` streams. Default: `false`. |
| `compress-embedded-fonts` | If `true`, compress embedded font subset streams. Default: `false`. |
| `ua` | If `true`, opt the whole document into tagged PDF output and accessibility structure generation. Default: `false`. |

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

When tagged output is enabled, LTML fills `/ActualText` automatically for
paragraphs and labels from their fully resolved plain text, including inline
links and dynamic text such as `<pageno>`. LTML does not emit `/ActualText` for
inline spans or links.

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
| `grid`        | Optional debug grid. Use `true` for the default `0.25in` grid or supply a measurement such as `0.5in`. |
| `overflow`    | If `true`, allow the page to retry unprinted direct children on additional physical pages. Current support is page-only. |
| `font`        | Reference to a named `<font>` style. |
| `fill`        | Reference to a named `<brush>` style for the background. |
| `border`      | Reference to a named `<pen>` style for all borders. |

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
| `font.size`        | Font size in points. |
| `font.color`       | Font color (named color or hex). |
| `font.weight`      | Font weight (`Bold`, or empty for normal). |
| `font.style`       | Font style (`Italic`, `Oblique`, or empty for normal). |
| `font.underline`   | `true` or `false`. |
| `font.strikeout`   | `true` or `false`. |
| `font.line-height` | Line spacing multiplier (e.g., `1.5`). |
| `style`            | Reference to a named `<para>` style. |
| `style.text-align` | Text alignment: `left`, `center`, `right`, `justify`. |
| `style.valign`     | Vertical alignment: `top`, `middle`, `bottom`, `baseline`. |
| `bullet`           | Reference to a named `<bullet>` style. |
| `width`, `height`  | Explicit dimensions. |
| `margin`, `margin-top`, `margin-right`, `margin-bottom`, `margin-left` | Outer spacing around the element. |
| `padding`, `padding-top`, `padding-right`, `padding-bottom`, `padding-left` | Inner spacing inside the element box. |
| `border`           | Reference to a named `<pen>` style. |
| `fill`             | Reference to a named `<brush>` style. |
| `rotate`           | Rotate the widget around its origin by the given degrees. |
| `origin-x`         | Rotation origin on the x axis: `left`, `center`, `right`, or a measurement. |
| `origin-y`         | Rotation origin on the y axis: `top`, `middle`, `bottom`, or a measurement. |
| `shift`            | Offset the widget after layout using `x,y` measurements. |
| `align`            | Position within parent vbox: `top` (header), `bottom` (footer). |
| `display`          | Retry/visibility policy for repeated page rendering: `once` (default), `always`, `first`, `succeeding`, `even`, `odd`. |
| `split`            | Whether a direct page-child paragraph may split across pages. Defaults to `true`. |
| `orphans`          | Minimum number of lines kept on the first fragment when splitting. Defaults to `2`. |
| `widows`           | Minimum number of lines carried to the continuation fragment when splitting. Defaults to `2`. |
| `colspan`, `rowspan` | Span multiple table cells (when inside a `table`). |
| `role` | Override the generated PDF structure type when `ua="true"`. Default tagged output uses `P`. |

When `ua="true"`, paragraphs automatically emit `/ActualText` from their fully
resolved plain text. That text includes inline links and `<pageno>` output, but
not decorative bullet chrome or non-text decoration.

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
still emit `Link` structure elements, but `/ActualText` remains on the enclosing
paragraph or label rather than on the span or link itself.

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
| `split`          | Whether a direct page-child `table` may split by whole rows across pages. Defaults to `true` for table layouts. |
| `header-rows`    | Number of leading table rows that repeat on every fragment page. Defaults to `0`. |
| `footer-rows`    | Number of trailing table rows that repeat on every fragment page. Defaults to `0`. |
| `base-angle`     | Base angle in degrees for radial sector boundaries. Default: `0`. |
| `angles`         | Comma-separated angular boundary bearings relative to `base-angle`. LTML normalizes, sorts, and deduplicates them before building sectors. |
| `sweep`          | Radial sector sweep direction: `ccw` (default) or `cw`. This changes how sectors span between boundaries without changing what the angle numbers mean. |
| `center-x`, `center-y` | Optional radial center coordinates in the container's content box. |
| `r`              | Optional outer radius for radial layout. Otherwise LTML infers it from the smaller content dimension. |
| `r0`             | Optional inner radius for radial layout. Preferred alias when paired with `r`. |
| `paragraph-style` | Default paragraph style for child `<p>` elements. |
| `role` | Override the computed PDF structure type when `ua="true"`, for example `L` or `Table`. |

---

### `<sector>` — Radial Cell

`<sector>` is a container used inside a `layout="radial"` or `layout="radial-out"` parent. It behaves
like a radial table cell: LTML assigns it one wedge or annular-sector region
based on the parent grid, and the sector owns the special paint and layout
behavior for that region.

```xml
<div layout="radial" rows="2" cols="6">
  <sector colspan="2" fill="AliceBlue" border="solid">Curved arc text</sector>
  <sector rowspan="2" colspan="2">
    <p>Paragraphs wrap to the changing line width of the sector.</p>
  </sector>
</div>
```

Direct non-`<sector>` children of a radial container are wrapped in an
implicit sector automatically. Their XML attributes are applied both to the
implicit sector and to the original child widget.

| Attribute | Description |
|-----------|-------------|
| `colspan`, `rowspan` | Span multiple radial slots, just like table cells. |
| `facing` | Curved-text/content facing: `auto` (default), `upright`, or `upside-down`. |
| `angle` | Absolute angle in degrees for sector content. Overrides the default tangent-based orientation. |
| `text-align` | For inline sector text, anchor to the sector `left`/start, `center`, or `right`/end. |
| `origin-x` | For positioned child widgets inside a sector, `start`, `center`, and `end` are radial aliases in addition to the normal box-relative values. |
| `origin-y` | For positioned child widgets inside a sector, `inner`, `middle`, and `outer` are radial aliases in addition to the normal box-relative values. |

Sector text comes in two flavors:

- Inline text written directly inside `<sector>` paints as curved text on the
  sector's midpoint arc.
- Nested widgets such as `<label>` and `<p>` use ordinary LTML paint/layout,
  but the sector can rotate and clip them to the wedge.

Paragraphs placed in a sector use true sector-aware wrapping. LTML computes
the usable line width from the actual wedge shape for each line instead of
wrapping to one fixed rectangle.

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
| `corners`  | Corner radius for rounded corners, in current units. |
| `margin`, `margin-top`, `margin-right`, `margin-bottom`, `margin-left` | Outer spacing around the rectangle widget. |
| `padding`, `padding-top`, `padding-right`, `padding-bottom`, `padding-left` | Inner spacing inside the rectangle widget. |

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
| `text-align` | Label text alignment: `left`, `center`, `right`. Affects the text anchor inside the label box. |
| `angle` | Rotate only the label text by the given degrees. Border/fill/background stay axis-aligned. |
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

When `ua="true"`, labels automatically emit `/ActualText` from their fully
resolved plain text, including inline links and `<pageno>` output.

---

### Curved Text Roadmap

Curved text is available today for inline text written directly inside
`<sector>` elements. LTML still does not have a general-purpose curved-text
widget for arbitrary circular or path-based text outside radial sectors.

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
When `<pre>` participates in tagged output through `role`, LTML uses that same
resolved preformatted text for `/ActualText`.

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
| `width`, `height` | Optional explicit dimensions. If only one is supplied, the other is inferred from the SVG aspect ratio. |
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

## Style Definitions

Style definitions are placed inside `<ltml>` (or `<page>` for page-scoped
styles) before the content that uses them.

### `<font>` — Font Style

```xml
<font id="title" name="Helvetica" size="24" color="navy" weight="Bold" />
```

| Attribute     | Description |
|---------------|-------------|
| `id`          | Name used to reference this style. |
| `name`        | Font family name. |
| `size`        | Font size in points. |
| `color`       | Text color. |
| `weight`      | `Bold`, or omit for normal. |
| `style`       | `Italic`, `Oblique`, or omit for normal. |
| `underline`   | `true` or `false`. |
| `strikeout`   | `true` or `false`. |
| `line-height` | Line spacing multiplier. |

**Default font:** Helvetica 12pt.

**Built-in font style:** `fixed` — Courier New 12pt.

---

### `<pen>` — Pen Style (for borders and lines)

```xml
<pen id="rule" color="black" width="1pt" pattern="solid" />
```

| Attribute  | Description |
|------------|-------------|
| `id`       | Name used to reference this style. |
| `color`    | Line color. |
| `width`    | Line width (with optional unit suffix, e.g. `2pt`, `0.5mm`). |
| `pattern`  | Line pattern: `solid`, `dashed`, `dotted`. |

**Built-in pen styles:** `solid`, `dashed`, `dotted` — all black, hairline width.

You may also use a color name directly as a border value to create an
auto-generated solid pen for that color:

```xml
<rect border="red" />
```

---

### `<brush>` — Brush Style (for fill colors)

```xml
<brush id="highlight" color="yellow" />
```

| Attribute | Description |
|-----------|-------------|
| `id`      | Name used to reference this style. |
| `color`   | Fill color. |

You may also use a color name directly as a fill value:

```xml
<rect fill="LightBlue" />
```

---

### `<para>` — Paragraph Style

```xml
<para id="body" text-align="justify" valign="top" bullet="bstar" />
```

| Attribute    | Description |
|--------------|-------------|
| `id`         | Name used to reference this style. |
| `text-align` | `left`, `center`, `right`, `justify`. |
| `valign`     | `top`, `middle`, `bottom`, `baseline`. |
| `bullet`     | Reference to a named `<bullet>` style. |

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
| `text`    | The bullet character(s) to render. |
| `width`   | Space reserved for the bullet. |
| `units`   | Units for `width`. |

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

| Name     | Width × Height (pt) |
|----------|---------------------|
| `letter` | 612 × 792 (default) |
| `legal`  | 612 × 1008 |
| `A4`     | 595 × 842 |
| `B5`     | 499 × 708 |
| `C5`     | 459 × 649 |

The `orientation` attribute (`portrait` or `landscape`) swaps width and height
when combined with `style`. Explicit `width` and `height` values can also be
set directly on `<page>` to use a one-off custom page size.

```xml
<page style="A4" orientation="landscape" units="cm" margin="2">
  <!-- content -->
</page>
```

#### Custom named page sizes (`<pagestyle>`)

Use `<pagestyle>` to define a reusable named page size anywhere a `<page>` has
not yet been opened (typically at the top of the document):

```xml
<pagestyle id="book" units="in" width="6" height="9"/>
<page style="book" margin="0.75in">
  <!-- content -->
</page>
```

| Attribute     | Description |
|---------------|-------------|
| `id`          | Name used in the `style` attribute of `<page>`. Required. |
| `width`       | Page width. |
| `height`      | Page height. |
| `units`       | Unit for `width` and `height`: `pt` (default), `in`, `cm`, `mm`. |
| `orientation` | `portrait` (default) or `landscape`. |

Once defined, the name is available to all subsequent `<page>` elements in the
same scope, just like the built-in size names.

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
| `<b>`   | `<span>`   | `font.weight="Bold"` |
| `<i>`   | `<span>`   | `font.style="Italic"` |
| `<u>`   | `<span>`   | `font.underline="true"` |
| `<s>`   | `<span>`   | `font.strikeout="true"` |
| `<hbox>` | `<div>`   | `layout="hbox"` |
| `<vbox>` | `<div>`   | `layout="vbox"` |
| `<table>` | `<div>` | `layout="table"` |
| `<disc>` | `<div>` | `layout="radial"` |
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

- Row and column pseudo-classes apply only to direct children of a
  `layout="table"` container.
- Row and column numbering is zero-based.
- For `rowspan` and `colspan`, LTML matches row/column pseudo-classes using the
  widget's anchor cell rather than every covered cell.

CSS-style `/* ... */` comments are ignored inside `<style>`. Rules inside
`<!-- XML comments -->` are also parsed, allowing selectors to be commented out
with nested comment delimiters.

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
This mirrors layout placement and also changes the default horizontal alignment of
paragraphs and labels to the right unless `text-align` is set explicitly. It does
not automatically change paragraph shaping or bidi behavior inside text widgets.

### VBox Details

- Children stack top to bottom.
- Width defaults to content width of the container.
- Use `align="top"` to pin an element to the top (header behavior).
- Use `align="bottom"` to pin an element to the bottom (footer behavior).
- Use `align-self="start"`, `align-self="center"`, or `align-self="end"` to control horizontal placement within the vbox row.
- In `dir="rtl"`, `align-self="start"` means right and `align-self="end"` means left.
- In `dir="rtl"`, children are flush against the right edge plus padding instead of the left.

### HBox Details

- Children are laid out left to right.
- Use `align="left"` to pin to the left side.
- Use `align="right"` to pin to the right side.
- Use `align-self="start"`, `align-self="center"`, or `align-self="end"` to control vertical placement within the hbox track.
- In `hbox`, `align-self="start"` means top and `align-self="end"` means bottom.
- Unaligned children share remaining width equally unless `width` is specified.
- In `dir="rtl"`, stacking order reverses: children flow right to left, `align="left"` pins to the right side, and `align="right"` pins to the left side.

### Table Details

- Set `cols` for row-major order (`order="rows"`, the default).
- Set `rows` for column-major order (`order="cols"`).
- Use `colspan` and `rowspan` attributes on cells to span multiple slots.
- Column widths can be fixed (`width="120pt"`), percentage (`width="40%"`), or
  automatic (equal share of remaining space).
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
- `sweep="ccw"` (default) spans each sector to the next boundary in ascending
  order; `sweep="cw"` spans each sector to the previous boundary in the cycle.
- `order="rows"` fills sectors around the circle before moving inward.
- `order="cols"` fills sectors inward before advancing to the next angular slot.
- Row `0` is the outermost track; higher row numbers move inward.
- `base-angle` rotates the whole radial grid.
- A single distinct `angles` value means one full-circle sector.
- `center-x`, `center-y`, `r`, and `r0` override inferred geometry.
- Inline text written directly in `<sector>` follows the arc.
- Direct non-`<sector>` children are wrapped in implicit sectors automatically.
- Positioned children inside a sector may use `origin-x="start|center|end"` and
  `origin-y="inner|middle|outer"` to anchor to radial reference points.

### Radial-Out Details

- Use `rows` to specify concentric tracks and `cols` to specify angular slots.
- If `angles` is present, LTML treats the values as boundary bearings, then
  normalizes, sorts, deduplicates, and closes the circle automatically.
- At least one of `rows` or `cols` must be specified unless `angles` supplies
  the columns and the missing dimension can be derived from the children.
- `sweep="ccw"` (default) spans each sector to the next boundary in ascending
  order; `sweep="cw"` spans each sector to the previous boundary in the cycle.
- `order="rows"` fills sectors around the circle before moving outward.
- `order="cols"` fills sectors outward before advancing to the next angular slot.
- Row `0` is the innermost track; higher row numbers move outward.
- `base-angle` rotates the whole radial grid.
- A single distinct `angles` value means one full-circle sector.
- `center-x`, `center-y`, `r`, and `r0` override inferred geometry.
- Inline text written directly inside `<sector>` follows the arc.
- Direct non-`<sector>` children are wrapped in implicit sectors automatically.
- Positioned children inside a sector may use `origin-x="start|center|end"` and
  `origin-y="inner|middle|outer"` to anchor to radial reference points.

Example:

```xml
<div layout="radial" rows="3" angles="0,45,90,135,225,315,360" base-angle="-90" width="4in" height="4in">
  <sector colspan="2">Curved title</sector>
  <p colspan="2">This paragraph is wrapped by an implicit sector.</p>
  <sector angle="0"><label>12</label></sector>
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
- `shift="x,y"` applies after layout and is especially useful for nudging
  layout-managed widgets.
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

- `display` defaults to `once`.
- `overflow` is currently honored only on `<page>`.
- Overflow retries only reconsider direct children of the page.
- Direct page-child paragraphs can split by wrapped lines.
- Direct page-child tables can split by whole rows.
- Paragraph splitting defaults to `split="true"`, `orphans="2"`, and `widows="2"`.
- Table splitting defaults to `split="true"` for `layout="table"` containers.
- Table `header-rows` and `footer-rows` repeat on every fragment page.
- `align="top"` and `align="bottom"` in `vbox` behave like repeating header/footer slots when paired with a repeating `display` value.
- `align="left"` and `align="right"` in `hbox` preserve source order, but `hbox` does not participate in overflow retries.
- `display="even"` and `display="odd"` follow physical PDF page sequence, not `<pageno>` display values.

---

## Measurement Values

Measurements can be expressed in several forms:

| Form          | Example     | Description |
|---------------|-------------|-------------|
| Bare number   | `72`        | In the current `units` of the element or container. |
| With unit     | `1in`, `2.5cm`, `14pt` | Explicit unit overrides the current `units`. |
| Percentage    | `50%`       | Percentage of the container's content width or height. |
| Relative      | `+10`, `-5` | Offset from the container's content dimension. |

**Supported units:** `pt` (points), `in` (inches, 72pt), `cm` (centimeters, 28.35pt).

Implementation note: LTML stores and computes measurements internally in
points. Alternate units are only used in markup attributes. `ParseMeasurement`
converts each specified value to points using:

1. the explicit unit suffix, if present, or
2. the element's current default `units` value, if the measurement is bare.

Measurements are valid wherever `width`, `height`, `margin`, `padding`,
`corners`, positional attributes (`top`, `right`, `bottom`, `left`), and
similar dimension values are accepted.

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
  <font id="zapf" name="ZapfDingbats" size="12" />
  <bullet id="dot" font="zapf" text="l" width="18pt" />
  <layout id="vbox" padding="4" />
  <page>
    <p bullet="dot">First item</p>
    <p bullet="dot">Second item</p>
    <p bullet="dot">Third item</p>
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
