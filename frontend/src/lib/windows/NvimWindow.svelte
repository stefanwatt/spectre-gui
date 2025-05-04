<script lang="ts">
	import Grid from './Grid.svelte';
	import MarkdownGrid from './markdown/MarkdownGrid.svelte';

	interface NvimWindowProps {
		win: App.NvimWindow;
		windowContentRowMap: App.WindowContentRowMap;
	}
	let { win, windowContentRowMap }: NvimWindowProps = $props();
	//TODO: this should not even happen
	function filterDuplicateIndices() {
		const seenIndices = new Set();
		const rows = [];
		if (!windowContentRowMap || !windowContentRowMap[win.id]) return [];
		for (let i = 0; i < windowContentRowMap[win.id].length; i++) {
			const row = windowContentRowMap[win.id][i];
			if (!seenIndices.has(row.index)) {
				seenIndices.add(row.index);
				rows.push(row);
			}
		}
		return rows;
	}
	let content = $derived(filterDuplicateIndices());
</script>

{#if win.filepath.endsWith('.md')}
	<MarkdownGrid {content} cursor={win.cursor} lineNumbers={win.lineNumbers} relativeLineNumbers={true} />
{:else}
	<Grid {content} cursor={win.cursor} lineNumbers={win.lineNumbers} relativeLineNumbers={true} />
{/if}
