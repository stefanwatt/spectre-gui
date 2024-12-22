<script>
	import { SendKey } from '$lib/wailsjs/go/main/App';
	import { onDestroy, onMount } from 'svelte';
	import { update_cursor } from './cursor.service';
	import StatusLine from './StatusLine.svelte';
	import VirtualList from './VirtualList.svelte';
	import { scroll_into_view } from './utils.service';
	import { clear_highlights, highlight_lines, highlight_range } from './visual-selection.service';
	import Node from './Node.svelte';

	/**@type{App.VimMode}*/
	let mode = $state('n');

	/**@type{boolean}*/
	//@ts-ignore
	let visual_selection_active = $derived(mode === 'v' || mode === 'V');

	/**@type{App.NvimGuiNode | undefined}*/
	let root = $state();

	/**@type{App.NvimRange}*/
	let selection_range = $state({
		start_row: -1,
		end_row: -1,
		start_col: 0,
		end_col: 0
	});

	// $effect(() => {
	// 	if (!visual_selection_active) {
	// 		clear_highlights();
	// 		return;
	// 	}
	// 	if (mode === 'V') {
	// 		highlight_lines(selection_range.start_row, selection_range.end_row);
	// 		return;
	// 	}
	// 	if (mode === 'v') {
	// 		highlight_range(selection_range, buf_lines);
	// 	}
	// });

	/**@type{App.NvimPosition}*/
	let cursor = $state({ row: 0, col: 1 });
	let top_row = $state(0);

	// $effect(() => {
	// 	console.log('cursor effect');
	// 	if (!buf_lines?.length || !cursor) return;
	// 	const row = buf_lines.find((buf_line) => buf_line.row === cursor.row);
	// 	if (!row?.tokens?.length) return;
	// 	const line_end_col = row.tokens.slice(-1)[0].end_col;
	// 	update_cursor({ row: cursor.row, col: cursor.col }, line_end_col, mode);
	// });
	/**@type{number}*/
	let container_height = $state(0);

	/**@type{number|undefined}*/
	let scroll_top = $state(0);
	$effect(() => {
		if (!top_row) return;
		scroll_top = top_row * 28;
		scroll_into_view(top_row);
	});
	/**@param {KeyboardEvent} e*/
	function send_key(e) {
		e.preventDefault();
		SendKey(e.key, e.altKey, e.shiftKey, e.ctrlKey);
	}

	onMount(async () => {
		window.addEventListener('keydown', send_key);
		const runtime = await import('$lib/wailsjs/runtime/runtime');
		runtime.EventsOn('buf-lines-changed', (updated_root) => {
			root = updated_root;
			console.log('updated tree:', updated_root);
		});

		runtime.EventsOn('cursor-changed', (e) => {
			console.log('cursor-changed event');
			cursor = { row: e.row, col: e.col };
			top_row = e.top_line - 1;
		});

		runtime.EventsOn('mode-changed', (new_mode) => {
			mode = new_mode;
		});
		runtime.EventsOn('visual-selection-changed', (new_selection_range) => {
			selection_range = new_selection_range;
			console.log('new_selection_range', new_selection_range);
		});
	});

	onDestroy(() => {
		window.removeEventListener('keydown', send_key);
	});
</script>

<div class="flex h-screen flex-col">
	<div
		bind:clientHeight={container_height}
		class="scr-top-{scroll_top} h-full w-screen grow snap-y auto-rows-min grid-cols-[4rem,auto] gap-0 overflow-y-scroll whitespace-pre font-mono text-xl"
	>
		{#if root?.children?.length}
			<Node children={root.children}></Node>
		{/if}
	</div>
	<StatusLine {mode}></StatusLine>
</div>

<style>
	.victor-mono {
		font-family: VictorMono Nerd Font Mono;
	}
</style>
