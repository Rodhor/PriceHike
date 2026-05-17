<script lang="ts">
	import { ACTION_LABELS, ACTION_TONES, type Rule, type RuleStat } from '$lib/backend';
	import Icon from './Icon.svelte';

	type Props = {
		rule: Rule;
		ordinal: number;
		priceHeaders: string[];
		selected: boolean;
		// Optional — only available after the user has run a Preview.
		stat?: RuleStat;
		// Set when GeneratePreview rejected with "rule #N: <detail>" pointing
		// at this rule. The card paints with the error variant.
		errored?: boolean;
		onSelect: () => void;
		onDelete: () => void;
	};

	let {
		rule,
		ordinal,
		priceHeaders,
		selected,
		stat,
		errored,
		onSelect,
		onDelete
	}: Props = $props();

	const actionTone = $derived(ACTION_TONES[rule.action.type]);
	const actionBadge = $derived.by(() => {
		const type = rule.action.type;
		switch (type) {
			case 'addPercent':
				return `+${rule.action.value}%`;
			case 'subtractPercent':
				return `−${rule.action.value}%`;
			case 'addFixed':
				return `+${rule.action.value}`;
			case 'subtractFixed':
				return `−${rule.action.value}`;
			case 'roundUp':
				return 'round ↑';
			case 'roundDown':
				return 'round ↓';
			default:
				return ACTION_LABELS[type];
		}
	});

	const filterChips = $derived.by(() => {
		const chips: string[] = [];
		const f = rule.filter;
		const hasWildcard = !!f.wildcard && f.wildcard.length > 0;
		const hasLess = f.lessThan != null;
		const hasGreater = f.greaterThan != null;

		if (!hasWildcard && !hasLess && !hasGreater) return chips;
		if (hasWildcard) chips.push(f.wildcard!);
		if (hasLess) {
			const h = priceHeaders[f.lessThan!.column] ?? `col${f.lessThan!.column}`;
			chips.push(`${h} < ${f.lessThan!.value}`);
		}
		if (hasGreater) {
			const h = priceHeaders[f.greaterThan!.column] ?? `col${f.greaterThan!.column}`;
			chips.push(`${h} > ${f.greaterThan!.value}`);
		}
		return chips;
	});

	const matchAll = $derived(filterChips.length === 0);

	const columnChips = $derived.by(() => {
		if (!rule.columns || rule.columns.length === 0) return ['all columns'];
		return rule.columns.map((c) => priceHeaders[c] ?? `col${c}`);
	});

	const decimalsLabel = $derived(
		rule.decimalSettings.mode === 'keep' ? 'keep' : `${rule.decimalSettings.places}dp`
	);

	const badgeClass = $derived(
		actionTone === 'positive'
			? 'badge badge-positive'
			: actionTone === 'negative'
				? 'badge badge-negative'
				: 'badge badge-neutral'
	);

	// Sub-meta line, in priority order: per-rule error → per-rule warning →
	// per-rule counts → nothing.
	const subMeta = $derived.by(() => {
		if (!stat) return null;
		if (stat.warning) return { kind: 'warning' as const, text: stat.warning };
		return {
			kind: 'counts' as const,
			text: `${stat.itemsMatched.toLocaleString()} items matched · ${stat.cellsChanged.toLocaleString()} cells changed`
		};
	});
</script>

<div
	class={['rule-card', selected && 'is-selected', errored && 'is-error', stat?.warning && 'is-warning']}
	role="button"
	tabindex="0"
	onclick={onSelect}
	onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && (e.preventDefault(), onSelect())}
>
	<div class="grip" aria-hidden="true">
		<Icon name="grip" />
	</div>
	<div class="ord">{ordinal}</div>
	<div class="body">
		<div class="summary">
			<span class={badgeClass}>{actionBadge}</span>
			{#if matchAll}
				<span class="t-secondary">on</span>
				<span class="chip">all items</span>
			{:else}
				<span class="t-secondary">on</span>
				{#each filterChips as chip, i (i)}
					<span class="chip">{chip}</span>
				{/each}
			{/if}
			<span class="t-tertiary">→</span>
			{#each columnChips as chip, i (i)}
				<span class="chip">{chip}</span>
			{/each}
			<span class="t-tertiary">·</span>
			<span class="mono" style="font-size: var(--text-xs); color: var(--text-secondary);"
				>{decimalsLabel}</span
			>
		</div>
		{#if errored}
			<div class="submeta" style="color: var(--negative); font-weight: 600;">
				preview rejected this rule — fix and re-run
			</div>
		{:else if subMeta}
			<div class="submeta">{subMeta.text}</div>
		{/if}
	</div>
	<div class="actions">
		<button
			class="btn btn-icon-sm btn-danger-ghost"
			type="button"
			title="Delete rule"
			aria-label="Delete rule"
			onclick={(e) => {
				e.stopPropagation();
				onDelete();
			}}
		>
			<Icon name="trash" size="sm" />
		</button>
	</div>
</div>

<style>
	.rule-card.is-error {
		background: var(--negative-bg);
	}
	.rule-card.is-error::before {
		background: var(--negative) !important;
	}
</style>
