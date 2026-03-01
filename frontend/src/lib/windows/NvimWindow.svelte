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

<div class="relative h-full">
	<div class="relative z-[1]">
		{#if win.filepath.endsWith('.md')}
			<MarkdownGrid {content} cursor={win.cursor} lineNumbers={win.lineNumbers} relativeLineNumbers={true} cursorLineColor={win.cursorLineColor} />
		{:else}
			<Grid {content} cursor={win.cursor} lineNumbers={win.lineNumbers} relativeLineNumbers={true} cursorLineColor={win.cursorLineColor} />
		{/if}
	</div>
	{#each win.colorColumns ?? [] as col (col)}
		<div
			class="color-column"
			style="left: calc({col}ch + {win.lineNumbers ? 6 : 0}ch); background: {win.colorColumnColor}"
		></div>
	{/each}
</div>

<style>
	.color-column {
		position: absolute;
		top: 0;
		bottom: 0;
		width: 1ch;
		pointer-events: none;
		opacity: 0.5;
		z-index: 0;
	}
</style>
