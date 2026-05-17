<script lang="ts">
	import Icon from './Icon.svelte';

	type Variant = 'dropzone' | 'no-matches' | 'no-rule-selected';

	type Action = { label: string; onClick: () => void; primary?: boolean };

	type Props = {
		variant: Variant;
		primaryAction?: Action;
		secondaryAction?: Action;
	};

	let { variant, primaryAction, secondaryAction }: Props = $props();

	const content = $derived.by(() => {
		switch (variant) {
			case 'dropzone':
				return {
					icon: 'upload' as const,
					title: 'Drop a spreadsheet to begin',
					body: 'Drag a file here, or click "Import file" in the toolbar. Supports CSV and Excel.',
					formats: ['.csv', '.xlsx', '.xls']
				};
			case 'no-matches':
				return {
					icon: 'filter' as const,
					title: 'No items matched the current rules',
					body: 'Adjust a rule’s filter, or view the raw dataset.',
					formats: null
				};
			case 'no-rule-selected':
				return {
					icon: 'edit' as const,
					title: 'No rule selected',
					body: 'Pick a rule on the left, or add a new one to start editing.',
					formats: null
				};
		}
	});
</script>

<div class="dropzone-host">
	<div class={['dropzone-card', variant === 'dropzone' ? 'dropzone' : 'surface b-subtle']}>
		<div class="glyph">
			<Icon name={content.icon} size="lg" />
		</div>
		<h2>{content.title}</h2>
		<p>{content.body}</p>

		{#if content.formats}
			<div class="format-row">
				{#each content.formats as fmt (fmt)}
					<span class="chip">{fmt}</span>
				{/each}
			</div>
		{/if}

		{#if primaryAction || secondaryAction}
			<div class="actions">
				{#if secondaryAction}
					<button class="btn" type="button" onclick={secondaryAction.onClick}>
						{secondaryAction.label}
					</button>
				{/if}
				{#if primaryAction}
					<button
						class={primaryAction.primary === false ? 'btn' : 'btn btn-primary'}
						type="button"
						onclick={primaryAction.onClick}
					>
						{primaryAction.label}
					</button>
				{/if}
			</div>
		{/if}
	</div>
</div>
