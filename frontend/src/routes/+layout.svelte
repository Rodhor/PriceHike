<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { appearance } from '$lib/state/appearance.svelte';
	import { onMount } from 'svelte';

	let { children } = $props();

	// Appearance preferences (palette + theme) need to be applied before the
	// first user-visible frame, otherwise the UI flashes the default indigo /
	// light theme before resolving to whatever the user chose. Doing it in
	// onMount is still client-side-only thanks to ssr:false in +layout.ts.
	onMount(() => {
		appearance.init();

		// Webview navigates to a dropped file unless we cancel the DOM default.
		// Wails' OS-level OnFileDrop is a separate channel — cancelling here
		// only stops the browser from opening the file as a page; the OS-level
		// handler still receives the drop.
		const swallow = (e: DragEvent) => e.preventDefault();
		window.addEventListener('dragover', swallow);
		window.addEventListener('drop', swallow);
		return () => {
			window.removeEventListener('dragover', swallow);
			window.removeEventListener('drop', swallow);
		};
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<main>
	{@render children()}
</main>
