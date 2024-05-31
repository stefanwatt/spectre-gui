<script>
	import { SendKey } from '$lib/wailsjs/go/main/App';
	import { onDestroy, onMount } from 'svelte';
	import StatusLine from './StatusLine.svelte';

	/**@type{App.VimMode}*/
	let mode = 'n';

	/**@type{App.BufLine[]}*/
	let lines = [];

	let cursor = { row: 5, col: 5, key: ' ' };

	function send_key(e) {
		SendKey(e.key, e.ctrlKey, e.shiftKey, e.altKey);
	}

	/**@param {App.CursorMoveEvent} e*/
	function on_cursor_moved(e) {
		cursor = { ...cursor, row: e.row, col: e.col, key: e.key };
		lines = lines;
		scroll_into_view(e.top_line);
	}

	onMount(async () => {
		window.addEventListener('keydown', send_key);
		const runtime = await import('$lib/wailsjs/runtime/runtime');
		runtime.EventsOn('buf-lines-changed', (updated_lines) => {
			lines = updated_lines;
		});

		runtime.EventsOn('cursor-changed', on_cursor_moved);
		runtime.EventsOn('mode-changed', (new_mode) => {
			mode = new_mode;
		});
	});

	onDestroy(() => {
		window.removeEventListener('keydown', send_key);
	});

	/**@param {number} line*/
	function scroll_into_view(line) {
		const lineElement = document.querySelector(`.buf-line-${line - 1}`);
		if (lineElement) {
			lineElement.scrollIntoView();
		}
	}
</script>

<div class="flex h-screen flex-col">
	<div
		class="grid h-full w-screen grow snap-y auto-rows-min grid-cols-[4rem,auto] overflow-y-scroll whitespace-pre font-mono text-xl"
	>
		{#each lines as buf_line, index}
			<div class="text-overlay0">
				<span class="text-right">{buf_line.sign}</span>
				<span class="ml-1 text-right">{buf_line.row}</span>
			</div>
			<div class="relative ml-8 snap-start buf-line-{index} whitespace-pre">
				<span>
					{#if !buf_line.line}
						{'\u00A0'}
					{:else}
						{@html buf_line.line}
					{/if}
				</span>
				{#if cursor.row === index}
					<span
						class:bg-rosewater={mode !== 'i'}
						class:border-l={mode === 'i'}
						class:bg-transparent={mode === 'i'}
						class="absolute border-rosewater bg-rosewater text-mantle"
						style="width:1ch;left: {cursor.col}ch;"
					>
						{#if mode !== 'i'}
							{cursor.key}
						{:else}
							⠀
						{/if}
					</span>
				{/if}
			</div>
		{/each}
	</div>
	<div class="h-8">
		<StatusLine {mode}></StatusLine>
	</div>
</div>
