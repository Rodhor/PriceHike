<script lang="ts">
	import Icon from './Icon.svelte';

	type Props = {
		title?: string;
		detail: string;
		onDismiss?: () => void;
	};

	let { title = 'Something went wrong', detail, onDismiss }: Props = $props();
</script>

<div class="banner" role="alert">
	<Icon name="alert" class="icon" />
	<div style="flex: 1; min-width: 0;">
		<div style="font-weight: 600; color: var(--negative); font-size: var(--text-sm);">
			{title}
		</div>
		<div
			class="mono"
			style="margin-top: 2px; font-size: var(--text-xs); color: var(--text-secondary); white-space: pre-wrap; word-break: break-word;"
		>
			{detail}
		</div>
	</div>
	{#if onDismiss}
		<button
			class="btn btn-icon-sm btn-ghost"
			type="button"
			aria-label="Dismiss"
			onclick={onDismiss}
		>
			<Icon name="x" size="sm" />
		</button>
	{/if}
</div>
