# Tagged PDF Accessibility Plan

## Summary

- Write the plan in `docs/tagged-pdf-accessibility-plan.md`.
- File one umbrella GitHub issue titled `Accessible PDF v1: tagged PDF structure and LTML accessibility attributes`.
- Scope the work in two implementation parts after an initial research section:
  1. `pdf`-level tagged-PDF substrate
  2. LTML authoring exposure with automatic defaults plus explicit overrides
- Target v1 at text, links, images, and artifacting decorative/page-chrome output. Defer broader table/list/container semantics.

## Research Findings

- Tagged PDF needs a logical structure tree rooted at `/StructTreeRoot`, catalog `/MarkInfo << /Marked true >>`, per-page marked-content association via `MCID`, and a `ParentTree`/`StructParents` back-map so viewers can resolve content to structure elements.
  Source: Adobe PDF Reference 1.6, Chapters 10.5-10.8.
- `/ActualText` can live on a structure element in PDF 1.4; marked-content-property-list `/ActualText` is a PDF 1.5 feature.
  V1 should use structure-element `/ActualText` so tagged output only needs a minimum PDF version of 1.4.
  Source: Adobe PDF Reference 1.6.
- Accessible links should pair `/Link` structure elements with link annotations, and repeated running headers/footers should be artifacts rather than normal reading-order content.
  Sources: W3C PDF11 and W3C PDF14.

## Part 1: `pdf` Tagged-PDF Substrate

- Add a tagged-PDF subsystem to `pdf.DocWriter` with document-level accessibility state, `StructTreeRoot`, `StructElem`, marked-content reference (`MCR`) support, object reference (`OBJR`) support, parent-tree serialization, and per-page `StructParents` values.
- Raise the PDF header version to at least `1.4` only when tagged output is enabled. Leave current untagged output at `1.3`.
- Add scoped `pdf` APIs for semantic and artifact output. Use an options type that initially exposes `ActualText string`.
- The semantic API should create or reuse structure elements and attach emitted content via `MCID`.
- The artifact API should wrap output in `/Artifact ... BMC/EMC` and keep it out of the structure tree.
- Record link annotations created inside an active `Link` semantic scope as `OBJR` children of that same structure element.
- Keep semantic tagging off by default so existing `pdf` callers and output stay unchanged unless they opt in.

## Part 2: LTML Accessibility Attributes

- Add LTML document opt-in with `pdf.tagged="true"` on `<ltml>`.
- Also auto-enable tagging if any descendant uses a `pdf.*` accessibility attribute so those attributes never silently no-op.
- Add LTML attributes `pdf.tag`, `pdf.actual-text`, and `pdf.artifact` on relevant widgets and inline elements.
- Use automatic defaults in tagged mode:
  - `<p>` and `<label>` => `P`
  - `<a>` => `Link`
  - `<image>` => `Figure`
  - `<span>` => untagged unless `pdf.tag` or `pdf.actual-text` is present
  - backgrounds, borders, shape primitives, debug grid, and `<pageno>` => artifact by default
- Preserve LTML compatibility by using an internal optional writer interface rather than widening the public `ltml.Writer` interface.
- Carry stable logical accessibility IDs through LTML and rich-text metadata so wrapped or split inline content can append multiple `MCID` and annotation references to one logical structure element instead of creating unrelated fragments.

## Test Plan

- Add `pdf` serialization tests that assert `/StructTreeRoot`, `/MarkInfo`, `/ParentTree`, page `/StructParents`, `MCR`/`OBJR`, and `/ActualText` output.
- Add `pdf` page-writer tests for tagged text, tagged images, artifact-wrapped decorations, and link annotations associated with `/Link` structure elements.
- Add LTML integration tests for:
  - default paragraph, label, link, and image mappings
  - explicit `pdf.tag` overrides
  - `pdf.actual-text` on inline text
  - `pdf.artifact="true"` suppressing structure output
  - wrapped links and spans still attaching all rendered fragments to one logical tag
  - `<pageno>` and decorative chrome rendering as artifacts
- Add one regression test confirming untagged LTML/PDF output remains unchanged when accessibility features are not enabled.

## Assumptions And Defaults

- V1 supports standard PDF structure types only. No custom role-map authoring surface in this slice.
- V1 stores `/ActualText` on structure elements, not marked-content property lists.
- `/Alt` for figures, document `/Lang`, full list/table semantics, PDF/UA conformance work, and richer artifact property lists are follow-up items, not part of this ticket.
- The GitHub issue should mirror the doc structure: research summary, Part 1, Part 2, test plan, and explicit out-of-scope items.
