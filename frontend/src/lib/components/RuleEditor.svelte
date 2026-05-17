<script lang="ts">
	import type {
		DecimalSettings,
		Rule,
		RuleAction,
		RuleFilter
	} from '$lib/backend';
	import ActionPicker from './ActionPicker.svelte';
	import ColumnSelector from './ColumnSelector.svelte';
	import DecimalSettingsField from './DecimalSettingsField.svelte';
	import EmptyState from './EmptyState.svelte';
	import FilterFieldset from './FilterFieldset.svelte';
	import Icon from './Icon.svelte';

	type Props = {
		rule: Rule | null;
		priceHeaders: string[];
		onChange: (next: Rule) => void;
		onDelete: (id: string) => void;
		onAdd: () => void;
	};

	let { rule, priceHeaders, onChange, onDelete, onAdd }: Props = $props();

	function setFilter(next: RuleFilter) {
		if (!rule) return;
		onChange({ ...rule, filter: next });
	}

	function setColumns(next: number[]) {
		if (!rule) return;
		onChange({ ...rule, columns: next });
	}

	function setAction(next: RuleAction) {
		if (!rule) return;
		onChange({ ...rule, action: next });
	}

	function setDecimals(next: DecimalSettings) {
		if (!rule) return;
		onChange({ ...rule, decimalSettings: next });
	}
</script>

<section class="rule-editor-panel">
	<div class="panel-header">
		<span class="uppercase-label">Editor</span>
		{#if rule}
			<button
				class="btn btn-sm btn-danger-ghost"
				type="button"
				onclick={() => rule && onDelete(rule.id)}
				title="Delete this rule"
			>
				<Icon name="trash" size="sm" />
				Delete
			</button>
		{/if}
	</div>

	<div class="panel-body">
		{#if !rule}
			<EmptyState
				variant="no-rule-selected"
				primaryAction={{ label: 'Add rule', onClick: onAdd }}
			/>
		{:else}
			<div class="editor">
				<FilterFieldset
					filter={rule.filter}
					{priceHeaders}
					onChange={setFilter}
				/>
				<ColumnSelector
					{priceHeaders}
					selected={rule.columns ?? []}
					onChange={setColumns}
				/>
				<ActionPicker action={rule.action} onChange={setAction} />
				<DecimalSettingsField settings={rule.decimalSettings} onChange={setDecimals} />
			</div>
		{/if}
	</div>
</section>
