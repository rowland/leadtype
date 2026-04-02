# LTML Hyperlinks And Indexes — Task Checklist

See [ltml-hyperlinks-and-indexes.md](ltml-hyperlinks-and-indexes.md) for the
design overview.

---

## PDF Foundation

- [x] Add PDF link annotation support.
- [x] Add page `/Annots` output.
- [x] Add URI action support for external links.
- [x] Add internal destination array support for target links.
- [x] Add a `DocWriter` destination registry with first-win semantics.
- [x] Validate unresolved internal targets at `WriteTo`.
- [x] Add rectangle-based public APIs for URI and target links.
- [x] Emit text-link annotations automatically from `PageWriter.flushText`.

## Rich Text Substrate

- [x] Add optional link metadata to `rich_text.RichText`.
- [x] Preserve link metadata through clone, split, trim, merge, and scale.
- [x] Prevent merge from collapsing pieces with different link metadata.

## LTML Markup

- [x] Add inline `<a>` with `uri` / `target` attributes.
- [x] Add explicit `<target>` destinations.
- [x] Treat printed widget/page `id` values as internal destinations.
- [x] Add `<index>` widgets.
- [x] Add hidden `<index_entry>` metadata widgets.
- [x] Render index entries as clickable internal links.
- [x] Enforce `<a>` requires exactly one of `uri` or `target`.
- [x] Enforce `<target>` requires `id`.
- [x] Enforce `<index>` ids are unique.
- [x] Enforce `<index_entry>` requires both `index` and `target`.

## Layout And Render Orchestration

- [x] Add full-tree LTML render-state reset between passes.
- [x] Add a preflight collector pass for destinations and index entries.
- [x] Re-run preflight until index layout stabilizes.
- [x] Cap convergence attempts and fail clearly if stabilization does not happen.
- [x] Keep LTML `Writer` unchanged by using optional writer interfaces.
- [x] Ensure zero-footprint metadata widgets do not force extra overflow pages.

## Tests

- [x] Add `rich_text` tests for link metadata preservation.
- [x] Add `pdf` tests for annotation serialization and unresolved targets.
- [x] Add `pdf` tests for automatic link annotation emission.
- [x] Add LTML parse/validation tests for `<a>`, `<target>`, `<index>`, and `<index_entry>`.
- [x] Add LTML TOC/index convergence tests.
- [x] Add LTML PDF integration coverage for wrapped hyperlink annotations.

## Samples And Docs

- [x] Add LTML samples for links and indexes.
- [x] Add this design document.
- [x] Add this implementation checklist.

## Follow-Up Issues

1. `pdf: expose more general widget-area link helpers and outline/bookmark support`
2. `ltml: support inline destination anchors inside paragraph text`
3. `ltml: preserve entry-local rich styling inside index labels`
4. `ltml: consider title inference / automatic TOC labels as an opt-in`
5. `ltml/pdf: revisit multi-cursor page insertion if future indexing needs true deferred layout`
