# LTML plan: component-backed `<pre>` with inline or src content

## Goal

Allow LTML `<pre>` to take its preformatted content from either:

- inline text inside the tag, or
- a `src` attribute pointing to a file or URL resolved through the existing LTML component asset-loading rules.

The implementation should preserve current inline `<pre>` behavior and should
not introduce a separate source-loading mechanism just for `<pre>`.

## Current state

- `ltml.StdPre` currently embeds `StdWidget`.
- Inline `<pre>` content is accumulated through `HasText` / `AddText(...)`.
- `<pre>` does not currently support `src`.
- `StdPre` owns its own text normalization flow through `Lines()`, including:
  - CRLF / CR normalization,
  - tab expansion to four spaces,
  - trimming one surrounding blank line,
  - dedenting common leading indentation from non-blank lines.
- Component-backed tags already use `StdComponent` for:
  - raw body capture,
  - `src` attribute handling,
  - asset FS support,
  - `ParseFile(...)` relative-path resolution,
  - optional network-backed asset loading.
- The built-in `<svg>` tag is the closest built-in example of a component-backed
  tag that supports both inline content and `src`, with `src` taking precedence.

## Desired public behavior

Inline `<pre>` should continue to work exactly as it does today:

```xml
<pre font="fixed" border="solid" padding="6pt">
  if x &lt; 1 {
    return
  }
</pre>
```

`<pre>` should also accept a `src` attribute:

```xml
<pre src="snippet.txt" font="fixed" border="solid" padding="6pt" />
```

When both inline content and `src` are present, `src` should win:

```xml
<pre src="snippet.txt">
  this inline body is ignored when src is present
</pre>
```

Target semantics:

- [ ] Inline `<pre>` keeps today’s text semantics.
- [ ] XML entities in inline content continue to decode to literal characters.
- [ ] Tabs expand to four spaces.
- [ ] Line endings normalize from `\r\n` / `\r` to `\n`.
- [ ] One surrounding blank line is trimmed.
- [ ] Common indentation is removed from non-blank lines.
- [ ] Blank lines inside the block remain preserved.
- [ ] `src` content is loaded lazily using the existing component asset rules.
- [ ] `src` takes precedence over inline content when both are present.

## Non-goals and guardrails

- [ ] Do not make `<pre>` interpret nested markup as rich LTML content.
- [ ] Do not change paragraph or wrapping behavior; `<pre>` remains a
      non-wrapping preformatted text widget.
- [ ] Do not change the existing `fixed` font default unless required to
      preserve current behavior.
- [ ] Do not introduce a new asset-loading system for `<pre>`.
- [ ] Do not change unrelated component-backed tag behavior while adding this
      support.

## Implementation checklist

### Phase 1: Refactor `StdPre` to use component-style content ownership

- [ ] Change `StdPre` to embed `StdComponent` instead of owning only
      `StdWidget` text state.
- [ ] Preserve widget/layout behavior already provided through the embedded
      widget/container chain.
- [ ] Remove direct dependence on incremental `AddText`-only storage as the
      source of truth.
- [ ] Keep `StdPre` registered as the built-in `pre` tag.

### Phase 2: Preserve existing inline `<pre>` semantics

- [ ] Keep inline `<pre>` content behaving as plain text, not raw XML markup.
- [ ] Ensure inline entity decoding still yields literal text in `Lines()`.
- [ ] Preserve current normalization rules:
  - [ ] CRLF / CR normalization.
  - [ ] Tab expansion to four spaces.
  - [ ] Surrounding blank-line trim.
  - [ ] Common-indent dedent for non-blank lines.
  - [ ] Blank-line preservation within the body.
- [ ] Keep `AccessibilityText()` based on the normalized resolved lines.
- [ ] Keep `PreferredWidth()` based on the normalized resolved lines.
- [ ] Keep `PreferredHeight()` based on the normalized resolved lines.
- [ ] Keep `DrawContent()` based on the normalized resolved lines.

### Phase 3: Add `src` support through component behavior

- [ ] Accept optional `src` on `<pre>`.
- [ ] Reuse `StdComponent.SetAttrs(...)` for `src` parsing.
- [ ] Reuse `StdComponent.Body()` / `assetSource()` behavior instead of
      duplicating path-resolution logic.
- [ ] Make `src` take precedence over inline content.
- [ ] Support asset FS resolution via `WithAssetFS(...)`.
- [ ] Support relative file resolution from `ParseFile(...)`.
- [ ] Support optional network sources using the same document/network rules as
      existing component-backed tags.
- [ ] Keep lazy loading semantics so external content is read on use, not at
      parse time.

### Phase 4: Parser and interface alignment

- [ ] Decide whether `StdPre` should continue implementing `HasText`, or
      whether parse-time body capture should become the single source path.
- [ ] If `HasText` remains, ensure it feeds the same normalization and caching
      pipeline as `src`-backed content.
- [ ] If parse-time body capture becomes primary, add a pre-specific body
      resolver or setter that preserves current text semantics instead of
      exposing raw XML behavior.
- [ ] Ensure `StdPre` still integrates correctly with parse stack push/pop
      behavior.
- [ ] Ensure sibling parsing around `<pre>` is unchanged.
- [ ] Ensure no regression in current parse behavior for inline `<pre>` bodies.

### Phase 5: Documentation updates

- [ ] Update `ltml/SYNTAX.md` to add `src` to the `<pre>` attribute table.
- [ ] Document that `src` wins over inline content.
- [ ] Document that `<pre>` uses the same source-resolution rules as
      component-backed tags.
- [ ] Add or update an LTML sample if that improves future verification or
      discoverability.

## Test checklist

### Unit tests

- [ ] Keep existing `ltml/std_pre_test.go` line-formatting coverage passing.
- [ ] Add a test for `<pre src="...">` loading from `WithAssetFS(...)`.
- [ ] Add a test for relative file resolution via `ParseFile(...)`.
- [ ] Add a test proving `src` overrides inline body.
- [ ] Add a test proving inline entity decoding still matches current behavior
      after the refactor.
- [ ] Add a test for missing or empty body behavior staying consistent with
      current rendering and layout expectations.

### Integration and focused verification

- [ ] Run focused LTML tests for `StdPre` behavior.
- [ ] Run focused LTML tests for component source-loading behavior.
- [ ] Run `go build ./...`.
- [ ] Run `go test ./...` when implementation happens.
- [ ] Note in the implementation PR or follow-up doc that some existing tests
      may require environment or network allowances outside the default sandbox.

## Suggested implementation sequence

- [ ] Land the `StdPre` internal refactor first, with no public behavior change.
- [ ] Add the shared resolved-body path so inline and `src` content feed the
      same normalization logic.
- [ ] Add `src` support and precedence behavior.
- [ ] Add tests for asset FS, relative file resolution, and precedence.
- [ ] Update `ltml/SYNTAX.md` and any sample coverage last.

This sequencing keeps the refactor reviewable and makes regressions easier to
localize.

## Acceptance criteria

- [ ] Existing inline `<pre>` documents render exactly as before.
- [ ] `<pre src="...">` works with asset FS-backed content.
- [ ] `<pre src="...">` works with relative local files when parsed through
      `ParseFile(...)`.
- [ ] `<pre>` follows existing component `src` precedence rules.
- [ ] `<pre>` follows existing component path and URL resolution rules.
- [ ] No regressions are introduced for `<svg>` or custom component source
      loading.
- [ ] `ltml/SYNTAX.md` accurately documents the new `<pre>` behavior.

## Assumptions and defaults

- Inline `<pre>` should preserve current text semantics rather than exposing raw
  inner XML.
- `src` should win when both `src` and inline content are present.
- The implementation should reuse `StdComponent` behavior rather than creating
  a parallel source-loading path for `<pre>`.
- `<pre>` should remain a text-oriented leaf widget even if its content source
  becomes component-backed internally.
