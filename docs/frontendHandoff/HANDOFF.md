# PriceHike — Backend Contract

**Stack:** Svelte 5 + Wails v2 + Go (binary money via `github.com/shopspring/decimal`)
**Scope:** every API call, type, and behavior the frontend depends on. Implement against this and the first integration should be green.

**Frontend assumption baseline:** the backend holds the loaded dataset in memory between calls; the frontend never passes it back. All decimal math happens server-side; numerics arrive as pre-formatted strings.

---

## 1. Wails API surface

All methods are bound from a single `App` struct (`app.Bind(NewApp())`). They surface in JS as `window.go.main.App.<Method>(...)` and return `Promise`. Errors thrown by Go reach JS as rejected promises whose `.message` is the Go `error.Error()` string.

### 1.1 `LoadFile(path string) (*Dataset, error)`

Reads CSV/XLSX/XLS at `path`, parses, and **caches the dataset in the App** for later calls.

- **path**: absolute filesystem path returned from `SelectImportFile` (frontend never types paths).
- **Returns**: full `Dataset` value, including all `Items`. The frontend may render the entire dataset (e.g. the "raw data" preview when no rules are defined) so do **not** truncate.
- **Errors**:
  - `"file not found: <path>"`
  - `"unsupported extension: <ext>"` — anything outside `.csv .xlsx .xls`
  - `"parse error at row N: <reason>"` — malformed CSV row or unparseable cell
  - `"no price columns detected"` — first column was non-numeric (article number) but no further numeric columns followed
- **Side effect**: replaces any previously-loaded dataset. Any prior `LastPreview` cache is cleared.

### 1.2 `GeneratePreview(rules []Rule, sampleSize int) (*PreviewResult, error)`

Applies `rules` sequentially against the cached dataset, returning a sampled set of affected rows.

- **rules**: array order **is** execution order. An empty array means "no transformations" — the result still returns up to `sampleSize` rows of pure passthrough (matching state #2).
- **sampleSize**: maximum number of `PreviewRow` entries to return. If `<= 0`, default to `500`.
- **Returns**: `PreviewResult` with `Rows` capped at `sampleSize` and totals reflecting the **full** computation (not the sample).
- **Sampling priority**: rows where **any cell changed** are emitted first; if fewer than `sampleSize` rows changed, fill the remainder with leading unaffected rows so the table is never empty for a loaded dataset.
- **Errors**:
  - `"no dataset loaded"` — `LoadFile` was never called
  - `"rule #N: <detail>"` — any per-rule validation error. The `#N` prefix is 1-indexed and is what the UI displays on the error banner / per-card warning.

### 1.3 `ApplyAndExportCSV(rules []Rule, destPath string) (*ExportReport, error)`

Re-applies `rules` against the **full** cached dataset (not a sample) and writes the result as UTF-8 CSV to `destPath`. Returns a small report for toast confirmation.

- **destPath**: from `SelectExportPath`; will be overwritten if it exists.
- Header row preserved (article-number column header + each `PriceHeaders` entry, unchanged).
- **Errors**: as `GeneratePreview` plus `"write failed: <reason>"`.

### 1.4 `ApplyAndExportXLSX(rules []Rule, destPath string) (*ExportReport, error)`

Same as `ApplyAndExportCSV` but produces an `.xlsx` workbook with one sheet named `prices`. Number cells are written as **strings** (no Excel-side recalculation); the frontend depends on the precision the rules dictate, not what Excel guesses.

### 1.5 `ExportRules(rules []Rule, destPath string) error`

Serializes `rules` to indented JSON and writes to `destPath`. Schema is the same as the input to `GeneratePreview` (see §3.4).

### 1.6 `ImportRules(path string) ([]Rule, error)`

Reads a `.json` file and returns the parsed `[]Rule`. Validates strictly:
- Unknown `RuleActionType` → `"unknown action: <value>"`
- Negative or zero `DecimalSettings.Places` outside `[0..8]` → `"invalid decimal places: N"`
- Reference to a column index not present in the *currently loaded* dataset is **allowed** here — the frontend can show the rule with a per-card warning — but `GeneratePreview` will fail on it.

### 1.7 `SelectImportFile() (string, error)`

Opens the OS-native open-file dialog filtered to `*.csv;*.xlsx;*.xls`. Returns the chosen path, or `""` and a non-nil error of `"cancelled"` if the user dismissed the dialog. The frontend treats `"cancelled"` as a soft signal (no toast).

### 1.8 `SelectExportPath(defaultName string, kind string) (string, error)`

Opens the OS-native save-file dialog. `kind` is one of `"csv" | "xlsx" | "rules-json"` and determines the filter + default extension. `defaultName` is the suggested filename (without path).

---

## 2. Type definitions

All types are in package `main`. JSON tags shown are the frontend-visible names.

```go
type Dataset struct {
    Filename     string   `json:"filename"`     // basename of loaded file
    Path         string   `json:"path"`         // absolute path on disk
    KeyHeader    string   `json:"keyHeader"`    // header of the first (article-number) column, e.g. "art.nr"
    PriceHeaders []string `json:"priceHeaders"` // headers for each price column, in source order, e.g. ["vk-0","vk-1",...]
    Items        []Item   `json:"items"`        // every row from the source file
    LoadedAt     string   `json:"loadedAt"`     // RFC 3339 timestamp
}

type Item struct {
    Key    string            `json:"key"`    // value of the article-number column (e.g. "AlphaBrush")
    Prices []decimal.Decimal `json:"prices"` // one entry per Dataset.PriceHeaders entry, same index
}

// ---------- Rules ----------

type Rule struct {
    ID       string          `json:"id"`        // frontend-generated UUID; backend treats as opaque
    Filter   RuleFilter      `json:"filter"`
    Columns  []int           `json:"columns"`   // indices into Dataset.PriceHeaders; see §4.4 for empty-array semantics
    Action   RuleAction      `json:"action"`
    Decimals DecimalSettings `json:"decimals"`
}

type RuleFilter struct {
    Wildcard  string          `json:"wildcard,omitempty"`  // empty string = match all keys; see §4.6
    Threshold *LessThanFilter `json:"threshold,omitempty"` // omitted/null = no threshold
}

type LessThanFilter struct {
    ColumnIndex int             `json:"columnIndex"` // index into Dataset.PriceHeaders
    Value       decimal.Decimal `json:"value"`       // serialized as JSON string
}

type RuleAction struct {
    Type  RuleActionType  `json:"type"`
    Value decimal.Decimal `json:"value"` // ignored for RuleRoundUp / RuleRoundDown — see §4.5
}

type RuleActionType string

const (
    RuleAddPercent RuleActionType = "add_pct"   // new = old * (1 + value/100)
    RuleSubPercent RuleActionType = "sub_pct"   // new = old * (1 - value/100)
    RuleAddFixed   RuleActionType = "add_fixed" // new = old + value
    RuleSubFixed   RuleActionType = "sub_fixed" // new = old - value
    RuleRoundUp    RuleActionType = "round_up"   // ceil to DecimalSettings precision
    RuleRoundDown  RuleActionType = "round_down" // floor to DecimalSettings precision
)

type DecimalSettings struct {
    Keep   bool `json:"keep"`            // if true, preserve the cell's incoming precision; Places ignored
    Places int  `json:"places,omitempty"` // 0..8, used only when Keep == false
}

// ---------- Preview ----------

type PreviewResult struct {
    TotalAffectedItems int          `json:"totalAffectedItems"` // distinct items where any cell changed (full dataset)
    TotalAffectedCells int          `json:"totalAffectedCells"`
    SampleSize         int          `json:"sampleSize"`         // echo of the requested cap
    PerRuleStats       []RuleStat   `json:"perRuleStats"`       // length == len(rules); aligned by index
    Rows               []PreviewRow `json:"rows"`               // already truncated to sampleSize
    GeneratedAt        string       `json:"generatedAt"`        // RFC 3339
}

type RuleStat struct {
    RuleID         string `json:"ruleId"`
    ItemsMatched   int    `json:"itemsMatched"`   // items the filter matched
    CellsChanged   int    `json:"cellsChanged"`   // cells whose value actually changed (action produced a different number)
    Warning        string `json:"warning,omitempty"` // e.g. "matched 0 items"; surfaced under the rule card
}

type PreviewRow struct {
    Key   string        `json:"key"`
    Cells []PreviewCell `json:"cells"` // one per Dataset.PriceHeaders entry; same index
}

type PreviewCell struct {
    Original string `json:"original"` // pre-formatted; see §3
    NewValue string `json:"newValue"` // pre-formatted; see §3
    Diff     string `json:"diff"`     // signed, pre-formatted: "+0.63" | "-1.25" | "0.00"
    Changed  bool   `json:"changed"`  // true iff Original != NewValue
}

// ---------- Exports ----------

type ExportReport struct {
    Path           string `json:"path"`
    ItemsWritten   int    `json:"itemsWritten"`
    BytesWritten   int64  `json:"bytesWritten"`
    DurationMillis int64  `json:"durationMillis"`
}
```

**Index alignment, in one sentence:** `Dataset.PriceHeaders[i]`, `Item.Prices[i]`, `PreviewRow.Cells[i]`, and `Rule.Columns` (as values) all index the same axis — the i-th price column. Keep this invariant religiously; the UI assumes it without checking.

---

## 3. Serialization notes

### 3.1 `shopspring/decimal` is JSON-stringified
`decimal.Decimal` marshals to and unmarshals from a JSON string. The frontend stores these as plain JS strings — no `parseFloat`, no `Number()`, no `Intl.NumberFormat`. Decimal math is **never** done on the client.

### 3.2 `PreviewCell.Original / NewValue / Diff` are pre-formatted strings
Format honors the **rule's** `DecimalSettings` that produced the cell's final value:
- If `Keep == true`, preserve the original incoming precision (count trailing zeros from the source data).
- If `Keep == false`, format to exactly `Places` decimal digits, zero-padded (e.g. `"5.20"`, not `"5.2"`).
- For passthrough cells (no rule affected them), use the original incoming precision verbatim.

### 3.3 `Diff` is always signed
Three forms only:
- `"+0.63"` — positive non-zero
- `"-1.25"` — negative non-zero
- `"0.00"` — zero (precision matches `NewValue`)

The frontend uses the leading character (`+`, `-`, or digit) to pick its color class; no parsing required.

### 3.4 Rule import/export JSON

Exactly the wire shape of `[]Rule`. Full example covering all variants:

```json
[
  {
    "id": "01J5VZ8K2NQ3F7M4XR9YH1WBPA",
    "filter": { "wildcard": "Alpha*" },
    "columns": [0, 1],
    "action":   { "type": "add_pct", "value": "5" },
    "decimals": { "keep": false, "places": 2 }
  },
  {
    "id": "01J5VZ8K2NQ3F7M4XR9YH1WBPB",
    "filter": {
      "wildcard":  "",
      "threshold": { "columnIndex": 2, "value": "10" }
    },
    "columns": [2],
    "action":   { "type": "sub_pct", "value": "2" },
    "decimals": { "keep": true }
  },
  {
    "id": "01J5VZ8K2NQ3F7M4XR9YH1WBPC",
    "filter": { "wildcard": "" },
    "columns": [],
    "action":   { "type": "round_down", "value": "0" },
    "decimals": { "keep": false, "places": 2 }
  }
]
```

Notes on the example:
- Rule 1: wildcard only, two target columns, +5% with 2-dp output.
- Rule 2: combined wildcard-all + threshold (`vk-2 < 10`), single target column, `Keep == true` (decimals.places omitted).
- Rule 3: no filter (match-all), `Columns: []` → all columns (see §4.4), action `round_down` whose `value` is meaningless (see §4.5) but still present.

---

## 4. Behavioral contracts

### 4.1 Execution order = array order
`rules[i]` runs after `rules[i-1]` against the running state of each cell. Reordering on the frontend = reordering the array on the next `GeneratePreview` call.

### 4.2 Dataset immutability
The cached `Dataset` is **never mutated** by rule application. Every `GeneratePreview` / `ApplyAndExport*` call rebuilds the working state from the original `Item.Prices`. The frontend depends on this — toggling rules off and rerunning must produce identical numbers.

### 4.3 Sampling priority
`GeneratePreview` returns affected rows first, then leading unaffected rows up to `sampleSize` — never the other way around. If `len(rules) == 0`, "affected" is empty and the first N rows of the dataset are returned (state #2).

`sampleSize <= 0` defaults to `500`. Frontend never relies on order beyond "affected first" — within those two buckets ordering follows the dataset's source row order (stable).

### 4.4 `Rule.Columns` empty-array semantics
**Empty `[]int{}` means "all columns".** A rule with `Columns: []` applies to every entry in `Dataset.PriceHeaders`. This is the only sane fit for the "Round down on all items, all columns" example rule, and the frontend renders an empty array as the `all columns` chip.

`nil` and `[]int{}` are treated identically. Indices out of range produce `"rule #N: column index out of range: K"`.

### 4.5 Rounding actions ignore `Action.Value`
For `round_up` and `round_down`, the precision comes from `DecimalSettings`. `Action.Value` may be any value (the frontend may send `"0"`) and must not affect output. Round semantics:
- `round_up`: `ceil(value × 10^places) / 10^places` when `Keep == false`; with `Keep == true`, use the cell's incoming precision.
- `round_down`: floor, same rule.

### 4.6 Wildcard syntax
Case-**sensitive**, **whole-string** match (anchored at both ends). Two metacharacters:
- `*` — zero or more characters
- `?` — exactly one character

Empty `Wildcard` (`""`) matches **every** key (i.e., "apply to all"). No regex syntax, no character classes, no escape — `*` and `?` are always metacharacters. (Article numbers in this dataset are alphanumeric; no need to escape them.)

### 4.7 Combined filter semantics
When both `Wildcard` and `Threshold` are present, both must match (logical AND). When neither produces a positive selector — `Wildcard == ""` and `Threshold == nil` — the rule applies to all items.

### 4.8 `LessThanFilter` against missing column
`GeneratePreview` returns `"rule #N: LessThanFilter references column \"<header>\" which does not exist in dataset (have <header-list>)"`. **Error, not skip.** This drives state #7. `ImportRules` does *not* validate this — it lets you load the rules so the user can fix them in the editor.

### 4.9 Filter that matches zero items
**Not an error.** `RuleStat.ItemsMatched == 0`, `Warning == "matched 0 items"`. The preview still completes; subsequent rules still run. State #5 surfaces this.

### 4.10 Diff zero with formatting drift
If a rule reduces precision (e.g. `Keep=false, places=0`) on a value that already had `0.00`, `Diff` is still `"0.00"` formatted to the new precision (`"0"` in that case). `Changed` reflects **string inequality** of `Original` vs `NewValue`, so a precision-only change with no numeric delta still counts as `Changed: true` and `Diff` is computed as the formatted difference (often `"0"` or `"0.00"`). The frontend colors it muted.

### 4.11 Export covers the full dataset
`ApplyAndExportCSV` / `XLSX` ignore `sampleSize` entirely. They apply rules to every row and write all rows.

---

## 5. Frontend expectations of the backend

These are the assumptions baked into the Svelte code. Honor them or expect breakage.

- **All Wails calls are async** (`Promise`-returning). UI optimistically locks affected buttons (Preview, Export*) until resolved.
- **Errors surface as rejected promises** whose `.message` is the Go error string verbatim. The frontend matches `"rule #N: ..."` to locate the offending rule card. Keep that prefix shape.
- **File dialogs are backend-driven.** The frontend never receives raw `File` objects, blobs, or paths from a browser file picker — it calls `SelectImportFile` / `SelectExportPath` and uses the returned path with `LoadFile` / `ApplyAndExport*`.
- **No streaming.** `GeneratePreview` returns the full sampled result in one call. For very large datasets, the frontend will show a determinate spinner using `RuleStat`-level granularity only after the call resolves — no progress events.
- **Backend holds dataset state.** Between `LoadFile` and the next `LoadFile`, the dataset is "the" dataset. The frontend does not send it back on every call.
- **No undo on the backend.** The frontend keeps `Rule[]` in component state; the backend is stateless about rule history.
- **Idempotency of `GeneratePreview`.** Calling it twice with the same `rules` against the same dataset must produce byte-identical `PreviewResult.Rows` (ignoring `GeneratedAt`). The frontend uses this for the "In sync" badge — it compares rule list hash + dataset hash against the last successful preview.
- **No client-side decimal math** anywhere — even for the per-rule "matched N items" summary, the frontend reads `RuleStat`.
- **Thread-safety.** The frontend may call `GeneratePreview` while the user is editing (debounced). Successive calls must not corrupt cached dataset state. A `sync.Mutex` on the App's compute path is sufficient.

---

## 6. Open questions (please confirm)

These are reasonable defaults the frontend has assumed; flag anything you'd change before integration.

1. **Numeric parsing locale.** CSV cells that use `,` as decimal separator (German export from Excel) — accept and normalize during `LoadFile`, or reject? Frontend assumes accept-and-normalize (output is always `.`).
2. **Article-number column detection.** Always the first column, regardless of header name? Frontend assumes yes.
3. **Threshold filter operator.** Only `<` for v1, no `<=` / `>` / `==`? Frontend currently exposes `<` only; locking that in.
4. **Round semantics on negative numbers.** Prices are non-negative in practice but if a `sub_fixed` action drives a value below zero, do we clamp at zero or let it through? Frontend renders whatever you send.
5. **Maximum `sampleSize`.** Should the backend cap this at, say, 5000 to keep IPC payloads bounded?
