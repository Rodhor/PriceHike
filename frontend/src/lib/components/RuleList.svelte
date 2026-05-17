<script lang="ts">
	import { dndzone, type DndEvent } from 'svelte-dnd-action';
	import { flip } from 'svelte/animate';
	import type { Rule, RuleStat } from '$lib/backend';
	import Icon from './Icon.svelte';
	import RuleCard from './RuleCard.svelte';

	type Props = {
		rules: Rule[];
		priceHeaders: string[];
		selectedId: string | null;
		// Map of rule.id → its RuleStat from the most recent preview. Empty
		// before the first preview; lookups for unknown IDs are undefined.
		statByRuleId: Map<string, RuleStat>;
		// 1-based index of the rule the last error pointed at, if any. Used
		// to paint that card in the error variant.
		erroredRuleIndex: number | null;
		onSelect: (id: string) => void;
		onAdd: () => void;
		onDelete: (id: string) => void;
		onReorder: (next: Rule[]) => void;
	};

	let {
		rules,
		priceHeaders,
		selectedId,
		statByRuleId,
		erroredRuleIndex,
		onSelect,
		onAdd,
		onDelete,
		onReorder
	}: Props = $props();

	const flipMs = 150;

	function handleConsider(e: CustomEvent<DndEvent<Rule>>) {
		onReorder(e.detail.items);
	}

	function handleFinalize(e: CustomEvent<DndEvent<Rule>>) {
		onReorder(e.detail.items);
	}
</script>

<section class="rule-list-panel">
	<div class="panel-header">
		<div style="display: flex; align-items: center; gap: 8px;">
			<span class="uppercase-label">Rules</span>
			<span
				class="badge badge-neutral mono"
				style="font-family: var(--font-mono); font-weight: 500;"
			>
				{rules.length}
			</span>
		</div>
		<button class="btn btn-sm" type="button" onclick={onAdd}>
			<Icon name="plus" size="sm" />
			Add rule
		</button>
	</div>

	<div class="panel-body">
		{#if rules.length === 0}
			<div
				style="display: grid; place-items: center; height: 100%; padding: 24px; text-align: center;"
			>
				<div style="display: grid; gap: 8px;">
					<p class="t-tertiary" style="font-size: var(--text-sm); margin: 0;">No rules yet.</p>
					<button
						class="btn btn-primary"
						type="button"
						onclick={onAdd}
						style="justify-self: center;"
					>
						<Icon name="plus" size="sm" />
						Add rule
					</button>
				</div>
			</div>
		{:else}
			<div
				use:dndzone={{ items: rules, flipDurationMs: flipMs, dropTargetStyle: {} }}
				onconsider={handleConsider}
				onfinalize={handleFinalize}
			>
				{#each rules as rule, i (rule.id)}
					<div animate:flip={{ duration: flipMs }}>
						<RuleCard
							{rule}
							ordinal={i + 1}
							{priceHeaders}
							selected={rule.id === selectedId}
							stat={statByRuleId.get(rule.id)}
							errored={erroredRuleIndex === i + 1}
							onSelect={() => onSelect(rule.id)}
							onDelete={() => onDelete(rule.id)}
						/>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</section>
