<script lang="ts">
	import Grid from './Grid.svelte';

	interface NvimWindowProps {
		win: App.NvimWindow;
		nvimWindows: App.NvimWindowMap;
		cursor: { row: number; col: number };
	}
	let { win, nvimWindows, cursor }: NvimWindowProps = $props();
	//TODO: this should not even happen
	function filterDuplicateIndices() {
		const seenIndices = new Set();
		const rows = [];
		if (!nvimWindows || !nvimWindows[win.id]) return [];
		for (let i = 0; i < nvimWindows[win.id].length; i++) {
			const row = nvimWindows[win.id][i];
			if (!seenIndices.has(row.index)) {
				seenIndices.add(row.index);
				rows.push(row);
			}
		}
		return rows;
	}
	let content = $derived(filterDuplicateIndices());
</script>

<Grid
	{content}
	{cursor}
	lineNumbers={win.lineNumbers}
	relativeLineNumbers={true}
/>
