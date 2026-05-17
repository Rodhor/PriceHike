<script lang="ts">
	import type { Dataset } from '$lib/backend';
	import Icon from './Icon.svelte';
	import FileMetaBadge from './FileMetaBadge.svelte';
	import ThemeToggle from './ThemeToggle.svelte';

	type Props = {
		dataset: Dataset | null;
		filename: string | null;
		busy: boolean;
		hasRules: boolean;
		onImport: () => void;
		onExportCsv: () => void;
		onExportXlsx: () => void;
		onImportRules: () => void;
		onExportRules: () => void;
		onOpenSettings: () => void;
		onPreview: () => void;
	};

	let {
		dataset,
		filename,
		busy,
		hasRules,
		onImport,
		onExportCsv,
		onExportXlsx,
		onImportRules,
		onExportRules,
		onOpenSettings,
		onPreview
	}: Props = $props();

	const loaded = $derived(dataset != null);
</script>

<header class="toolbar">
	<div class="toolbar-brand">
		<div class="logo">P</div>
		<div class="brand-text">
			<span class="name">PriceHike</span>
			<span class="version">v0.4 · local</span>
		</div>
	</div>

	{#if loaded && dataset && filename}
		<FileMetaBadge {filename} {dataset} />
	{/if}

	<div class="toolbar-spacer"></div>

	<button class="btn" type="button" onclick={onImport} disabled={busy}>
		<Icon name="upload" />
		Import file
	</button>

	<div class="toolbar-divider"></div>

	<button
		class="btn"
		type="button"
		onclick={onExportCsv}
		disabled={!loaded || busy}
		title={!loaded ? 'Load a file first' : 'Export the dataset (with rules applied) to CSV'}
	>
		<Icon name="download" />
		Export CSV
	</button>
	<button
		class="btn"
		type="button"
		onclick={onExportXlsx}
		disabled={!loaded || busy}
		title={!loaded ? 'Load a file first' : 'Export the dataset (with rules applied) to XLSX'}
	>
		<Icon name="download" />
		Export XLSX
	</button>

	<div class="toolbar-divider"></div>

	<button
		class="btn btn-ghost"
		type="button"
		onclick={onImportRules}
		disabled={busy}
		title="Load rules from a .json file (replaces the current rule list)"
	>
		<Icon name="upload" />
		Import rules
	</button>
	<button
		class="btn btn-ghost"
		type="button"
		onclick={onExportRules}
		disabled={!hasRules || busy}
		title={!hasRules ? 'Add at least one rule first' : 'Save the current rules as JSON'}
	>
		<Icon name="download" />
		Export rules
	</button>

	<div class="toolbar-divider"></div>

	<button
		class="btn btn-primary"
		type="button"
		onclick={onPreview}
		disabled={!loaded || !hasRules || busy}
		title={!loaded
			? 'Load a file first'
			: !hasRules
				? 'Add at least one rule first'
				: 'Run preview'}
	>
		{#if busy}
			<Icon name="loader" />
		{:else}
			<Icon name="play" />
		{/if}
		Preview changes
	</button>

	<button
		class="btn btn-icon btn-ghost"
		type="button"
		onclick={onOpenSettings}
		title="Settings"
		aria-label="Settings"
	>
		<Icon name="settings" />
	</button>

	<ThemeToggle />
</header>
