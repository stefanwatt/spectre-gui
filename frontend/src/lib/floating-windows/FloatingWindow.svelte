<script lang="ts">
	import FloatingGrid from './FloatingGrid.svelte';
	import { calculatePosition } from '../windows/window.service';

	interface Position {
		top: string;
		left: string;
	}

	import type { SvelteMap } from 'svelte/reactivity';

	interface FloatingWindowProps {
		win: App.FloatingWindow;
		windowContentRowMap: SvelteMap<number, App.NvimRow[]>;
		anchorWindowContent: App.NvimRow[];
	}
	let { win, windowContentRowMap, anchorWindowContent }: FloatingWindowProps = $props();

	let content = $derived(windowContentRowMap.get(win.id));

	// Map neovim grid row → buffer line number via the anchor window's content rows.
	// This accounts for markdown rich elements that render at variable heights.
	let anchorBufferLine = $derived(anchorWindowContent?.[win.row]?.index ?? null);

	let domTop: number | null = $state(null);

	$effect(() => {
		if (anchorBufferLine === null) { domTop = null; return; }
		const el = document.getElementById(`row-${anchorBufferLine}`);
		if (el) {
			domTop = el.offsetTop;
		} else {
			domTop = null;
		}
	});

	let position: Position = $derived.by(() => {
		const fallback = calculatePosition(win.row, win.col, win.filetype, win.anchor, win.height);
		if (domTop !== null) {
			const anchor = win.anchor ?? 'NW';
			const lineHeight = 28; // 22px font + 3px padding top + 3px padding bottom
			const topPx = anchor.startsWith('S') ? domTop - win.height * lineHeight : domTop;
			const nonOffsetFiletypes = new Set(['treesitter_context', 'wk', 'fidget']);
			const offsetLeft = nonOffsetFiletypes.has(win.filetype) ? 0 : 6;
			return { top: `${topPx}px`, left: `${win.col + offsetLeft}ch` };
		}
		return fallback;
	});

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
	<FloatingGrid {content} {decode} lineNumbers={false} relativeLineNumbers={false} />
</div>

<style>
	.floating-win {
		font-size: 19px;
	}
</style>
