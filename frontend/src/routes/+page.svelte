<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import Icons from '$lib/components/Icons.svelte';
	import Toolbar from '$lib/components/Toolbar.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';
	import Notice from '$lib/components/Notice.svelte';
	import PreviewSkeleton from '$lib/components/PreviewSkeleton.svelte';
	import PreviewTable from '$lib/components/PreviewTable.svelte';
	import RuleList from '$lib/components/RuleList.svelte';
	import RuleEditor from '$lib/components/RuleEditor.svelte';
	import SettingsDialog from '$lib/components/SettingsDialog.svelte';
	import { newRule, type PreviewResult } from '$lib/backend';
	import { app } from '$lib/state/app.svelte';

	let settingsOpen = $state(false);

	onMount(() => {
		app.init();
	});
	onDestroy(() => {
		app.dispose();
	});

	function addRule() {
		const r = newRule();
		app.addRule(r);
	}

	function clearPreview() {
		app.preview = null;
	}

	// "15 May 2026 · 14:32:08" — locale-stable format derived from the
	// preview's RFC 3339 `generatedAt`. Returns null if the value can't be
	// parsed (e.g. older backend that doesn't populate it).
	function formatGeneratedAt(iso: string): string | null {
		const d = new Date(iso);
		if (Number.isNaN(d.getTime())) return null;
		const day = d.getDate();
		const month = d.toLocaleString('en-GB', { month: 'short' });
		const year = d.getFullYear();
		const hh = String(d.getHours()).padStart(2, '0');
		const mm = String(d.getMinutes()).padStart(2, '0');
		const ss = String(d.getSeconds()).padStart(2, '0');
		return `${day} ${month} ${year} · ${hh}:${mm}:${ss}`;
	}

	// Count rows in the displayed sample that actually changed. The backend
	// tops up the sample with leading unaffected rows when there are fewer
	// affected items than the sample size, so this is "shown affected", not
	// "shown rows".
	function shownAffected(p: PreviewResult): number {
		let n = 0;
		for (const row of p.rows) {
			if (row.cells.some((c) => c.changed)) n++;
		}
		return n;
	}
</script>

<Icons />

<div class="app">
	<Toolbar
		dataset={app.dataset}
		filename={app.filename}
		busy={app.busy}
		hasRules={app.hasRules}
		onImport={() => app.loadFromPicker()}
		onExportCsv={() => app.exportCSV()}
		onExportXlsx={() => app.exportXLSX()}
		onImportRules={() => app.importRulesFromPicker()}
		onExportRules={() => app.exportRules()}
		onOpenSettings={() => (settingsOpen = true)}
		onPreview={() => app.runPreview()}
	/>

	{#if !app.dataset}
		<EmptyState
			variant="dropzone"
			primaryAction={{ label: 'Import file', onClick: () => app.loadFromPicker() }}
		/>
	{:else}
		<div class="workspace">
			<div class="panel-row">
				<RuleList
					rules={app.rules}
					priceHeaders={app.dataset.priceHeaders}
					selectedId={app.selectedRuleId}
					statByRuleId={app.statByRuleId}
					erroredRuleIndex={app.errorRuleIndex}
					onSelect={(id) => (app.selectedRuleId = id)}
					onAdd={addRule}
					onDelete={(id) => app.removeRule(id)}
					onReorder={(next) => app.reorderRules(next)}
				/>
				<RuleEditor
					rule={app.selectedRule}
					priceHeaders={app.dataset.priceHeaders}
					onChange={(next) => app.updateRule(next)}
					onDelete={(id) => app.removeRule(id)}
					onAdd={addRule}
				/>
			</div>

			<section class="preview-panel">
				{#if app.preview && !app.error}
					{@const p = app.preview}
					{@const shown = shownAffected(p)}
					{@const when = formatGeneratedAt(p.generatedAt)}
					<div class="preview-table-status">
						<span class="label">Preview</span>
						<span class="sep">·</span>
						<span>
							Showing <span class="mono num-strong">{shown}</span> of
							<span class="mono num-strong">{p.totalAffectedItems.toLocaleString()}</span>
							affected items
						</span>
						<span class="sep">·</span>
						<span class="mono">
							{app.rules.length}
							{app.rules.length === 1 ? 'rule' : 'rules'} applied sequentially
						</span>
						{#if when}
							<span class="sep">·</span>
							<span>last run <span class="mono">{when}</span></span>
						{/if}
						<span class="spacer"></span>
						<span class="badge badge-positive" style="text-transform: none; letter-spacing: 0;">
							Preview is in sync
						</span>
					</div>
				{:else}
					<div class="panel-header">
						<span class="uppercase-label">Preview</span>
					</div>
				{/if}

				{#if app.error}
					<div style="padding: 10px 12px;">
						<ErrorBanner
							title={app.errorRuleIndex
								? `Preview failed at rule #${app.errorRuleIndex}`
								: 'Preview failed'}
							detail={app.error}
							onDismiss={() => app.dismissError()}
						/>
					</div>
				{/if}

				{#if app.runningPreview}
					<PreviewSkeleton
						rows={12}
						cols={Math.min(5, app.dataset.priceHeaders.length || 1)}
					/>
				{:else if app.preview && app.preview.totalAffectedItems === 0 && app.hasRules}
					<EmptyState
						variant="no-matches"
						primaryAction={{ label: 'View raw data', onClick: clearPreview }}
					/>
				{:else if app.preview}
					<PreviewTable mode="diff" dataset={app.dataset} rows={app.preview.rows} />
				{:else}
					<PreviewTable mode="raw" dataset={app.dataset} />
				{/if}
			</section>
		</div>
	{/if}
</div>

<SettingsDialog
	open={settingsOpen}
	settings={app.settings}
	onClose={() => (settingsOpen = false)}
	onSave={(next) => app.applySettings(next)}
/>

<Notice notice={app.notice} onDismiss={() => app.dismissNotice()} />
