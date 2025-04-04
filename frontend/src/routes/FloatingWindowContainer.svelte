<script lang="ts">
	import FloatingWindow from './FloatingWindow.svelte';

	interface FloatingWindowContainerProps {
		floatingWindows: App.FloatingWindow[];
		anchorWindow: number;
		nvimWindows: App.NvimWindowMap;
	}

	let { floatingWindows, anchorWindow, nvimWindows }: FloatingWindowContainerProps = $props();

	let filteredFloatingWindows = $derived(
		floatingWindows.filter((fw) => fw.anchorWindow === anchorWindow)
	);
</script>

{#each filteredFloatingWindows as floatingWin}
	<FloatingWindow  win={floatingWin} {nvimWindows} />
{/each}
