<script lang="ts">
	import FloatingGrid from './FloatingGrid.svelte';
	import { calculatePosition } from './window.service';

	interface Position {
		top: string;
		left: string;
	}

	interface FloatingWindowProps {
		win: App.FloatingWindow;
		nvimWindows: App.NvimWindowMap;
	}
	let { win, nvimWindows }: FloatingWindowProps = $props();

	let content = $derived(nvimWindows[win.id]);
	let position: Position = $derived(calculatePosition(win.row, win.col, win.filetype));
	function decode(message: string) {
		if (!win.isHex) return message;
		let result = '';
		try {
			const hexValues = message.split(',');

			for (const hex of hexValues) {
				const codePoint = parseInt(hex, 16);
				result += String.fromCodePoint(codePoint);
			}
		} catch (error) {
			return message;
		}
		return result;
	}
</script>

<div
	id={'win-' + win.id}
	class="floating-win absolute overflow-hidden whitespace-pre rounded-md border border-surface0 bg-base-100 text-text drop-shadow-md filetype-{win.filetype}"
	style="top: {position.top}; left: {position.left}; z-index: {win.zIndex || 100};"
>
	<FloatingGrid {content} {decode} />
</div>

<style>
	.floating-win {
		font-size: 19px;
	}
</style>
