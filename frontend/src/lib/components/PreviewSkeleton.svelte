<script lang="ts">
	type Props = { rows?: number; cols?: number };
	let { rows = 12, cols = 5 }: Props = $props();

	const rowIndices = $derived(Array.from({ length: rows }, (_, i) => i));
	const colIndices = $derived(Array.from({ length: cols }, (_, i) => i));
</script>

<div class="preview-scroll" aria-busy="true" aria-label="Generating preview">
	<table class="preview-table">
		<thead>
			<tr>
				<th class="col-key">&nbsp;</th>
				{#each colIndices as _, i (i)}
					<th class="num"><span class="skeleton" style="display: inline-block; width: 40px; height: 12px;"></span></th>
				{/each}
			</tr>
		</thead>
		<tbody>
			{#each rowIndices as _, i (i)}
				<tr>
					<td class="col-key">
						<span class="skeleton" style="display: inline-block; width: 80px; height: 12px;"></span>
					</td>
					{#each colIndices as _, c (c)}
						<td class="num">
							<span class="skeleton" style="display: inline-block; width: 50px; height: 12px;"></span>
						</td>
					{/each}
				</tr>
			{/each}
		</tbody>
	</table>
</div>
