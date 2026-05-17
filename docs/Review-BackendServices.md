# Review — Backend after the Services refactor

- **Date:** 2026-05-17
- **Scope:** full backend after the split into `backend/services/{dataset,dialog,rules,settings}.go`, the `app.go` + `main.go` drag-and-drop changes, and the rest of the codebase.
- **Verdict in one line:** the services split is a real improvement and `go build` / `go vet` are clean, but the `Backend` façade leaks the entire `state.Store` interface to the JS side (including `Set`, `Clear`, `Get`), the new `GetCSVDelimiters()` shape silently broke the frontend dropdown (already pinned down to the codepoint-as-number JSON issue), and a handful of dead fields / leftover lifecycle hooks need pruning.

---

## Architecture — what changed and why it works

- `Backend` is now a thin façade that **embeds** four service pointers (`DatasetService`, `RulesService`, `SettingsService`, `DialogService`) plus `*state.Store`. All those types' methods are promoted onto `Backend` and Wails picks them up automatically. The actual logic lives in each service file. ✓
- `DatasetService` accepts a `SettingsProvider` interface and a `DataStore` interface (`backend/services/dataset.go:15-22`). That keeps it independent of the concrete settings/store types and would let you swap either in a test. Good.
- `DialogService` is the only service that needs the Wails ctx; `AttachDialogContext` is set once in `Backend.Attach`. Other services are pure Go. Clean.
- `app.go` keeps the drag-and-drop pipeline (Wails-shell adapter, app-specific). `main.go` configures the window options. Both sit at the project root in `package main`, as before.

The `Backend` constructor in `backend/backend.go:20-39` wires the dependency graph explicitly:

```go
store := state.New()
dialogSvc := services.NewDialogService()
settingsSvc := services.NewSettingsService()
rulesSvc := services.NewRulesService()
datasetSvc := services.NewDatasetService(settingsSvc, store)
```

Readable, swappable, no globals. This is how the refactor should look.

---

## 🔴 High — exposed surface / dead code

### H1 — Embedding `*state.Store` exposes the *entire* store interface to JS

`backend/backend.go:9-18`:

```go
type Backend struct {
    *services.DatasetService
    *services.RulesService
    *services.SettingsService
    *state.Store           // ← all of these methods become JS-callable
    *services.DialogService
    ...
}
```

Because Wails binds every exported method on a bound struct (including methods promoted by embedding), the frontend can now call:

| Method | What it does |
|---|---|
| `Backend.Get()` | Returns the cached dataset — bypasses snapshot semantics |
| `Backend.Set(ds, path)` | **Overwrites the store from JS without going through `LoadFile`** |
| `Backend.Clear()` | Wipes the store from JS |
| `Backend.IsLoaded()` | Status query |
| `Backend.Filepath()` | Current path |

`Backend.Set` is the worrying one. From the frontend a buggy or malicious caller can replace the dataset with any shape it wants — skipping the CSV/Excel reader, skipping the extension sniff in `DatasetService.LoadFile`, skipping the path-tracking. `Clear()` similarly wipes the dataset out from under any in-flight operation.

You can confirm this by looking at the regenerated bindings:

```ts
// frontend/src/lib/wailsjs/go/backend/Backend.d.ts
export function Clear():Promise<void>;
export function Filepath():Promise<string>;
export function Get():Promise<domain.Dataset>;
export function IsLoaded():Promise<boolean>;
export function Set(arg1:domain.Dataset,arg2:string):Promise<void>;
```

All bound, all callable, all unintended.

**Fix — stop embedding `*state.Store` on `Backend`:**

```go
type Backend struct {
    *services.DatasetService
    *services.RulesService
    *services.SettingsService
    *services.DialogService

    ctx   context.Context
    store *state.Store        // ← unexported field, not embedded
}

func New() *Backend {
    store := state.New()
    ...
    return &Backend{
        store:           store,
        DatasetService:  datasetSvc,
        RulesService:    rulesSvc,
        SettingsService: settingsSvc,
        DialogService:   dialogSvc,
    }
}
```

`DatasetService` already takes the store via the `DataStore` interface (`services/dataset.go:19-22`), so the wiring doesn't change. The store stops being a public method-source on `Backend`.

After this, re-run `wails generate module` and confirm `Clear` / `Filepath` / `Get` / `IsLoaded` / `Set` disappear from `Backend.d.ts`.

### H2 — `DialogService.AttachDialogContext` is bound to JS

Same root cause as H1, narrower blast radius. `AttachDialogContext` (`services/dialog.go:18-20`) takes a `context.Context` from JS, where it can do exactly nothing useful — JS can't construct a real Wails ctx. But the binding is there:

```ts
export function AttachDialogContext(arg1:context.Context):Promise<void>;
```

If a frontend calls it with `null` or anything reachable from JS, it overwrites `s.ctx` and **every subsequent file dialog throws** because `runtime.OpenFileDialog(null, …)` panics or errors.

**Fix:** rename to an unexported helper and call it from `Backend.Attach` directly. Since `Backend` is in a different package from `services`, the unexported method needs a thin exported wrapper that's somehow not bound — which Wails doesn't allow. Two workable options:

- **Option A — pass the ctx to the constructor** *and* leave `Attach` as a top-up. The constructor is only called from `backend.New()`, never from JS. Move the ctx assignment to a non-method route: a package-level function `services.InitDialogContext(s *DialogService, ctx context.Context)` does the assignment. Methods on `s` aren't the only path to its private fields when the helper is in the same package.

   ```go
   // services/dialog.go
   func InitDialogContext(s *DialogService, ctx context.Context) { s.ctx = ctx }
   ```

   ```go
   // backend/backend.go:Attach
   services.InitDialogContext(b.DialogService, ctx)
   ```

   `InitDialogContext` is a free function, not a method on a bound struct — Wails ignores it.

- **Option B — document and live with it.** Wails-bound surface API hygiene is a known limitation; a comment on `AttachDialogContext` plus a top-level "do not call from JS" note in `Backend.d.ts` (via a doc comment on the method) at least warns the next reader.

Recommend Option A.

### H3 — `Backend.dialogService` is a dead duplicate field

`backend/backend.go:14, 17`:

```go
type Backend struct {
    ...
    *services.DialogService    // ← embedded (used)

    ctx           context.Context
    dialogService *services.DialogService   // ← duplicate, never assigned, never read
}
```

The embedded `*services.DialogService` is what `New()` populates and what every caller hits via method promotion. The unexported `dialogService` field is dead code from an earlier refactor. Remove it.

### H4 — `Backend.ctx` is dead too

Same struct. `ctx context.Context` (`backend.go:16`) is assigned in `Attach` (`backend.go:44`) but never read anywhere — `DialogService` carries its own ctx now. Remove.

After H3 + H4 the struct collapses to:

```go
type Backend struct {
    *services.DatasetService
    *services.RulesService
    *services.SettingsService
    *services.DialogService

    store *state.Store   // ← per H1, unexported
}
```

---

## 🟠 Medium — contract / shape issues

### M1 — `GetCSVDelimiters()` returns `map[string]rune` → reaches JS as `Record<string, number>`

`backend/services/settings.go:42-50`:

```go
func (s *SettingsService) GetCSVDelimiters() map[string]rune {
    return map[string]rune{
        "Comma":     ',',
        "Semicolon": ';',
        "Pipe":      '|',
        "Tab":       '\t',
        "Colon":     ':',
    }
}
```

`rune` is `int32`. When Go's `encoding/json` marshals a map value of type `int32`, it produces a **number**, not a string. The regenerated TS binding confirms:

```ts
export function GetCSVDelimiters():Promise<Record<string, number>>;
```

So at runtime the frontend receives `{Comma: 44, Semicolon: 59, Pipe: 124, Tab: 9, Colon: 58}` — labels mapped to Unicode codepoints.

**This is the dropdown bug.** The frontend's `getCSVDelimiters()` wrapper currently casts the result to `string[]` (a hold-over from when the backend returned `[]string`), and the SettingsDialog does `{#each labels as label}` — iterating a `string[]`. Iterating a plain object isn't iterable in JS, so the dropdown is empty (or the iteration throws, depending on how Svelte handles it).

The shape itself is fine — using codepoint numbers preserves the value alongside the label in a single map. Cleaner than the old "`[]string` of labels + frontend keeps its own value map" arrangement.

**No backend change needed** — this is fully a frontend fix. Convert `Record<string, number>` to a `CSVDelimiter[]` of `{name, value}` pairs by mapping `String.fromCodePoint(codepoint)`. The local `DELIMITER_VALUES` constant in `SettingsDialog.svelte` finally goes away. Done as part of this pass (see below).

(If you prefer string values on the wire: change the return type to `map[string]string` with the literal `","`, `";"`, etc. Either works; the codepoint version is arguably more honest about what a "delimiter" is, and avoids string-escape ambiguity.)

### M2 — Drag-and-drop pairing relies on the frontend's `preventDefault`

`main.go:39-44` keeps `DisableWebViewDrop: false`. The user note explains why: with `true`, drag-and-drop didn't work on the target platform at all. That's a real environmental constraint, not a regression.

The pairing that makes it work today:

- **`main.go`:** `DisableWebViewDrop: false` → the webview *also* receives the drop event.
- **`app.go:29`:** `runtime.OnFileDrop` fires alongside the webview's handler and emits `file:dropped` / `file:error`.
- **`frontend/src/routes/+layout.svelte:13-21`:** window-level `dragover` / `drop` `preventDefault` stops the webview from navigating to `file://...`.

All three pieces are now load-bearing. Worth a comment in `main.go`:

```go
// DisableWebViewDrop:false is required on GTK — `true` swallows the drop.
// The frontend's +layout.svelte cancels DOM drag defaults to stop the
// webview from navigating; without that, dropped files would open as a
// page inside the panel.
DisableWebViewDrop: false,
```

…so the next person who reads "shouldn't this be `true`?" doesn't flip it. The comment is the only fix needed.

### M3 — `main.go:28-29` has top-level `CSSDragProperty` + `CSSDragValue` *and* a nested copy inside `DragAndDrop`

```go
CSSDragProperty: "--wails-drop-target",   // ← top level
CSSDragValue:    "drop",
...
DragAndDrop: &options.DragAndDrop{
    ...
    CSSDropProperty: "--wails-drop-target",   // ← inside DragAndDrop
    CSSDropValue:    "drop",
},
```

I can't find `CSSDragProperty` / `CSSDragValue` in the Wails v2 `options.App` struct (the canonical names live inside `DragAndDrop`). The compiler accepts them probably because they were added in a later v2 minor; they may or may not do anything. The intent is duplicated either way.

**Either** remove the top-level pair (cleaner) **or** confirm they're a real `options.App` field on your Wails version and the duplication is intentional. `go vet` won't catch unused struct fields.

### M4 — `Backend.Attach` is still bound to JS

Same problem as H2 but for `Backend.Attach` itself (`backend.go:43`). The frontend has:

```ts
export function Attach(arg1:context.Context):Promise<void>;
```

Same fix family as H2 — pass ctx in via a free function and rename. Or live with it and document.

### M5 — Carry-overs from `Cleanup-Backend.md` are still open

Verified still present:

- **M4 (cleanup)** — `ApplyDecimalSettings` silent default on unknown mode (`domain/format.go:16-18`).
- **M6 (cleanup)** — `ErrColumnOutOfRange`, `ErrEmptyFile`, `ErrInvalidHeader` declared but never returned.
- **M7 (cleanup)** — now resolved as part of the services refactor (`DatasetService.exportWith` takes a writer function). ✓ Can be struck from the cleanup doc.
- **M8 (cleanup)** — `excelio/writer.go:12` still uses bare `defer f.Close()`; reader uses the idiomatic discarded form. Cosmetic.
- **M9 (cleanup)** — partially done: the store field is now named `path` (`state/state.go:11`), but the public getter is still `Filepath()` (`:39`). Either rename the getter to `Path()` to match, or rename the field back. Currently the names disagree.
- **L5 (cleanup)** — Excel writer doesn't pad short rows. Still open.

Nothing else from the old doc has been touched.

---

## 🟡 Low — polish

### L1 — Method ordering inside `backend.go` is fine

Three lines for the type, ten for `New()`, three for `Attach`. That's the entire file. No ordering issue at this size.

### L2 — The Vite deprecation warning is not your problem

```
(node:570352) [DEP0205] DeprecationWarning: `module.register()` is deprecated.
Use `module.registerHooks()` instead.
```

This is a Node.js DEP0205 warning fired from inside Vite (or one of its dev-mode dependencies — likely `@sveltejs/vite-plugin-svelte` or the Tailwind v4 plugin). It's an upstream issue that resolves when Vite catches up to the newer Node ESM hook API. Nothing in this codebase causes it; nothing in this codebase fixes it. Suppressible with `NODE_OPTIONS='--disable-warning=DEP0205'` if it's noisy; otherwise ignore.

### L3 — `state.New()` returns `&Store{}` with zero-value `mu` and `dataset`

Fine — Go zero-value for `sync.RWMutex` is a usable unlocked mutex. Pointer to `dataset` is nil, which `Get` checks. No bug, just worth knowing.

---

## What's deliberately not flagged

- The services split itself — the dependency injection is clean and the boundaries make sense.
- HANDOFF.md still describes a contract that doesn't match the backend (Dataset.Filename, Item.Key, Decimals field name, etc.) — known doc drift that the frontend's `types.ts` shim already handles. Update the doc when you have time; the code works.
- The new `docs/frontendHandoff/tweaks-panel.jsx` is a self-contained React+Babel prototyping helper for mockups — it's a generic design-tweaking shell, not project-specific code to integrate. The Svelte frontend doesn't need to consume it.

---

## Suggested fix order

Cheap-first, blast-radius-aware:

1. **Dropdown fix** (frontend) — implemented in this same pass. Detail in the next section.
2. **H1** — stop embedding `*state.Store` on `Backend`. Three-line struct change + one constructor field rename + `wails generate module`.
3. **H3 + H4** — delete the two dead fields (`dialogService`, `ctx`).
4. **M2** — one-line comment on `DisableWebViewDrop: false` in `main.go`.
5. **H2 + M4** — `AttachDialogContext` / `Attach` Go-internal helpers. Slightly fiddly because of the cross-package private-method issue; the `services.InitDialogContext` free-function pattern works.
6. **M3** — sort out the duplicate `CSSDragProperty` lines.
7. **M5** — pick off remaining `Cleanup-Backend.md` items at leisure.

Build + vet stay clean throughout.

---

## Frontend: what changes alongside this review

Just the dropdown. The current `backend/index.ts:getCSVDelimiters` claims `Promise<string[]>` and casts a `Record<string, number>` into it; the result is an object Svelte's `{#each}` can't iterate. Fix:

- `backend/index.ts:getCSVDelimiters` now reads the codepoint map and returns a `CSVDelimiter[]` (`{name, value}` pairs built via `String.fromCodePoint`).
- `components/SettingsDialog.svelte` drops its local `DELIMITER_VALUES` map and renders directly from the backend pairs.

Closes the silent label/value coupling that used to require a hand-maintained map on the frontend.

---

## Addendum (later in the same session) — drop event omits the full path

After the H1 cleanup (no more embedded `*state.Store`) `Backend.Filepath()` is gone from the JS bindings — correct, that's what H1 asked for. But now the frontend can't ask the backend where the current file lives, so the "re-parse the loaded file when the CSV delimiter changes" flow needs the frontend to track the path itself.

For picker-loaded files that works fine — `SelectImportFile` returns the path and the frontend stores it. For **drop-loaded files** the event payload only carries the basename (`app.go:48`):

```go
runtime.EventsEmit(ctx, "file:dropped", filepath.Base(path), ds)
```

The frontend therefore can't reload dropped files after a settings change — it falls back to a notice ("Re-import the file to apply the new CSV delimiter to the current dataset"). To close that gap:

```go
runtime.EventsEmit(ctx, "file:dropped", filepath.Base(path), path, ds)
```

Three positional args instead of two. Frontend follow-up: `onFileDropped((filename, dataset) => …)` becomes `onFileDropped((filename, path, dataset) => …)`, and `acceptLoaded` stores the new arg as `app.path`.

Five lines on each side; closes the last fallback path.

---

## Addendum — design refresh applied to the frontend

The user dropped a new `tokens.css` (v2, "bold") in `docs/frontendHandoff/` and asked for a richer Preview infobar plus automatic file reload on settings change. Done in this pass:

- `frontend/src/lib/styles/tokens.css` replaced verbatim with the v2 file. New palette is indigo (was deep teal), with selectable palettes via `[data-palette]` on the html element (violet / teal / emerald / amber / crimson).
- Toolbar moved to the **dark chrome** by default (per the v2 default `[data-chrome="dark"]` on body — the tokens default already covers this so no html/body change is needed). `frontend/src/lib/styles/app.css` updated with the new toolbar styles (`--toolbar-bg`, `--toolbar-fg`, `--toolbar-chip-bg`, the `.toolbar::after` accent gradient line, the `.toolbar .btn` overrides for dark chrome).
- `Toolbar.svelte` now uses the structured `.toolbar-brand .brand-text .name / .version` markup the new CSS expects; inline color styles removed so the dark-chrome tokens take over.
- `FileMetaBadge.svelte` strips its inline `t-tertiary` color so it inherits from `.toolbar-file .meta` (which knows to use the toolbar's tertiary token).
- The preview panel now renders the rich infobar when a preview exists, matching the mockup:
  > **Preview** · Showing **10** of **10** affected items · 3 rules applied sequentially · last run 15 May 2026 · 14:32:08 — *Preview is in sync*
  — built from `app.preview.totalAffectedItems`, the count of rows in the sample where any cell is `changed`, `app.rules.length`, and `formatGeneratedAt(app.preview.generatedAt)`. When no preview exists, falls back to the simple `panel-header` "Preview" label.
- `applySettings()` in `app.svelte.ts` now re-reads the loaded file when the CSV delimiter changed and the path is known, then shows an info notice. When the path is unknown (drop case, see N4 above) it shows a graceful "Re-import the file to apply" notice instead.
- The "Preview is in sync" badge is shown whenever a preview exists. It's reliable because every rule edit (`addRule` / `removeRule` / `updateRule` / `reorderRules`) clears `preview` to null, so a non-null preview is by definition synced to the current rules.
