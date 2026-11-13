# PriceHike

A desktop application for applying configurable price rules to CSV and Excel datasets. It uses exact decimal arithmetic and lets you preview changes before exporting a transformed file.

## Features

- Load CSV, XLS, or XLSX files and inspect their item and price columns.
- Define and reuse rules that select rows by item-number wildcard or price thresholds.
- Add or subtract fixed amounts or percentages, round prices up or down, and control decimal precision.
- Preview affected rows and per-rule counts before exporting.
- Export the result as CSV or XLSX; import and export rule sets.
- Open a dataset by dropping a file onto the desktop window.

The application is built with Go and [Wails](https://wails.io/), with a Svelte 5 and TypeScript frontend.

## Development

Requirements: Go 1.24 or newer, [Wails v2](https://wails.io/docs/gettingstarted/installation), and Bun for the frontend dependencies. Platform-specific Wails prerequisites are listed in the Wails installation guide.

From the repository root:

```sh
wails dev
```

To create a production build:

```sh
wails build
```

See the root `wails.json` for the frontend build commands.
