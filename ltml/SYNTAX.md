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
| `align`            | Position within parent vbox: `top` (header), `bottom` (footer). |
| `colspan`, `rowspan` | Span multiple table cells (when inside a `table`). |

---

### `<span>` — Inline Text Run

Applies font styling to a portion of text within a `<p>`. Must be a child of
`<p>` or another `<span>`.

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
