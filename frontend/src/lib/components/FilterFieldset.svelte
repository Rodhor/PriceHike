<script lang="ts">
	import type { RuleFilter } from '$lib/backend';

	type Props = {
		filter: RuleFilter;
		priceHeaders: string[];
		onChange: (next: RuleFilter) => void;
	};

	let { filter, priceHeaders, onChange }: Props = $props();

	const wildcard = $derived(filter.wildcard ?? '');
	const thresholdOn = $derived(filter.lessThan != null);
	const thresholdColumn = $derived(filter.lessThan?.column ?? 0);
	const thresholdValue = $derived(filter.lessThan?.value ?? '');

	function emit(next: Partial<RuleFilter>) {
		// applyToAll is recomputed at submit-time by backend.normalizeFilter — we
		// keep it false here and let the wrapper handle the match-all case.
		onChange({
			applyToAll: false,
			wildcard: filter.wildcard ?? '',
			lessThan: filter.lessThan,
			greaterThan: filter.greaterThan,
			...next
		});
	}

	function onWildcardInput(e: Event) {
		const v = (e.target as HTMLInputElement).value;
		emit({ wildcard: v });
	}

	function toggleThreshold(e: Event) {
		const on = (e.target as HTMLInputElement).checked;
		if (on) {
			emit({ lessThan: { column: 0, value: '0' } });
		} else {
			emit({ lessThan: undefined });
		}
	}

	function onThresholdColumn(e: Event) {
		const col = parseInt((e.target as HTMLSelectElement).value, 10);
		emit({ lessThan: { column: col, value: filter.lessThan?.value ?? '0' } });
	}

	function onThresholdValue(e: Event) {
		const v = (e.target as HTMLInputElement).value;
		emit({ lessThan: { column: filter.lessThan?.column ?? 0, value: v } });
	}
</script>

<div class="full">
	<h3>Filter</h3>

	<label class="field-label" for="filter-wildcard">Item-number wildcard</label>
	<input
		id="filter-wildcard"
		class="field field-mono"
		type="text"
		placeholder="* matches all  ·  use * and ? as wildcards"
		value={wildcard}
		oninput={onWildcardInput}
	/>
	<p class="t-tertiary" style="font-size: var(--text-xs); margin: 4px 0 0 0;">
		Leave empty to match every item.
	</p>
</div>

<div class="full" style="margin-top: 6px;">
	<label
		class="field-label"
		style="display: flex; align-items: center; gap: 8px; cursor: pointer;"
	>
		<input
			class="checkbox"
			type="checkbox"
			checked={thresholdOn}
			onchange={toggleThreshold}
		/>
		Only apply when a price column is less than…
	</label>

	{#if thresholdOn}
		<div class="input-row" style="margin-top: 6px;">
			<span class="t-secondary" style="font-size: var(--text-xs);">column</span>
			<select class="field" value={thresholdColumn} onchange={onThresholdColumn}>
				{#each priceHeaders as h, i (i)}
					<option value={i}>{h}</option>
				{/each}
			</select>
			<span class="t-secondary" style="font-size: var(--text-xs);">{`<`}</span>
			<input
				class="field field-mono"
				type="text"
				inputmode="decimal"
				placeholder="0.00"
				value={thresholdValue}
				oninput={onThresholdValue}
			/>
		</div>
	{/if}
</div>
