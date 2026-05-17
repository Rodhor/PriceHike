<script lang="ts">
	import { appearance } from '$lib/state/appearance.svelte';
	import Icon from './Icon.svelte';
</script>

<button
	class="btn btn-icon btn-ghost"
	type="button"
	title={appearance.theme === 'light' ? 'Switch to dark theme' : 'Switch to light theme'}
	onclick={() => appearance.toggleTheme()}
>
	<Icon name={appearance.theme === 'light' ? 'moon' : 'sun'} />
</button>
