<script lang="ts">
	import {
		ACTION_LABELS,
		ACTION_TYPES,
		isRoundingAction,
		type ActionType,
		type RuleAction
	} from '$lib/backend';

	type Props = {
		action: RuleAction;
		onChange: (next: RuleAction) => void;
	};

	let { action, onChange }: Props = $props();

	const isPercent = $derived(action.type === 'addPercent' || action.type === 'subtractPercent');
	const isFixed = $derived(action.type === 'addFixed' || action.type === 'subtractFixed');
	const isRounding = $derived(isRoundingAction(action.type));

	function setType(t: ActionType) {
		// Rounding actions ignore Action.Value per HANDOFF §4.5 but we still keep
		// a non-zero value so backend's RequiresValue check stays happy. On switch,
		// pick a sensible default when none is set.
		const value = action.value && action.value.length > 0 ? action.value : '0';
		onChange({ type: t, value });
	}

	function setValue(e: Event) {
		const v = (e.target as HTMLInputElement).value;
		onChange({ ...action, value: v });
	}
</script>

<div class="full">
	<h3>Action</h3>

	<div class="segmented" style="display: flex; flex-wrap: wrap;">
		{#each ACTION_TYPES as t (t)}
			<button
				type="button"
				class={action.type === t ? 'is-active' : ''}
				onclick={() => setType(t)}
			>
				{ACTION_LABELS[t]}
			</button>
		{/each}
	</div>

	{#if !isRounding}
		<label
			class="field-label"
			for="action-value"
			style="margin-top: 10px; display: flex; align-items: center; gap: 6px;"
		>
			<span>Amount</span>
			{#if isPercent}
				<span class="t-tertiary" style="font-weight: 400;">(percent)</span>
			{:else if isFixed}
				<span class="t-tertiary" style="font-weight: 400;">(fixed amount)</span>
			{/if}
		</label>
		<div style="display: flex; align-items: center; gap: 6px;">
			<input
				id="action-value"
				class="field field-mono"
				type="text"
				inputmode="decimal"
				placeholder={isPercent ? '5' : '0.00'}
				value={action.value}
				oninput={setValue}
				style="max-width: 160px;"
			/>
			{#if isPercent}
				<span class="t-secondary mono">%</span>
			{/if}
		</div>
	{:else}
		<p class="t-tertiary" style="font-size: var(--text-xs); margin: 8px 0 0 0;">
			Rounding actions take their precision from the Decimals settings below.
		</p>
	{/if}
</div>
