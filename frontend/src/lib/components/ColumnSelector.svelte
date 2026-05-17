<script lang="ts">
	type Props = {
		priceHeaders: string[];
		selected: number[];
		onChange: (next: number[]) => void;
	};

	let { priceHeaders, selected, onChange }: Props = $props();

	const selectedSet = $derived(new Set(selected));
	const allColumns = $derived(selected.length === 0);

	function toggle(idx: number) {
		const next = new Set(selectedSet);
		if (next.has(idx)) next.delete(idx);
		else next.add(idx);
		// Sort so the array shape is stable; backend doesn't care but it makes
		// diffs and rule JSON output predictable.
		onChange([...next].sort((a, b) => a - b));
	}

	function selectAll() {
		onChange([]); // empty = all (HANDOFF §4.4)
	}

	function selectNone() {
		onChange([]); // collapse to "all" rather than have no targets — engine no-ops on empty list anyway
	}
</script>

<div class="full">
	<h3 style="display: flex; align-items: center; justify-content: space-between;">
		<span>Target columns</span>
		<span class="t-tertiary" style="font-size: var(--text-2xs); letter-spacing: 0;">
			empty selection = all
		</span>
	</h3>

	<div class="column-grid">
		{#each priceHeaders as h, i (i)}
			<label class={['column-check', selectedSet.has(i) && 'is-checked']}>
				<input
					class="checkbox"
					type="checkbox"
					checked={selectedSet.has(i)}
					onchange={() => toggle(i)}
				/>
				<span>{h}</span>
			</label>
		{/each}
	</div>

	<div style="display: flex; gap: 8px; margin-top: 6px;">
		<button class="btn btn-sm btn-ghost" type="button" onclick={selectAll}>All</button>
		<button class="btn btn-sm btn-ghost" type="button" onclick={selectNone}>None</button>
		{#if allColumns}
			<span
				class="t-tertiary"
				style="font-size: var(--text-xs); align-self: center; margin-left: auto;"
			>
				applies to every column
			</span>
		{/if}
	</div>
</div>
