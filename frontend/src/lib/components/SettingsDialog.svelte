<script lang="ts">
	import { backend, type CSVDelimiter, type Settings } from '$lib/backend';
	import {
		appearance,
		PALETTES,
		PALETTE_LABELS,
		PALETTE_SWATCHES,
		type Palette
	} from '$lib/state/appearance.svelte';
	import Icon from './Icon.svelte';

	type Props = {
		open: boolean;
		settings: Settings | null;
		onClose: () => void;
		onSave: (next: Settings) => Promise<void> | void;
	};

	let { open, settings, onClose, onSave }: Props = $props();

	let delimiters = $state<CSVDelimiter[]>([]);
	let workingName = $state<string>('Comma');
	let workingPlaces = $state<number>(2);
	let saving = $state(false);
	let saveError = $state<string | null>(null);

	$effect(() => {
		if (open) hydrate();
	});

	async function hydrate() {
		saveError = null;
		if (settings) {
			workingName = settings.csvDelimiter.name;
			workingPlaces = settings.defaultDecimalPlaces;
		}
		try {
			delimiters = await backend.getCSVDelimiters();
		} catch (err) {
			saveError = err instanceof Error ? err.message : String(err);
		}
	}

	const chosenDelimiter = $derived<CSVDelimiter | null>(
		delimiters.find((d) => d.name === workingName) ?? delimiters[0] ?? null
	);

	const placesValid = $derived(
		Number.isInteger(workingPlaces) && workingPlaces >= 0 && workingPlaces <= 10
	);

	const chosenValid = $derived(chosenDelimiter != null && chosenDelimiter.value.length > 0);

	function describe(d: CSVDelimiter): string {
		switch (d.value) {
			case '\t':
				return '\\t';
			case '\n':
				return '\\n';
			case ' ':
				return '" "';
			default:
				return d.value;
		}
	}

	async function handleSave() {
		if (!placesValid) {
			saveError = 'Decimal places must be an integer in 0..10.';
			return;
		}
		if (!chosenValid || !chosenDelimiter) {
			saveError = `No delimiter is selected.`;
			return;
		}
		saving = true;
		saveError = null;
		try {
			await onSave({
				csvDelimiter: chosenDelimiter,
				defaultDecimalPlaces: workingPlaces
			});
			onClose();
		} catch (err) {
			saveError = err instanceof Error ? err.message : String(err);
		} finally {
			saving = false;
		}
	}

	function handleBackdrop(e: MouseEvent) {
		if (e.target === e.currentTarget) onClose();
	}

	// Palette and theme apply live — they're frontend-only UI preferences, no
	// need to round-trip through Save. Tracked in localStorage by the
	// appearance module so they survive reloads.
	function pickPalette(p: Palette) {
		appearance.setPalette(p);
	}
</script>

{#if open}
	<div
		class="settings-backdrop"
		role="button"
		tabindex="-1"
		onclick={handleBackdrop}
		onkeydown={(e) => e.key === 'Escape' && onClose()}
	>
		<div
			class="settings-card surface b-default"
			role="dialog"
			aria-modal="true"
			aria-labelledby="settings-title"
		>
			<header class="settings-header">
				<h2 id="settings-title">Settings</h2>
				<button
					class="btn btn-icon-sm btn-ghost"
					type="button"
					aria-label="Close"
					onclick={onClose}
				>
					<Icon name="x" size="sm" />
				</button>
			</header>

			<div class="settings-body">
				<section>
					<h3 class="settings-section">File handling</h3>

					<div class="settings-field">
						<label class="field-label" for="csv-delimiter">CSV delimiter</label>
						<select
							id="csv-delimiter"
							class="field"
							value={workingName}
							onchange={(e) => (workingName = (e.target as HTMLSelectElement).value)}
						>
							{#each delimiters as d (d.name)}
								<option value={d.name}>{d.name} ({describe(d)})</option>
							{/each}
						</select>
						<p class="hint">
							Used when loading CSV files and when exporting back to CSV. Changing this re-reads
							the currently loaded file.
						</p>
					</div>

					<div class="settings-field">
						<label class="field-label" for="default-places">Default decimal places</label>
						<input
							id="default-places"
							class="field field-mono"
							type="number"
							min="0"
							max="10"
							step="1"
							value={workingPlaces}
							oninput={(e) =>
								(workingPlaces = parseInt((e.target as HTMLInputElement).value, 10))}
							style="max-width: 120px;"
						/>
						<p class="hint">Applied to price columns that no rule explicitly rounded. 0–10.</p>
					</div>
				</section>

				<section>
					<h3 class="settings-section">Appearance</h3>

					<div class="settings-field">
						<span class="field-label">Accent palette</span>
						<div class="palette-row" role="radiogroup" aria-label="Accent palette">
							{#each PALETTES as p (p)}
								<button
									type="button"
									role="radio"
									aria-checked={appearance.palette === p}
									class={['palette-chip', appearance.palette === p && 'is-active']}
									title={PALETTE_LABELS[p]}
									onclick={() => pickPalette(p)}
								>
									<span class="swatch" style:background-color={PALETTE_SWATCHES[p]}></span>
									<span class="palette-label">{PALETTE_LABELS[p]}</span>
								</button>
							{/each}
						</div>
						<p class="hint">Applies immediately; persists across reloads.</p>
					</div>
				</section>

				{#if saveError}
					<div class="banner" role="alert">
						<Icon name="alert" class="icon" />
						<span style="flex: 1; font-size: var(--text-sm);">{saveError}</span>
					</div>
				{/if}
			</div>

			<footer class="settings-footer">
				<button class="btn" type="button" onclick={onClose} disabled={saving}>Cancel</button>
				<button
					class="btn btn-primary"
					type="button"
					onclick={handleSave}
					disabled={saving || !placesValid || !chosenValid}
				>
					{saving ? 'Saving…' : 'Save'}
				</button>
			</footer>
		</div>
	</div>
{/if}

<style>
	.settings-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(15, 23, 42, 0.45);
		display: grid;
		place-items: center;
		z-index: 100;
		padding: 24px;
	}
	.settings-card {
		width: 100%;
		max-width: 460px;
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-pop);
		display: flex;
		flex-direction: column;
	}
	.settings-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		border-bottom: 1px solid var(--border-subtle);
	}
	.settings-header h2 {
		margin: 0;
		font-size: var(--text-md);
		font-weight: 600;
	}
	.settings-body {
		padding: 16px;
		display: grid;
		gap: 18px;
	}
	.settings-section {
		font-size: var(--text-2xs);
		text-transform: uppercase;
		letter-spacing: 0.06em;
		font-weight: 700;
		color: var(--text-tertiary);
		margin: 0 0 10px 0;
	}
	.settings-field {
		display: grid;
		gap: 4px;
	}
	.settings-field + .settings-field {
		margin-top: 12px;
	}
	.hint {
		font-size: var(--text-xs);
		color: var(--text-tertiary);
		margin: 4px 0 0 0;
	}
	.settings-footer {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
		padding: 10px 16px;
		border-top: 1px solid var(--border-subtle);
		background: var(--bg-surface-2);
		border-bottom-left-radius: var(--radius-lg);
		border-bottom-right-radius: var(--radius-lg);
	}

	/* Palette picker — chip per palette with a coloured swatch and the name. */
	.palette-row {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 6px;
		margin-top: 2px;
	}
	.palette-chip {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 8px;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-sm);
		background: var(--bg-surface);
		color: var(--text-secondary);
		font: inherit;
		font-size: var(--text-sm);
		cursor: pointer;
		text-align: left;
		transition:
			border-color 100ms,
			background 100ms,
			color 100ms;
	}
	.palette-chip:hover {
		background: var(--bg-hover);
		border-color: var(--border-default);
	}
	.palette-chip.is-active {
		background: var(--accent-soft);
		border-color: var(--accent-soft-border);
		color: var(--accent);
		font-weight: 600;
	}
	.palette-chip .swatch {
		display: inline-block;
		width: 14px;
		height: 14px;
		border-radius: 50%;
		flex-shrink: 0;
		box-shadow:
			inset 0 0 0 1px rgba(0, 0, 0, 0.12),
			0 1px 2px rgba(0, 0, 0, 0.15);
	}
	.palette-chip .palette-label {
		flex: 1;
		min-width: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
</style>
