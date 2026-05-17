# PriceHike — Frontend component inventory

This document is the **bridge from mockup to Svelte**. For every visual component in the static HTML, it lists: where it appears, what variants it has, the props/state it'll need when ported to Svelte 5, and which `HANDOFF.md` types it touches.

**Stack target:** Svelte 5 (runes mode) · plain Tailwind utility classes · CSS custom properties from `tokens.css` (no library lock-in).

**File map**
- `index.html` — preview-ready state (state #4) at full size
- `states.html` — all 7 application states stacked
- `tokens.css` — colors, type, spacing, radii, shadows · light + dark
- `HANDOFF.md` — backend contract (read first if porting)

---

## Component list

### Toolbar

**Where it appears:** every state.

**Variants**
| Variant | Trigger |
|---|---|
| `empty` | no dataset loaded; only Import enabled, all other buttons disabled |
| `loaded` | dataset present; all buttons enabled |
| `loading` | preview in flight; Export* and Preview disabled, Preview shows spinner |
| `dirty` | (future) rules changed since last preview; Preview button gets `is-dirty` accent ring |

**Slots**: brand · file metadata badge · spacer · action button group

**Props (Svelte)**
```ts
type ToolbarProps = {
  dataset: Dataset | null;          // for the file badge
  busy: boolean;                    // disables Preview/Export, shows spinner on primary
  onImport():       void;
  onExportCsv():    void;
  onExportXlsx():   void;
  onImportRules():  void;
  onExportRules():  void;
  onPreview():      void;
  hasRules: boolean;                // disables Export* if no rules and no dataset
};
```

**Appears in states:** 1, 2, 3, 4, 5, 6, 7

---

### FileMetaBadge

The small monospace pill in the toolbar showing `filename · N rows · M price cols`.

**Variants**: `present` | `absent` (hidden in state 1).

**Props**
```ts
type FileMetaProps = { dataset: Dataset };
```

**Touches:** `Dataset.filename`, `Dataset.items.length`, `Dataset.priceHeaders.length`

---

### RuleList

The scrollable left rail of rule cards.

**Variants**
| Variant | Notes |
|---|---|
| `empty` | shows "No rules yet" centered placeholder + Add CTA |
| `populated` | renders `<RuleCard>` per rule + a sticky-feeling "Add rule" pseudo-row at the bottom |
| `dragging` | (future) ghost row during drag-reorder |

**Props**
```ts
type RuleListProps = {
  rules: Rule[];                                  // execution order
  perRuleStats: RuleStat[];                       // aligned by index, may be empty before first preview
  selectedId: string | null;
  onSelect(id: string): void;
  onReorder(fromIdx: number, toIdx: number): void;
  onAdd(): void;
  onDelete(id: string): void;
};
```

**Appears in states:** 2 (empty), 3, 4, 5, 6, 7

---

### RuleCard

A single rule row with drag handle, ordinal, summary chips, and action buttons.

**Variants**
| Variant | Visual |
|---|---|
| `idle` | white surface, neutral chips |
| `hover` | surface-2 background, actions revealed |
| `selected` | accent-soft background, left accent border, actions visible |
| `warning` | warning-colored submeta line — "matched 0 items" (state #5) |
| `error` | negative-bg, negative-bordered chip on the offending field (state #7) |
| `compact` | smaller variant used when the rail is narrowed (states 4, 6) |

**Props**
```ts
type RuleCardProps = {
  rule: Rule;
  ordinal: number;       // 1-based, displayed as is
  stat?: RuleStat;       // null before first preview
  selected: boolean;
  onSelect(): void;
  onEdit(): void;
  onDelete(): void;
  dragHandleProps: object;  // from svelte-dnd-action or similar
};
```

**Summary line content (derived from `Rule`)** — chips + badge:
- Action badge: tone is positive (`add_*`), negative (`sub_*`), neutral (`round_*`)
- Filter chip(s): wildcard verbatim if non-empty; `<header> < <value>` for threshold
- Target column chips: one per `Rule.Columns` index, mapped through `Dataset.priceHeaders`; if `Columns.length === 0`, render a single `all columns` chip
- Decimals: `Keep` → "keep"; otherwise `Ndp`

**Appears in states:** 3, 4, 5, 6, 7

---

### RuleEditor

The right-side form for the selected (or newly-added) rule.

**Variants**
| Variant | When |
|---|---|
| `empty-prompt` | no rule selected — shows the "No rule selected · Add rule" CTA card |
| `editing` | rule loaded; all controls visible |
| `editing-dirty` | unsaved changes — header shows `Unsaved` badge, footer enables Save |
| `editing-invalid` | client-side validation failed (e.g. negative percent in a `_pct` action); Save disabled |

**Composition** (each is its own component below):
- `<FilterFieldset>`
- `<ColumnSelector>`
- `<ActionPicker>`
- `<DecimalSettings>`

**Props**
```ts
type RuleEditorProps = {
  rule: Rule | null;
  priceHeaders: string[];        // for column labels and threshold dropdown
  onChange(next: Rule): void;    // fires on every keystroke (debounce in parent for preview)
  onSave(): void;
  onCancel(): void;
  onDelete(): void;
  dirty: boolean;
};
```

**Appears in states:** 2 (empty-prompt), 3 (editing), 5 (editing), 7 (editing-invalid)

---

### FilterFieldset

Two-up sub-form: wildcard input + optional threshold (column dropdown + `<` + value).

**Variants**: `threshold-off` (default) | `threshold-on`

**Props**
```ts
type FilterFieldsetProps = {
  filter: RuleFilter;
  priceHeaders: string[];
  matchedKeys?: string[];        // computed by backend after a preview run; first ~4 shown as hint
  onChange(next: RuleFilter): void;
};
```

**Validation rules surfaced inline**
- Wildcard pattern is shown verbatim with a small mono hint underneath listing up to 4 matched keys (or "0 matches" warning if zero).
- Threshold value field accepts only valid decimal strings; backend will reject malformed values.

**Appears in:** RuleEditor (states 3, 5, 7)

---

### ColumnSelector

The 5-up grid of checkbox chips, one per price column.

**Variants**: `none-selected` (= empty array = "all columns" per §4.4) | `some-selected`

**Props**
```ts
type ColumnSelectorProps = {
  priceHeaders: string[];        // labels and indices
  selected: number[];            // empty array = all
  onChange(next: number[]): void;
};
```

**Behavior**
- A small "All / None" toggle pair sits below the grid (not in the mockup — flag for v1.1).
- Empty selection visually treated as "all" via a faint helper line.

**Appears in:** RuleEditor (states 3, 7)

---

### ActionPicker

Segmented control over the six `RuleActionType` values + a value input that re-labels itself based on the chosen type.

**Variants**
| Action chosen | Value input |
|---|---|
| `add_pct` / `sub_pct` | number with `%` suffix |
| `add_fixed` / `sub_fixed` | number with currency-agnostic numeric input |
| `round_up` / `round_down` | value input **hidden** (action ignores value per §4.5); decimals fieldset becomes the primary control |

**Props**
```ts
type ActionPickerProps = {
  action: RuleAction;
  onChange(next: RuleAction): void;
};
```

**Appears in:** RuleEditor (states 3, 5, 7)

---

### DecimalSettings

Segmented `Keep | 0 | 2 | 3 | 4` plus a custom-N input.

**Variants**: `keep` (segmented `Keep` active, custom disabled) | `places` (one numeric option active, custom mirrors)

**Props**
```ts
type DecimalSettingsProps = {
  settings: DecimalSettings;
  onChange(next: DecimalSettings): void;
};
```

**Appears in:** RuleEditor (states 3, 7)

---

### PreviewTable

The dense `<table>` with sticky top + sticky first column.

**Variants**
| Variant | Driven by |
|---|---|
| `raw` | no rules — single header row (`item`, each price column), no diffs (state 2) |
| `diff` | rules present — grouped header (each price column spans `old / new / diff`) (state 4) |
| `empty-no-matches` | preview returned 0 rows (state 5) |
| `loading` | replaced by `<PreviewSkeleton>` overlay (state 6) |
| `stale` | dimmed, non-interactive — shown above a banner when last call errored (state 7) |

**Props**
```ts
type PreviewTableProps = {
  mode: 'raw' | 'diff';
  priceHeaders: string[];
  rows: PreviewRow[] | Item[];   // PreviewRow in diff mode, Item in raw mode
  stale?: boolean;
};
```

**Performance notes**
- Use a virtualizer (`svelte-virtual` or hand-rolled `IntersectionObserver` window) above ~500 rows. The backend caps `sampleSize` so the DOM never sees the full 100k-row dataset, but `raw` mode shows the full `items` array — virtualize that.
- All numeric cells get `font-variant-numeric: tabular-nums` via the `.num` class in `tokens.css`.
- Pre-formatted strings only; no `toFixed`, no `Intl.NumberFormat`.

**Appears in states:** 2 (raw), 4 (diff), 5 (empty), 6 (loading-overlay), 7 (stale)

---

### PreviewCell

A single `<td>` rendering an `(old, new, diff)` triple. In raw mode, only a single value cell.

**Variants**
| Variant | Trigger |
|---|---|
| `unchanged-muted` | `Changed == false`; muted text |
| `changed-positive` | `Diff` starts with `+` |
| `changed-negative` | `Diff` starts with `-` |
| `changed-zero` | `Diff == "0.00"` but `Changed == true` (precision-only change) |

**Props**
```ts
type PreviewCellProps = { cell: PreviewCell };
```

**Touches:** `PreviewCell.original`, `.newValue`, `.diff`, `.changed`. All pre-formatted strings.

**Appears in:** PreviewTable rows (state 4, 7-stale)

---

### PreviewSkeleton

The loading rows + overlay shown while `GeneratePreview` is in flight.

**Props**
```ts
type PreviewSkeletonProps = {
  progressLabel?: string;   // e.g. "pass 2 / 3 · 87,201 / 112,451 rows" (purely cosmetic — backend doesn't stream)
};
```

**Appears in:** state 6 only

---

### EmptyState

Generic centered icon + headline + body + actions card. Two specific instances:
- **DropZone** (state 1) — full-body, dashed-border card with format chips
- **NoMatches** (state 5) — preview-scoped, ∅ glyph, "Edit rule" / "View raw data" CTAs

**Props**
```ts
type EmptyStateProps = {
  variant: 'dropzone' | 'no-matches' | 'no-rule-selected';
  primaryAction?:   { label: string; onClick(): void };
  secondaryAction?: { label: string; onClick(): void };
};
```

**Appears in states:** 1 (dropzone), 2 (no-rule-selected in editor pane), 5 (no-matches)

---

### ErrorBanner

Inline banner shown above the preview table when `GeneratePreview` rejects. Per the brief, it must exist; the toast pattern is reserved for `Export*` confirmations.

**Variants**: `error` (red) | `warning` (amber, for soft failures like "matched 0 items" if surfaced at top level — not used in the current 7 states).

**Props**
```ts
type ErrorBannerProps = {
  title: string;            // e.g. "Preview failed"
  detail: string;           // verbatim Go error message
  ruleOrdinal?: number;     // when the message starts with "rule #N: ...", parsed out by parent
  onEditRule?(): void;
  onDismiss(): void;
};
```

**Appears in:** state 7

---

### ThemeToggle

Small icon button in the toolbar that flips `[data-theme]` on `<html>`.

**Props**: none (reads/writes `localStorage["pricehike.theme"]`).

**Appears in:** every state (top-right of the toolbar).

---

### Badge

Generic uppercase pill from `tokens.css`. Variants by tone: `neutral | accent | positive | negative | warning`. Used throughout: rule action badges, status badges in the preview header, state labels in `states.html`.

### Chip

Mono "tag" pill (lower case, with monospace font) used in rule summaries for wildcards, column names, and threshold expressions.

---

## State × component matrix

| Component         | 1 Empty | 2 Loaded | 3 Editing | 4 Preview | 5 No-match | 6 Loading | 7 Error |
|---|:-:|:-:|:-:|:-:|:-:|:-:|:-:|
| Toolbar           | ✓ disabled | ✓ | ✓ | ✓ | ✓ | ✓ busy | ✓ |
| FileMetaBadge     | — | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| RuleList          | — | empty | ✓ | ✓ | ✓ | ✓ | ✓ |
| RuleCard          | — | — | ✓ selected | ✓ | ✓ warning | ✓ | ✓ error |
| RuleEditor        | — | empty-prompt | editing | — | editing | — | editing-invalid |
| PreviewTable      | — | raw | — (collapsed) | diff | — (no-matches state) | skeleton | stale |
| PreviewSkeleton   | — | — | — | — | — | ✓ | — |
| EmptyState        | dropzone | no-rule-selected | — | — | no-matches | — | — |
| ErrorBanner       | — | — | — | — | — | — | ✓ |
| ThemeToggle       | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

---

## Implementation order suggestion

1. `tokens.css` + `Badge` + `Chip` + `Button` styles — visual primitives
2. `Toolbar` + `FileMetaBadge` (driven by mocked `Dataset`)
3. `PreviewTable` raw mode against mocked `Dataset.items`
4. `RuleList` + `RuleCard` against a `Rule[]` literal
5. `RuleEditor` + sub-components, wired to the same `Rule[]`
6. Wire `GeneratePreview` mock returning hand-rolled `PreviewResult`
7. Swap mock for real Wails bindings; add `ErrorBanner` + `PreviewSkeleton`
8. Drag-reorder (`svelte-dnd-action` or similar) on `RuleList`
9. Theme persistence + `ThemeToggle`

Out of scope for v1, do not build: onboarding, settings panel, auth, saved projects/sessions, undo history, multi-file open, cell-level edits in the preview.
