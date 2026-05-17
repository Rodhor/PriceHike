import {
	backend,
	onFileDropped,
	onFileError,
	parseRuleError,
	type Dataset,
	type ExportReport,
	type PreviewResult,
	type Rule,
	type RuleStat,
	type Settings
} from '$lib/backend';

const DEFAULT_SAMPLE_SIZE = 500;
const NOTICE_TIMEOUT_MS = 4000;

function basename(p: string): string {
	const idx = Math.max(p.lastIndexOf('/'), p.lastIndexOf('\\'));
	return idx >= 0 ? p.slice(idx + 1) : p;
}

function defaultExportName(filename: string | null, ext: string): string {
	const base = filename ? filename.replace(/\.(csv|xlsx|xls|json)$/i, '') : 'export';
	return `${base}.${ext}`;
}

function extOf(path: string): string {
	const m = path.toLowerCase().match(/\.([^./\\]+)$/);
	return m ? m[1] : '';
}

const SPREADSHEET_EXTS = new Set(['csv', 'xlsx', 'xls']);

export interface Notice {
	id: number;
	tone: 'success' | 'info' | 'error';
	message: string;
}

class AppState {
	// ---- core state ----
	dataset = $state<Dataset | null>(null);
	filename = $state<string | null>(null);
	// Full filesystem path of the currently loaded file, when known. Tracked
	// so `applySettings()` can re-read the file with the new CSV delimiter
	// without a fresh import.
	path = $state<string | null>(null);

	rules = $state<Rule[]>([]);
	selectedRuleId = $state<string | null>(null);

	preview = $state<PreviewResult | null>(null);

	loadingFile = $state(false);
	runningPreview = $state(false);
	exporting = $state(false);
	busy = $derived(this.loadingFile || this.runningPreview || this.exporting);

	error = $state<string | null>(null);
	errorRuleIndex = $state<number | null>(null);

	notice = $state<Notice | null>(null);
	private noticeSeq = 0;
	private noticeTimer: ReturnType<typeof setTimeout> | null = null;

	settings = $state<Settings | null>(null);

	// ---- derived ----
	hasDataset = $derived(this.dataset != null);
	hasRules = $derived(this.rules.length > 0);
	selectedRule = $derived(
		this.selectedRuleId
			? (this.rules.find((r) => r.id === this.selectedRuleId) ?? null)
			: null
	);
	statByRuleId = $derived.by(() => {
		const m = new Map<string, RuleStat>();
		for (const s of this.preview?.perRuleStats ?? []) m.set(s.ruleId, s);
		return m;
	});

	// ---- lifecycle ----
	private disposers: Array<() => void> = [];

	async init() {
		try {
			this.settings = await backend.getSettings();
		} catch (err) {
			console.error('GetSettings failed', err);
		}

		this.disposers.push(
			onFileDropped((filename, dataset) => {
				// Drop events only carry the basename. We don't have the full
				// path, so `applySettings` reloads will fall back to the
				// "re-import to apply" notice for dropped files. Tracked in
				// docs/Cleanup-Backend.md as a small drop-event payload tweak.
				this.acceptLoaded(dataset, filename, null);
			})
		);
		this.disposers.push(
			onFileError((message) => {
				this.setError(message);
			})
		);
	}

	dispose() {
		for (const fn of this.disposers) fn();
		this.disposers = [];
		if (this.noticeTimer) clearTimeout(this.noticeTimer);
	}

	// ---- file load ----
	async loadFromPicker(): Promise<void> {
		this.error = null;
		let path: string;
		try {
			path = await backend.selectImportFile();
		} catch (err) {
			const msg = messageFor(err);
			if (msg !== 'cancelled') this.setError(msg);
			return;
		}
		if (!path) return;
		const ext = extOf(path);
		if (ext === 'json') {
			this.setError(
				`That looks like a rules file (.json). Use the "Import rules" button to load it.`
			);
			return;
		}
		if (!SPREADSHEET_EXTS.has(ext)) {
			this.setError(`Unsupported file extension: .${ext || '?'}`);
			return;
		}
		await this.loadFromPath(path);
	}

	async loadFromPath(path: string): Promise<void> {
		this.loadingFile = true;
		this.error = null;
		try {
			const ds = await backend.loadFile(path);
			this.acceptLoaded(ds, basename(path), path);
		} catch (err) {
			this.setError(messageFor(err));
		} finally {
			this.loadingFile = false;
		}
	}

	private acceptLoaded(dataset: Dataset, filename: string, path: string | null) {
		this.dataset = dataset;
		this.filename = filename;
		if (path !== null) this.path = path;
		this.preview = null;
		this.error = null;
		this.errorRuleIndex = null;
	}

	// ---- rule import ----
	async importRulesFromPicker(): Promise<void> {
		this.error = null;
		let path: string;
		try {
			path = await backend.selectImportFile();
		} catch (err) {
			const msg = messageFor(err);
			if (msg !== 'cancelled') this.setError(msg);
			return;
		}
		if (!path) return;
		const ext = extOf(path);
		if (ext !== 'json') {
			this.setError(
				`Expected a rules file (.json) but got .${ext || '?'}. Use the "Import file" button for spreadsheets.`
			);
			return;
		}
		await this.importRulesFromPath(path);
	}

	async importRulesFromPath(path: string): Promise<void> {
		this.loadingFile = true;
		this.error = null;
		this.errorRuleIndex = null;
		try {
			const rules = await backend.importRules(path);
			this.replaceRules(rules);
			this.showNotice(
				'success',
				rules.length === 1
					? `Imported 1 rule from ${basename(path)}`
					: `Imported ${rules.length} rules from ${basename(path)}`
			);
		} catch (err) {
			this.setError(messageFor(err));
		} finally {
			this.loadingFile = false;
		}
	}

	// ---- error / notice ----
	setError(message: string) {
		this.error = message;
		const parsed = parseRuleError(message);
		this.errorRuleIndex = parsed?.index ?? null;
	}
	dismissError() {
		this.error = null;
		this.errorRuleIndex = null;
	}
	showNotice(tone: Notice['tone'], message: string) {
		this.noticeSeq++;
		this.notice = { id: this.noticeSeq, tone, message };
		if (this.noticeTimer) clearTimeout(this.noticeTimer);
		const id = this.noticeSeq;
		this.noticeTimer = setTimeout(() => {
			if (this.notice?.id === id) this.notice = null;
		}, NOTICE_TIMEOUT_MS);
	}
	dismissNotice() {
		this.notice = null;
		if (this.noticeTimer) {
			clearTimeout(this.noticeTimer);
			this.noticeTimer = null;
		}
	}

	// ---- settings ----
	async applySettings(next: Settings): Promise<void> {
		const prevDelim = this.settings?.csvDelimiter.value;
		await backend.setSettings(next);
		this.settings = next;

		// Re-parse the loaded file when the CSV delimiter changed; the new
		// delimiter only affects loading, so for non-CSV datasets a reload is
		// a no-op. Best to do it unconditionally and let the reader sort it
		// out (Excel ignores the delimiter setting anyway).
		const delimiterChanged = next.csvDelimiter.value !== prevDelim;
		if (delimiterChanged && this.dataset && this.path) {
			try {
				await this.loadFromPath(this.path);
				this.showNotice('info', `Re-read ${this.filename} with the new delimiter.`);
			} catch (err) {
				this.setError(messageFor(err));
			}
			return;
		}

		// No path tracked (e.g. dropped file on a backend build where Filepath
		// isn't exposed): tell the user they need to re-import for it to take
		// effect on the CURRENT dataset. Future loads will already use it.
		if (delimiterChanged && this.dataset && !this.path) {
			this.showNotice(
				'info',
				'Settings saved. Re-import the file to apply the new CSV delimiter to the current dataset.'
			);
		}
	}

	// ---- rules ----
	addRule(rule: Rule) {
		this.rules = [...this.rules, rule];
		this.selectedRuleId = rule.id;
		this.preview = null;
	}
	removeRule(id: string) {
		this.rules = this.rules.filter((r) => r.id !== id);
		if (this.selectedRuleId === id) this.selectedRuleId = null;
		this.preview = null;
	}
	updateRule(next: Rule) {
		this.rules = this.rules.map((r) => (r.id === next.id ? next : r));
		this.preview = null;
	}
	reorderRules(next: Rule[]) {
		this.rules = next;
		this.preview = null;
	}
	replaceRules(next: Rule[]) {
		this.rules = next;
		this.selectedRuleId = next[0]?.id ?? null;
		this.preview = null;
	}

	// ---- preview ----
	async runPreview(): Promise<void> {
		if (!this.dataset) return;
		this.runningPreview = true;
		this.error = null;
		this.errorRuleIndex = null;
		try {
			this.preview = await backend.generatePreview(this.rules, DEFAULT_SAMPLE_SIZE);
		} catch (err) {
			this.setError(messageFor(err));
			this.preview = null;
		} finally {
			this.runningPreview = false;
		}
	}

	// ---- export ----
	async exportCSV(): Promise<void> {
		await this.exportSpreadsheet('csv');
	}
	async exportXLSX(): Promise<void> {
		await this.exportSpreadsheet('xlsx');
	}
	private async exportSpreadsheet(kind: 'csv' | 'xlsx'): Promise<void> {
		if (!this.dataset) return;
		const defaultName = defaultExportName(this.filename, kind);
		let path: string;
		try {
			path = await backend.selectExportPath(defaultName, kind);
		} catch (err) {
			const msg = messageFor(err);
			if (msg !== 'cancelled') this.setError(msg);
			return;
		}
		if (!path) return;

		this.exporting = true;
		this.error = null;
		try {
			const report =
				kind === 'csv'
					? await backend.applyAndExportCSV(this.rules, path)
					: await backend.applyAndExportXLSX(this.rules, path);
			this.showNotice(
				'success',
				`Exported ${report.itemsWritten.toLocaleString()} rows · ${formatBytes(report.bytesWritten)} · ${report.durationMillis} ms → ${basename(report.path)}`
			);
		} catch (err) {
			this.setError(messageFor(err));
		} finally {
			this.exporting = false;
		}
	}

	async exportRules(): Promise<void> {
		const defaultName = defaultExportName(this.filename, 'json');
		let path: string;
		try {
			path = await backend.selectExportPath(defaultName, 'rules-json');
		} catch (err) {
			const msg = messageFor(err);
			if (msg !== 'cancelled') this.setError(msg);
			return;
		}
		if (!path) return;

		this.exporting = true;
		this.error = null;
		try {
			await backend.exportRules(this.rules, path);
			this.showNotice('success', `Saved ${this.rules.length} rules → ${basename(path)}`);
		} catch (err) {
			this.setError(messageFor(err));
		} finally {
			this.exporting = false;
		}
	}
}

function messageFor(err: unknown): string {
	if (err instanceof Error) return err.message;
	if (typeof err === 'string') return err;
	return String(err);
}

function formatBytes(n: number): string {
	if (n < 1024) return `${n} B`;
	if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} kB`;
	return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}

export const app = new AppState();
