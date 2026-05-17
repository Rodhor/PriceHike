<script lang="ts">
	type IconName =
		| 'upload'
		| 'download'
		| 'play'
		| 'file'
		| 'grip'
		| 'edit'
		| 'trash'
		| 'plus'
		| 'x'
		| 'sun'
		| 'moon'
		| 'alert'
		| 'chevron-down'
		| 'filter'
		| 'sheet'
		| 'loader'
		| 'settings';

	type Size = 'sm' | 'md' | 'lg';

	let { name, size = 'md', class: klass = '' }: { name: IconName; size?: Size; class?: string } =
		$props();

	const sizeClass = { sm: 'icon icon-sm', md: 'icon', lg: 'icon icon-lg' };
</script>

<svg class="{sizeClass[size]} {klass}" aria-hidden="true"><use href="#i-{name}" /></svg>
