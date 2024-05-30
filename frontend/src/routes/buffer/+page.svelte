<script>
	import { SendKey } from '$lib/wailsjs/go/main/App';
	import { onMount } from 'svelte';

	/**@type{string[]}*/
	let lines = [];

	let cursor = { row: 5, col: 5, key: ' ' };

	function send_key(e) {
		SendKey(e.key, e.ctrlKey, e.shiftKey, e.altKey);
	}

	onMount(async () => {
		const runtime = await import('$lib/wailsjs/runtime/runtime');
		runtime.EventsOn('buf-lines-changed', (updated_lines) => {
			lines = updated_lines;
		});

		runtime.EventsOn('cursor-changed', (row, col, key) => {
			cursor = { row: +row, col: col + 1, key };
			lines = lines;
		});
	});
</script>

<input autofocus class="hidden" on:keydown|preventDefault={send_key} type="text" />
<div class="relative h-screen w-screen overflow-y-scroll whitespace-pre">
	{#each lines as line, index}
		{#if !line}
			<div class="whitespace-pre">{'\u00A0'}</div>
		{:else}
			<div class="whitespace-pre">
				<span>
					<span>
						{@html line}
					</span>
					{#if cursor.row === index}
						<span class="absolute bg-rosewater text-mantle" style="width:1ch;left: {cursor.col}ch;"
							>{cursor.key}</span
						>
					{/if}
				</span>
			</div>
		{/if}
	{/each}
</div>
