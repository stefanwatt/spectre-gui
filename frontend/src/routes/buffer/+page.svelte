<script>
	import { SendKey } from '$lib/wailsjs/go/main/App';
	import { onDestroy, onMount } from 'svelte';

	/**@type{string[]}*/
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
	});

	onDestroy(() => {
		window.removeEventListener('keydown', send_key);
	});

	/**@param {number} line*/
	function scroll_into_view(line) {
		const lineElement = document.querySelector(`.buf-line-${line - 1}`);
		if (lineElement) {
			lineElement.scrollIntoView({ behavior: 'smooth' });
		}
	}
</script>

<div class="relative h-screen w-screen snap-y overflow-y-scroll whitespace-pre font-mono text-xl">
	{#each lines as line, index}
		<div class="snap-start buf-line-{index} whitespace-pre">
			<span>
				{#if !line}
					{'\u00A0'}
				{:else}
					{@html line}
				{/if}
			</span>
			{#if cursor.row === index}
				<span class="absolute bg-rosewater text-mantle" style="width:1ch;left: {cursor.col}ch;">
					{cursor.key}
				</span>
			{/if}
		</div>
	{/each}
</div>
