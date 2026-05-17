// Plain wire types mirroring the backend's domain package.
//
// The Wails-generated bindings in $lib/wailsjs/go/models.ts export *classes*
// with helper methods (convertValues, createFrom, ...). Those are awkward to
// construct as plain object literals because TypeScript demands the methods
// exist. We never need the helpers — Wails serializes whatever shape we pass
// across the bridge as JSON — so define plain interfaces here and use them
// everywhere in the frontend instead.

export interface Item {
	itemNumber: string;
	prices: string[];
}

export interface Dataset {
	itemHeader: string;
	priceHeaders: string[];
	items: Item[];
	exportPrecision: number[];
}

export interface LessThanFilter {
	column: number;
	value: string;
}

export interface GreaterThanFilter {
	column: number;
	value: string;
}

export interface RuleFilter {
	applyToAll: boolean;
	wildcard?: string;
	lessThan?: LessThanFilter;
	greaterThan?: GreaterThanFilter;
}

export type ActionType =
	| 'addPercent'
	| 'subtractPercent'
	| 'addFixed'
	| 'subtractFixed'
	| 'roundUp'
	| 'roundDown';

export interface RuleAction {
	type: ActionType;
	value: string;
}

export type DecimalMode = 'keep' | 'fixed';

export interface DecimalSettings {
	mode: DecimalMode;
	places: number;
}

export interface Rule {
	id: string;
	filter: RuleFilter;
	columns: number[];
	action: RuleAction;
	decimalSettings: DecimalSettings;
}

// Backend now emits one PreviewCell per price column (dense), so the
// "Dense*" types the wrapper used to compute have been collapsed into the
// regular types.
export interface PreviewCell {
	column: number;
	original: string;
	newValue: string;
	diff: string;
	changed: boolean;
}

export interface PreviewRow {
	itemNumber: string;
	cells: PreviewCell[];
}

export interface RuleStat {
	ruleId: string;
	itemsMatched: number;
	cellsChanged: number;
	warning?: string;
}

export interface PreviewResult {
	rows: PreviewRow[];
	totalAffectedItems: number;
	totalAffectedCells: number;
	totalItems: number;
	sampleSize: number;
	perRuleStats: RuleStat[];
	generatedAt: string; // RFC 3339
}

export interface ExportReport {
	path: string;
	itemsWritten: number;
	bytesWritten: number;
	durationMillis: number;
}

// Settings.CSVDelimiter is a {name, value} struct on the backend so the
// chosen delimiter's character is preserved across the bridge. The label list
// returned by `GetCSVDelimiters` is just strings; the frontend maps label →
// value locally in SettingsDialog.
export interface CSVDelimiter {
	name: string;
	value: string;
}

export interface Settings {
	csvDelimiter: CSVDelimiter;
	defaultDecimalPlaces: number;
}

export type ExportKind = 'csv' | 'xlsx' | 'rules-json';
