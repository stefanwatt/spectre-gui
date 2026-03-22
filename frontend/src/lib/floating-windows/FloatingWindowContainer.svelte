<script lang="ts">
	import FloatingWindow from './FloatingWindow.svelte';
	import type { SvelteMap } from 'svelte/reactivity';

	interface FloatingWindowContainerProps {
		floatingWindows: App.FloatingWindow[];
		anchorWindow: number;
		windowContentRowMap: SvelteMap<number, App.NvimRow[]>;
	}

	let { floatingWindows, anchorWindow, windowContentRowMap }: FloatingWindowContainerProps =
		$props();

	let filteredFloatingWindows = $derived(
		floatingWindows.filter((fw) => fw.anchorWindow === anchorWindow)
	);
	let anchorWindowContent = $derived(windowContentRowMap.get(anchorWindow) ?? []);
</script>

{#each filteredFloatingWindows as floatingWin (floatingWin.id)}
	<FloatingWindow win={floatingWin} {windowContentRowMap} {anchorWindowContent} />
{/each}
