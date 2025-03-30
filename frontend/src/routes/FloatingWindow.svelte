<script lang="ts">
	import Grid from './Grid.svelte';

	interface Position {
		top: string;
		left: string;
	}
	interface FloatingWindow {
		z_index: number;
		grid: App.NvimCell[][];
	}
	interface FloatingWindowProps {
		position: Position;
		win: FloatingWindow;
	}
	let { position, win }: FloatingWindowProps = $props();
	function decode(message: string) {
		const hexValues = message.split(',');
		let result = '';

		for (const hex of hexValues) {
			const codePoint = parseInt(hex, 16);
			result += String.fromCodePoint(codePoint);
		}
		return result;
	}
  $effect(()=>{console.log("floating grid:", win.grid)})
</script>

<div
	class="absolute overflow-hidden whitespace-pre rounded-md border border-surface0 bg-crust p-1 text-text"
	style="top: {position.top}; left: {position.left}; z-index: {win.z_index || 100};"
>
	<Grid content={win.grid} {decode} />
</div>
