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
(`<define>`), and rules (`<rules>`) placed inside a `<page>` are visible only
to that page. Definitions placed directly inside `<ltml>` are visible to all
pages. A page can always reference definitions from its parent `<ltml>` scope,
but other pages cannot see definitions made inside a sibling page.

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

---

### `<page>` — Page

Defines a single page in the document. Pages must be direct children of `<ltml>`.

| Attribute     | Description |
|---------------|-------------|
| `units`       | Unit for measurements on this page. |
| `margin`      | Margin for all sides. |
| `margin-top`, `margin-right`, `margin-bottom`, `margin-left` | Per-side margins. |
| `style`       | Reference to a named `<page>` style. |
| `layout`      | Layout manager to use (`vbox`, `hbox`, `table`, `flow`, `absolute`, `relative`). Default: `vbox`. |
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
| `margin`, `padding` | Spacing around and inside the element. |
| `border`           | Reference to a named `<pen>` style. |
| `fill`             | Reference to a named `<brush>` style. |
| `rotate`           | Rotate the widget around its origin by the given degrees. |
| `origin_x`         | Rotation origin on the x axis: `left`, `center`, `right`, or a measurement. |
| `origin_y`         | Rotation origin on the y axis: `top`, `middle`, `bottom`, or a measurement. |
| `shift`            | Offset the widget after layout using `x,y` measurements. |
| `align`            | Position within parent vbox: `top` (header), `bottom` (footer). |
| `display`          | Retry/visibility policy for repeated page rendering: `once` (default), `always`, `first`, `succeeding`, `even`, `odd`. |
| `split`            | Whether a direct page-child paragraph may split across pages. Defaults to `true`. |
| `orphans`          | Minimum number of lines kept on the first fragment when splitting. Defaults to `2`. |
| `widows`           | Minimum number of lines carried to the continuation fragment when splitting. Defaults to `2`. |
| `colspan`, `rowspan` | Span multiple table cells (when inside a `table`). |

---

### `<span>` — Inline Text Run

Applies font styling to a portion of text within a `<p>`. Must be a child of
`<p>`, `<label>`, or another `<span>`.

```xml
<p>Normal <span font.weight="Bold">bold</span> normal.</p>
```

Supports the same `font.*` attributes as `<p>`.

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
| `layout`         | Layout manager name (see [Layout Managers](#layout-managers)). |
| `cols`           | Number of columns (required for `table` layout). |
| `rows`           | Number of rows (for column-order `table` layout). |
| `order`          | Table fill order: `rows` (default) or `cols`. |
| `split`          | Whether a direct page-child `table` may split by whole rows across pages. Defaults to `true` for table layouts. |
| `header-rows`    | Number of leading table rows that repeat on every fragment page. Defaults to `0`. |
| `footer-rows`    | Number of trailing table rows that repeat on every fragment page. Defaults to `0`. |
| `paragraph-style` | Default paragraph style for child `<p>` elements. |

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
| `margin`, `padding` | Spacing. |

---

### Shape Widgets — `<circle>`, `<ellipse>`, `<polygon>`, `<star>`, `<arc>`, `<pie>`, `<arch>`

Draw shape primitives using the existing PDF shape APIs. These widgets are also
containers, so nested LTML content can still be laid out inside their content
boxes.

```xml
<circle width="1.5in" height="1.5in" border="solid" fill="LightBlue" />
<polygon width="1.5in" height="1.5in" sides="6" rotation="30" border="solid" fill="Gold" />
<pie width="1.5in" height="1.5in" start_angle="30" end_angle="150" border="solid" fill="Pink" />
```

Shared attributes:

| Attribute | Description |
|-----------|-------------|
| `width`, `height` | Layout box dimensions. Shapes are centered within their content boxes. |
| `border` | Reference to a named `<pen>` style for the shape outline. |
| `fill` | Reference to a named `<brush>` style for the interior. |
| `margin`, `padding` | Spacing around and inside the shape widget. |
| `reverse` | Reverse the path direction for shapes that support it. |

Shape-specific attributes:

| Tag | Attributes |
|-----|------------|
| `<circle>` | `r` for an explicit radius. Otherwise radius is inferred from the content box. |
| `<ellipse>` | `rx`, `ry` for explicit radii. Otherwise radii are inferred from width and height. |
| `<polygon>` | `r`, `sides`, `rotation`. |
| `<star>` | `r` or `r1` for outer radius, `r2` for inner radius, `points`, `rotation`. |
| `<arc>` | `r`, `start_angle`, `end_angle`. |
| `<pie>` | `r`, `start_angle`, `end_angle`. |
| `<arch>` | `r1`, `r2`, `start_angle`, `end_angle`. |

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
| `fit="shrink"` | If `width` is set and the text is too wide, shrink the label text proportionally until it fits, down to a minimum of 6pt. |
| `width`, `height` | Optional explicit dimensions. |
| `margin`, `padding` | Spacing around and inside the label. |
| `border` | Reference to a named `<pen>` style. |
| `fill` | Reference to a named `<brush>` style. |

The built-in `<br/>` alias expands to an empty `<label/>`, which behaves as a
line break in stacked layouts. Labels still do not wrap; `fit="shrink"` scales
the rendered text instead.

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
| `margin`, `padding` | Spacing around and inside the block. |
| `border` | Reference to a named `<pen>` style. |
| `fill` | Reference to a named `<brush>` style. |

`<pre>` preserves internal spaces and line breaks, does not wrap lines, trims a
single surrounding newline from block content, removes common leading
indentation from non-blank lines, and expands tab characters to four spaces.

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
| `margin`, `padding` | Spacing around and inside the widget box. |
| `border` | Reference to a named `<pen>` style. |
| `fill` | Reference to a named `<brush>` style. |

The current implementation supports the same file-based JPEG workflow as the
PDF layer. The LTML API stays format-agnostic so future PNG support can be
added without changing the tag shape.

---

### `<line>` — Line Segment

Draws a line segment using the configured pen style.

```xml
<line width="100%" height="12pt" style="dashed" />
```

| Attribute | Description |
|-----------|-------------|
| `style` | Reference to a named `<pen>` style. |
| `angle` | Line angle in degrees. `0` points right; positive angles rotate clockwise in page coordinates. |
| `length` | Optional explicit line length. |
| `width`, `height` | Optional layout box dimensions used to infer the line length and placement when `length` is omitted. |
| `margin`, `padding` | Spacing around and inside the widget box. |
| `border` | Optional enclosing widget border, separate from the line stroke. |

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

Page size and orientation cannot be defined via custom styles in LTML markup.
The `style` attribute on `<page>` accepts only the built-in size names listed
below. Orientation must be set directly on the `<page>` element.

```xml
<page style="A4" orientation="landscape" units="cm" margin="2">
  <!-- content -->
</page>
```

**Built-in page sizes** (set via `style` on `<page>`):

| Name     | Width × Height (pt) |
|----------|---------------------|
| `letter` | 612 × 792 (default) |
| `legal`  | 612 × 1008 |
| `A4`     | 595 × 842 |
| `B5`     | 499 × 708 |
| `C5`     | 459 × 649 |

The `orientation` attribute (`portrait` or `landscape`) swaps width and height
when combined with `style`. Explicit `width` and `height` values (in points)
can also be set directly on `<page>` to use a custom page size.

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
| `<layer>` | `<div>` | `position="relative"`, `width="100%"`, `height="100%"` |
| `<br>`  | `<label>`  | *(empty line break)* |

---

## Rules (CSS-like Selectors)

Apply attributes to elements based on tag name, id, and class selectors.

```xml
<rules>
  p { font.size: 14; }
  p.intro { font.weight: Bold; style.text-align: justify; }
  div#footer { margin-top: 20; }
</rules>
```

Rules use CSS-style selector syntax:

| Pattern         | Matches |
|-----------------|---------|
| `p`             | All `<p>` elements |
| `p.classname`   | `<p>` elements with `class="classname"` |
| `p#myid`        | `<p>` elements with `id="myid"` |
| `div p`         | `<p>` elements anywhere inside a `<div>` |
| `div > p`       | `<p>` elements that are direct children of a `<div>` |
| `p, span`       | All `<p>` and `<span>` elements |

Rules inside `<!-- XML comments -->` are also parsed, allowing rules to be
commented out with nested comment delimiters.

Attribute priority (lowest to highest):
1. Default attributes from an alias (`<define>`)
2. Attributes from matching rules
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

### VBox Details

- Children stack top to bottom.
- Width defaults to content width of the container.
- Use `align="top"` to pin an element to the top (header behavior).
- Use `align="bottom"` to pin an element to the bottom (footer behavior).

### HBox Details

- Children are laid out left to right.
- Use `align="left"` to pin to the left side.
- Use `align="right"` to pin to the right side.
- Unaligned children share remaining width equally unless `width` is specified.

### Table Details

- Set `cols` for row-major order (`order="rows"`, the default).
- Set `rows` for column-major order (`order="cols"`).
- Use `colspan` and `rowspan` attributes on cells to span multiple slots.
- Column widths can be fixed (`width="120pt"`), percentage (`width="40%"`), or
  automatic (equal share of remaining space).

### Flow Details

- Children are placed left to right, wrapping to the next row when the container
  width is exceeded.

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
- `origin_x` defaults to the widget's left edge; `origin_y` defaults to the
  widget's top edge.
- `z_index="N"` controls paint order among siblings in the same container.
  Lower values paint first, higher values paint later and appear on top.
  Equal `z_index` values preserve source order.

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

### Rules and Classes

```xml
<ltml units="in" margin="1">
  <rules>
    p.heading { font.size: 18; font.weight: Bold; style.text-align: center; }
    p.body    { font.size: 11; style.text-align: justify; }
  </rules>
  <page>
    <p class="heading">Section Title</p>
    <p class="body">Body text goes here.</p>
  </page>
</ltml>
```
