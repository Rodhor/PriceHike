<script lang="ts">
	import type { Dataset, PreviewRow } from '$lib/backend';

	type Props =
		| { mode: 'raw'; dataset: Dataset; rows?: undefined }
		| { mode: 'diff'; dataset: Dataset; rows: PreviewRow[] };

	let props: Props = $props();

	// Cap raw mode for now; virtualization is a Task G follow-up if it matters.
	// 1000 rows is plenty to feel "scrollable dataset" without DOM blowup.
	const RAW_CAP = 1000;

	const rawRows = $derived(
		props.mode === 'raw' ? props.dataset.items.slice(0, RAW_CAP) : []
	);
	const rawHidden = $derived(
		props.mode === 'raw' ? Math.max(0, props.dataset.items.length - RAW_CAP) : 0
	);

	const headers = $derived(props.dataset.priceHeaders);
	const keyHeader = $derived(props.dataset.itemHeader || 'item');

	function diffClass(diff: string, changed: boolean): string {
		if (!changed) return 'cell-unchanged-num';
		if (diff.startsWith('+')) return 'diff-pos';
		if (diff.startsWith('-')) return 'diff-neg';
		return 'diff-zero';
	}
</script>

{#if props.mode === 'raw'}
	{#if rawHidden > 0}
		<div class="preview-table-status">
			Showing first {RAW_CAP.toLocaleString()} of {props.dataset.items.length.toLocaleString()} rows.
			Run a preview to see the effect of your rules.
		</div>
	{/if}
	<div class="preview-scroll">
		<table class="preview-table">
			<thead>
				<tr>
					<th class="col-key">{keyHeader}</th>
					{#each headers as h, i (i)}
						<th class="num">{h}</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#each rawRows as item, rowIdx (rowIdx)}
					<tr>
						<td class="col-key">{item.itemNumber}</td>
						{#each headers as _, c (c)}
							<td class="num">{item.prices[c] ?? ''}</td>
						{/each}
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{:else}
	<div class="preview-scroll">
		<table class="preview-table">
			<thead>
				<tr>
					<th class="col-key" rowspan="2">{keyHeader}</th>
					{#each headers as h, i (i)}
						<th class="col-group" colspan="3">{h}</th>
					{/each}
				</tr>
				<tr>
					{#each headers as _, i (i)}
						<th class="subhead col-group-start">old</th>
						<th class="subhead">new</th>
						<th class="subhead">diff</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#each props.rows as row, rowIdx (rowIdx)}
					<tr>
						<td class="col-key">{row.itemNumber}</td>
						{#each row.cells as cell, c (c)}
							<td class={['num', cell.changed && 'cell-changed', 'col-group-start']}>
								{cell.original}
							</td>
							<td class={['num', cell.changed && 'cell-changed']}>{cell.newValue}</td>
							<td class={['num', cell.changed && 'cell-changed', diffClass(cell.diff, cell.changed)]}>
								{cell.diff}
							</td>
						{/each}
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
