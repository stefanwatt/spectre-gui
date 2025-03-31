<script lang="ts">
	import Grid from './Grid.svelte';

	interface Position {
		top: string;
		left: string;
	}

	interface FloatingWindowProps {
		position: Position;
		win: App.FloatingWindow;
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
</script>

<div
	id="{"win-"+win.id}"
	class="absolute overflow-hidden whitespace-pre rounded-md border border-surface0 bg-crust p-1 text-text"
	style="top: {position.top}; left: {position.left}; z-index: {win.z_index || 100};"
>
	<Grid content={win.grid} {decode} />
</div>
