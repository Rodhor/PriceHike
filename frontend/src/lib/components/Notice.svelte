<script lang="ts">
	import type { Notice } from '$lib/state/app.svelte';
	import Icon from './Icon.svelte';

	type Props = {
		notice: Notice | null;
		onDismiss: () => void;
	};

	let { notice, onDismiss }: Props = $props();

	const toneClass = $derived.by(() => {
		if (!notice) return '';
		switch (notice.tone) {
			case 'success':
				return 'notice-success';
			case 'info':
				return 'notice-info';
			case 'error':
				return 'notice-error';
		}
	});

	const toneIcon = $derived.by<'play' | 'alert' | 'sheet'>(() => {
		if (!notice) return 'sheet';
		switch (notice.tone) {
			case 'success':
				return 'sheet';
			case 'error':
				return 'alert';
			case 'info':
			default:
				return 'sheet';
		}
	});
</script>

{#if notice}
	<div class="notice-host" role="status" aria-live="polite">
		<div class="notice {toneClass}">
			<Icon name={toneIcon} size="sm" />
			<span class="notice-msg">{notice.message}</span>
			<button
				class="btn btn-icon-sm btn-ghost"
				type="button"
				onclick={onDismiss}
				aria-label="Dismiss"
			>
				<Icon name="x" size="sm" />
			</button>
		</div>
	</div>
{/if}

<style>
	.notice-host {
		position: fixed;
		bottom: 16px;
		right: 16px;
		z-index: 200;
		max-width: min(560px, calc(100vw - 32px));
	}
	.notice {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		background: var(--bg-surface);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-pop);
		font-size: var(--text-sm);
		color: var(--text-primary);
	}
	.notice-msg {
		flex: 1;
		min-width: 0;
		font-variant-numeric: tabular-nums;
	}
	.notice-success {
		border-color: var(--positive-border);
		background: var(--positive-bg);
		color: var(--positive);
	}
	.notice-error {
		border-color: var(--negative-border);
		background: var(--negative-bg);
		color: var(--negative);
	}
	.notice-info {
		border-color: var(--info-border);
		background: var(--info-bg);
		color: var(--info);
	}
</style>
