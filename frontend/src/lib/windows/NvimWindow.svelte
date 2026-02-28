<script lang="ts">
	import Grid from './Grid.svelte';
	import MarkdownGrid from './markdown/MarkdownGrid.svelte';

	interface NvimWindowProps {
		win: App.NvimWindow;
		windowContentRowMap: App.WindowContentRowMap;
	}
	let { win, windowContentRowMap }: NvimWindowProps = $props();
	let content = $derived(windowContentRowMap?.[win.id] ?? []);
</script>

{#if win.filepath.endsWith('.md')}
	<MarkdownGrid {content} cursor={win.cursor} lineNumbers={win.lineNumbers} relativeLineNumbers={true} />
{:else}
	<Grid {content} cursor={win.cursor} lineNumbers={win.lineNumbers} relativeLineNumbers={true} />
{/if}
