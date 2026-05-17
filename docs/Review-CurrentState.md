# Review — Current state of the app

- **Date:** 2026-05-17
- **Scope:** Full code surface — `backend/`, `app.go`, `main.go`, and the Svelte frontend — against the state captured in `docs/Review-BackendServices.md` from earlier today.
- **Verdict in one line:** the services split has held up; H1 from the prior review is genuinely resolved (the `state.Store` surface is no longer reachable from JS), the dropdown is fixed in the frontend, and the rest of the architecture is in good shape — but the JS binding surface still leaks three Go-internal entrypoints (`Attach`, `AttachDialogContext`, `SnapshotSettings`), `main.go` still carries unresolved duplication and a missing comment, and one frontend-visible contract (the `"cancelled"` sentinel) is still a stringly-typed cross-language coupling worth tightening.

---

## What's already been actioned since the last review

Verified from the current files. Keep these as-is.

| Item | Where | Status |
|---|---|---|
| H1 — `*state.Store` no longer embedded on `Backend` | `backend/backend.go:9-16` | ✓ Fixed. `Clear`/`Filepath`/`Get`/`IsLoaded`/`Set` are gone from `Backend.d.ts`. |
| H4 — dead `Backend.ctx` field | `backend/backend.go` | ✓ Fixed. Field is gone. |
| Dropdown fix (codepoint → char) | `frontend/src/lib/backend/index.ts:151-162` | ✓ Fixed. `getCSVDelimiters` returns `CSVDelimiter[]` of `{name, value}` pairs. |
| M8 (from `Cleanup-Backend.md`) — idiomatic `defer` in Excel writer | `backend/io/excel/writer.go:13-15` | ✓ Fixed. Now `defer func() { _ = f.Close() }()`. |
| Preview infobar | `frontend/src/routes/+page.svelte:107-133` | ✓ Implemented and rendering. |
| Re-read file on delimiter change | `frontend/src/lib/state/app.svelte.ts:232-260` | ✓ Implemented, with graceful fallback when path is unknown. |

These all stay. No regressions worth flagging.

---

## 🔴 Must fix — public surface still leaks

### M-1 — Three `Backend` methods are still bound to JS that shouldn't be

Current `frontend/src/lib/wailsjs/go/backend/Backend.d.ts`:

```ts
export function Attach(arg1:context.Context):Promise<void>;             // ← M4 carry-over
export function AttachDialogContext(arg1:context.Context):Promise<void>;// ← H2 carry-over
export function SnapshotSettings():Promise<domain.Settings>;            // ← NEW LEAK
```

All three are Go-internal wiring that the frontend has no business calling.

- **`Attach`** (`backend/backend.go:40`) is the constructor's "give me the Wails ctx" hook. JS calling it with `null` or any reachable value will silently rewrite `DialogService.ctx`, after which every file dialog will fail.
- **`AttachDialogContext`** (`backend/services/dialog.go:18-20`) is the inner half of the same problem.
- **`SnapshotSettings`** (`backend/services/settings.go:36-40`) is a new leak that the prior review didn't flag, exposed by the same root cause as H1: `Backend` embeds `*services.SettingsService`, so *every* exported method on it gets bound — including `SnapshotSettings`, which was intended as an internal `SettingsProvider` interface method for `DatasetService`. Functionally identical to `GetSettings` today, but it's an accidental public API: any frontend code that adopts it will be hard to remove later.

**Fix — adopt the free-function pattern from the prior review (H2 Option A) and apply it consistently.** The same package-private helper trick works for all three:

```go
// backend/services/dialog.go
func InitDialogContext(s *DialogService, ctx context.Context) { s.ctx = ctx }

// backend/services/settings.go — make SnapshotSettings package-private:
func (s *SettingsService) snapshotSettings() domain.Settings { … }      // unexported
// Then expose a free function for the cross-package call site:
func SnapshotSettings(s *SettingsService) domain.Settings { return s.snapshotSettings() }
```

`DatasetService` currently calls `settingsProvider.SnapshotSettings()` through the `SettingsProvider` interface (`backend/services/dataset.go:15-17`). Two ways to keep that working:

- **Option A (cheapest):** rename the interface method to a lowercase one — won't work, Go interfaces can't have unexported methods across packages. So:
- **Option B:** keep the `SettingsProvider` interface but make the concrete implementation an *adapter* type defined inside `dataset.go` that wraps `*SettingsService` and calls the package-private helper. The adapter is only constructed in `backend.New()`. JS never sees the helper.

Both `Attach` and `AttachDialogContext` are simpler: convert them to package-level free functions (`services.InitDialogContext(s, ctx)` and a top-level `backend.attachContext(b, ctx)` called from `app.startup`). Wails ignores free functions.

After the fix, the regenerated `Backend.d.ts` should show only the eight methods that are legitimately part of the JS contract:
`ApplyAndExportCSV`, `ApplyAndExportXLSX`, `ExportRules`, `GeneratePreview`, `GetCSVDelimiters`, `GetSettings`, `ImportRules`, `LoadFile`, `SelectExportPath`, `SelectImportFile`, `SetSettings`.

### M-2 — Dead field `Backend.dialogService` is still there

`backend/backend.go:15`:

```go
type Backend struct {
    *services.DatasetService
    *services.RulesService
    *services.SettingsService
    *services.DialogService

    dialogService *services.DialogService   // ← never assigned, never read
}
```

This is H3 from the prior review. Three-line delete. Nothing depends on it — the embedded `*services.DialogService` is what `New()` populates and what every caller hits via method promotion. Removing it doesn't change behavior; leaving it just keeps the misleading "we have two of these" signal in the type definition.

---

## 🟠 Should fix — correctness traps and contract fragility

### S-1 — Drop event still doesn't carry the full path

`app.go:48`:

```go
runtime.EventsEmit(ctx, "file:dropped", filepath.Base(path), ds)
```

The frontend acknowledges this gap in two places and falls back to a "Re-import to apply" notice (`app.svelte.ts:93-100` and `:255-260`). The fix is the three-arg version from the prior review's addendum:

```go
runtime.EventsEmit(ctx, "file:dropped", filepath.Base(path), path, ds)
```

Plus the matching frontend update in `onFileDropped` (`backend/index.ts:165-171`) to read `args[1]` as the path, and `app.svelte.ts:93-100` to pass that path to `acceptLoaded` so `applySettings` can re-read dropped files too. Five lines on each side; closes the last fallback path. Worth doing now while the rest of the leak fixes are open — same release, same handoff.

### S-2 — `main.go` duplicated drag/drop config and missing comment

Two carry-overs from the prior review, both still present:

- **M2:** No comment on `DisableWebViewDrop: false` (`main.go:41`). The pairing — `DisableWebViewDrop: false` + `runtime.OnFileDrop` + the frontend's `preventDefault` — is load-bearing on GTK. The next reader will look at `false` and wonder why it isn't `true`; the comment from the prior review (4 lines) prevents that.
- **M3:** `CSSDragProperty` and `CSSDragValue` are set both at the top level (`main.go:28-29`) and nested inside `DragAndDrop` (`:42-43`). They're not in the canonical `options.App` struct in Wails v2 — they live under `DragAndDrop`. The top-level pair is either silently ignored or doing nothing useful. Pick one and drop the other.

### S-3 — `"cancelled"` is a stringly-typed sentinel across the Go/JS boundary

`backend/services/dialog.go:36`:

```go
if path == "" {
    return "", fmt.Errorf("cancelled")
}
```

…matched in the frontend four times:

```ts
if (msg !== 'cancelled') this.setError(msg);    // app.svelte.ts:122, 170, 319, 349
```

If anyone ever capitalises it, translates it, or wraps it with a prefix, every cancel will be reported to the user as an error. Two clean fixes:

- **Cheapest:** stop returning an error for "user cancelled". Return `("", nil)` and let the frontend check for empty path. No sentinel, no string comparison.
- **Or:** define a `domain.ErrUserCancelled = errors.New("user cancelled")` and return it; the frontend still matches a string, but at least the Go side has one canonical place to change it.

Recommend the first. `SelectExportPath` (`:54-58`) already returns `("", err)` without wrapping the cancellation as an error, so the pattern is inconsistent within the file too.

### S-4 — Stale doc reference in code

`frontend/src/lib/state/app.svelte.ts:97`:

```ts
// Tracked in docs/Cleanup-Backend.md as a small drop-event payload tweak.
```

That file doesn't exist any more (the work folded into `Review-BackendServices.md`). Either point the comment at the current review or delete the trailing sentence — the comment above it explains the situation fine on its own.

### S-5 — `state.Store.Clear()` is unused now

`backend/state/state.go:53-58`. With H1 done, `Clear` is no longer reachable from JS, and `grep` finds no Go caller either. Either remove it (keep the interface surface honest) or, if you anticipate a "reset workspace" feature soon, add a TODO comment naming the planned use — silent unused code is the bad option.

### S-6 — Filter-column range checks live in the wrong place

`backend/engine/engine.go:16-25` does explicit out-of-range checks for `rule.Filter.LessThan.Column` and `rule.Filter.GreaterThan.Column` *outside* `rule.Validate(columnCount)`. `Validate` itself (`backend/domain/rule.go:101-111`) only validates `r.Columns[]`. The two paths drift if anyone adds a new column-bearing field to a filter and only updates one site.

Move both filter-column checks into `rule.Validate`, so all column-range validation is a single function. The engine then just calls `rule.Validate(out.ColumnCount())` and trusts the result.

---

## 🟡 Consider — polish

### C-1 — `Filepath()` getter vs `path` field naming mismatch

`backend/state/state.go:11, 39`. M9 from `Cleanup-Backend.md`. Field is `path`, getter is `Filepath()`. Pick one. The getter is now only called from inside the package (since the binding was removed) so renaming it to `Path()` is the cheaper option.

### C-2 — Unused error constants in `domain/errors.go` (if M6 still applies)

The prior review noted `ErrColumnOutOfRange`, `ErrEmptyFile`, `ErrInvalidHeader` were declared but never returned. The current `backend/domain/errors.go` no longer declares them at all — looks like that pass already happened. Strike from the open list if so; flagged here only because the prior review still tracked it as open.

### C-3 — `SettingsService.NewSettingsService` explicitly sets the mutex zero value

`backend/services/settings.go:13-18`:

```go
return &SettingsService{
    mu:       sync.RWMutex{},
    settings: domain.DefaultSettings(),
}
```

The `mu: sync.RWMutex{}` line is harmless but redundant — Go zero-initialises the mutex anyway. Drop the line for consistency with `state.New()` which doesn't set the mutex explicitly.

### C-4 — `ApplyDecimalSettings` default silently returns the price unchanged

`backend/domain/format.go:18-20`. M4 from the cleanup doc, still open. An unknown mode silently returns the price as-is rather than surfacing a programmer error. Low impact because `ValidateShape` rejects unknown modes at the boundary, but a `panic("unknown decimal mode: " + string(settings.Mode))` would catch the case where someone adds a new mode and forgets to update `ApplyDecimalSettings`. Pure defensive-coding choice; leave it if you prefer permissive defaults.

### C-5 — `ValidateShape` action checks aren't symmetric

`backend/domain/rule.go:72-87`. Only `ActionSubtractPercent` validates that the value is within `(0, 100]`. `ActionAddPercent`, `ActionAddFixed`, `ActionSubtractFixed` only get the `IsZero` check at the end. The frontend can submit, say, an `addPercent` rule with value `-50` (which the engine will execute as a subtraction). Worth deciding once whether negative values are allowed for additive percent/fixed actions and either documenting "value is signed; sign determines direction" or validating in `ValidateShape`. Either is fine; the current half-implementation is the part to fix.

---

## What should stay exactly as is

These are the parts I'd resist changing in the next pass:

- **The services split itself** — `Backend` as a façade over four small services with a private store, the `SettingsProvider` / `DataStore` interface boundary on `DatasetService`, and the dependency wiring in `New()`. Nothing to do here.
- **`engine.ApplyRules` and `engine.Preview` shape** — the per-rule `RuleStat` tracking, the "matched 0 items" warning, and the sample-padding behavior (top up affected rows with leading unaffected rows up to sample size) all read like deliberate UX decisions. The preview infobar in `+page.svelte:107-133` depends on them being stable.
- **`Dataset.Clone()` deep copy** — three slice copies plus per-item slice copy. Verbose but correct, and the cost is bounded by `O(rows × columns)`. Don't try to make this clever.
- **The frontend `normalizeFilter`/`normalizeRule` helpers** (`backend/index.ts:65-80`) — they bridge the HANDOFF "empty filter means match all" UI model and the backend's required `applyToAll: true`. Removing them would push the conversion into every callsite. Keep.
- **The preview infobar markup and `shownAffected()` helper** (`+page.svelte:53-59, 107-133`). The "shown vs total" split is meaningful because the sample is capped at `DEFAULT_SAMPLE_SIZE = 500`; collapsing them would mislead on large datasets.
- **`Toolbar.svelte` / `FileMetaBadge.svelte` integration with the v2 dark-chrome tokens.** The structure (`.toolbar-brand .brand-text .name / .version`, no inline colour overrides) is the *correct* way to consume tokens; preserve it as the pattern for future toolbar widgets.
- **`Store.Set` cloning before storing** (`state/state.go:18-26`) and `Store.Get` cloning on the way out — defensive on both sides, prevents callers mutating the stored dataset by accident.
- **The Excel writer's `for c := range ds.PriceHeaders` row loop** (`io/excel/writer.go:34-43`). Implicitly pads short rows with zero `Decimal` cells, which closes the prior L5 concern about short-row handling. Don't replace with `range item.Prices` — you'd reintroduce the bug.

---

## Suggested fix order

Cheapest first, blast-radius aware. All of these can land in one branch and rebuild cleanly.

1. **M-2** — delete dead `dialogService` field (1 line).
2. **S-4** — fix stale doc reference comment (1 line).
3. **C-3** — drop redundant mutex init (1 line).
4. **S-2** — `main.go` comment + duplicate field cleanup (≈6 lines net).
5. **S-3** — drop the `"cancelled"` error; return `("", nil)` instead (≈4 lines across Go + frontend).
6. **S-6** — move filter-column checks into `rule.Validate` (≈15 lines moved, no logic change).
7. **S-1** — three-arg `file:dropped` event + matching frontend wiring (≈5 lines each side).
8. **M-1** — three-method JS surface leak, plus regenerate Wails bindings (the biggest chunk; do it after the small ones so the regenerated `.d.ts` diff is clean).
9. **C-1** — rename `Filepath()` → `Path()`.
10. **C-4, C-5** — only if you actually want the stricter behavior. Skip otherwise.

Build + vet should stay clean throughout. After step 8, eyeball the regenerated `Backend.d.ts` to confirm only the eleven legitimate methods are bound.

---

## What's deliberately not flagged

- **Vite DEP0205 warning** — still an upstream Node ESM hook issue; nothing in this codebase causes it. Same disposition as the prior review.
- **HANDOFF.md drift** vs. the backend contract — the `types.ts` shim handles the rename gap. Update the doc when convenient; the code works.
- **`docs/frontendHandoff/tweaks-panel.jsx`** — design-tool prototype, not project code. Leave it where it is.
- **The `decimal.Decimal` precision discipline** in `engine/preview.go`, `domain/format.go`, and the IO layers — I didn't re-audit it this pass because the prior reviews covered it. If you want a fresh look at the rounding semantics, that's its own review.

---

**Session summary**
- Mode: REVIEW
- Files written: `docs/Review-CurrentState.md`
- Files updated: (none)
- Next step: FEATURE — pick up the Must/Should items in the suggested order, starting with the cheap deletions before the JS-surface fix. Or, if you'd rather treat this as planning-only, the list above is ready to hand to whoever implements it.
