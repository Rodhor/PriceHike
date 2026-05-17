<script lang="ts">
	import type { Dataset } from '$lib/backend';
	import Icon from './Icon.svelte';

	let { filename, dataset }: { filename: string; dataset: Dataset } = $props();

	const rowCount = $derived(dataset.items.length);
	const priceColCount = $derived(dataset.priceHeaders.length);
</script>

<div class="toolbar-file" title={filename}>
	<Icon name="sheet" />
	<span class="filename">{filename}</span>
	<span class="meta">·</span>
	<span class="meta">{rowCount.toLocaleString()} rows</span>
	<span class="meta">·</span>
	<span class="meta">{priceColCount} price cols</span>
</div>
