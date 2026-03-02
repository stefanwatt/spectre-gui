<script lang="ts">
	import FloatingWindow from './FloatingWindow.svelte';

	interface FloatingWindowContainerProps {
		floatingWindows: App.FloatingWindow[];
		anchorWindow: number;
		windowContentRowMap: App.WindowContentRowMap;
	}

	let { floatingWindows, anchorWindow, windowContentRowMap }: FloatingWindowContainerProps =
		$props();

	let filteredFloatingWindows = $derived(
		floatingWindows.filter((fw) => fw.anchorWindow === anchorWindow)
	);
	let anchorWindowContent = $derived(windowContentRowMap[anchorWindow] ?? []);
</script>

{#each filteredFloatingWindows as floatingWin (floatingWin.id)}
	<FloatingWindow win={floatingWin} {windowContentRowMap} {anchorWindowContent} />
{/each}
