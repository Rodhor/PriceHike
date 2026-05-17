<script lang="ts">
	import type { DecimalSettings } from '$lib/backend';

	type Props = {
		settings: DecimalSettings;
		onChange: (next: DecimalSettings) => void;
	};

	let { settings, onChange }: Props = $props();

	const PRESETS: Array<{ label: string; mode: 'keep' | 'fixed'; places: number }> = [
		{ label: 'Keep', mode: 'keep', places: 2 },
		{ label: '0', mode: 'fixed', places: 0 },
		{ label: '2', mode: 'fixed', places: 2 },
		{ label: '3', mode: 'fixed', places: 3 },
		{ label: '4', mode: 'fixed', places: 4 }
	];

	function isActive(p: { mode: 'keep' | 'fixed'; places: number }): boolean {
		if (p.mode === 'keep') return settings.mode === 'keep';
		return settings.mode === 'fixed' && settings.places === p.places;
	}

	function applyPreset(p: { mode: 'keep' | 'fixed'; places: number }) {
		onChange({ mode: p.mode, places: p.mode === 'fixed' ? p.places : settings.places });
	}

	function onCustomInput(e: Event) {
		const n = parseInt((e.target as HTMLInputElement).value, 10);
		if (Number.isFinite(n)) onChange({ mode: 'fixed', places: n });
	}

	const customValue = $derived(settings.mode === 'fixed' ? String(settings.places) : '');
</script>

<div class="full">
	<h3>Decimals</h3>

	<div style="display: flex; align-items: center; gap: 12px; flex-wrap: wrap;">
		<div class="segmented">
			{#each PRESETS as p (p.label)}
				<button type="button" class={isActive(p) ? 'is-active' : ''} onclick={() => applyPreset(p)}>
					{p.label}
				</button>
			{/each}
		</div>

		<label
			class="field-label"
			style="display: flex; align-items: center; gap: 6px; margin: 0;"
		>
			<span class="t-secondary" style="font-weight: 400;">custom</span>
			<input
				class="field field-mono"
				type="number"
				min="0"
				max="10"
				step="1"
				placeholder="N"
				value={customValue}
				disabled={settings.mode === 'keep'}
				oninput={onCustomInput}
				style="width: 80px;"
			/>
		</label>
	</div>

	<p class="t-tertiary" style="font-size: var(--text-xs); margin: 8px 0 0 0;">
		{#if settings.mode === 'keep'}
			Keep each cell’s original precision (count trailing zeros from the source).
		{:else}
			Format to exactly {settings.places} decimal place{settings.places === 1 ? '' : 's'}, zero-padded.
		{/if}
	</p>
</div>
