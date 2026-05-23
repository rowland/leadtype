# LTML Hyperlinks And Indexes

This document describes the first LTML hyperlink and index implementation.
It is intentionally TOC-first, but the PDF and LTML substrate is general enough
to support broader internal linking over time.

---

## Scope

The implementation adds three capabilities:

- Inline hyperlinks with `<a>` for LTML text.
- Internal destinations from page/widget `id` values and explicit `<target>` elements.
- A general `index` / `index-entry` mechanism for tables of contents and similar lists.

V1 keeps the authoring model intentionally small:

- `<a uri="...">` creates an external link.
- `<a target="...">` creates an internal link.
- `href` and `page` are not part of v1.
- `index-entry` labels are explicit author text; there is no title inference fallback.
- Duplicate destination names resolve to the first printed destination.

---

## LTML Markup

### Inline links

Use `<a>` inside paragraphs and labels:

```xml
<p><a uri="https://example.com">External link</a></p>
<p><a target="details">Jump to details</a></p>
```

Exactly one of `uri` or `target` must be present.

### Destinations

Any printed page or widget with an `id` becomes a destination.

```xml
<page id="chapter_1">
  <label id="chapter_1_title">Chapter 1</label>
</page>
```

Explicit non-rendering targets use `<target>`:

```xml
<target id="appendix_a" />
```

Page destinations resolve to the page content origin. Other widget destinations
resolve to the widget's top-left corner.

### Indexes

`<index>` renders collected entries. `<index-entry>` contributes hidden metadata
from later pages.

```xml
<page layout="vbox">
  <index id="main_toc">
    <p><index-title font.weight="Bold" /><index-leader /><index-page /></p>
  </index>
</page>

<page>
  <label id="intro">Introduction</label>
  <index-entry index="main_toc" target="intro">Introduction</index-entry>
</page>
```

`<index>` accepts one row-template child, typically a `<p>` or `<label>`.
That template is replayed once per resolved index entry. Three inline
placeholders are available inside the row template:

- `<index-title />` resolves to the entry label text
- `<index-leader />` expands into a dot leader between title and page number
- `<index-page />` resolves to the destination page number

`<index-leader />` is the index-facing alias for the general inline
`<leader />` widget. It accepts the same attributes, including
`text="..."` for a custom repeated pattern, and may be styled with inline
`font.*` attributes like the other index placeholders.

Entries are still rendered in encounter order. By default, an empty
`<index id="..."/>` keeps the legacy behavior of dot leaders with a
right-aligned page number. The rendered row links to the entry's target
destination.

---

## PDF Implementation

The `pdf` package now supports:

- link annotations with `/Subtype /Link`
- URI actions for external links
- internal destination arrays for GoTo-style navigation
- page `/Annots` output
- a destination registry on `DocWriter`
- generic rectangle-based APIs for callers outside LTML

LTML text links piggyback on `rich_text.RichText` metadata so wrapped links are
emitted as multiple annotation rectangles automatically.

---

## Render Strategy

Indexes are rendered with an iterative preflight loop:

1. Run a no-output preflight pass to collect printed destinations and index
   entries.
2. Resolve destination page numbers into an index snapshot.
3. Re-run preflight until the snapshot stabilizes, or fail after a bounded
   number of passes.
4. Run one final real render with the stable snapshot.

This avoids a larger `NewPageAfter` / multi-cursor refactor in the first
delivery while still supporting TOCs that grow to an unknown number of pages.

---

## Current Limits

- V1 links are inline-text only.
- Explicit `<target>` is block/widget-level only, not an inline anchor inside a
  paragraph.
- `index-entry` labels are still plain text; rich entry-local label markup is
  not part of this pass.
- Missing internal targets are render errors.
- Multiple `<index>` widgets may exist, but each `index` id must be unique.
