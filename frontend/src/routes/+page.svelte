<script>
	import { SendKey } from '$lib/wailsjs/go/main/App';
	import { onDestroy, onMount } from 'svelte';
	import StatusLine from './StatusLine.svelte';
	import { scroll_into_view } from './utils.service';
	import { mode } from './state.svelte';

	/**@type{App.NvimGuiNode | undefined}*/
	let root = $state();

	/**@type{App.NvimPosition}*/
	let cursor = $state({ row: 0, col: 1 });

	let top_row = $state(0);

	/**@type{number}*/
	let container_height = $state(0);

	/**@type{number|undefined}*/
	let scroll_top = $state(0);

	/**@type{{char:string, fg:string, bg:string,classes:string}[][]}*/
	let content = $state([]);

	$effect(() => {
		console.log('top row changed', top_row);
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
		runtime.EventsOn('flush', (/**@type{any[]}*/ updated_content) => {
			console.log(updated_content)
			content=updated_content
			// if (updated_content === null) return;
			// for (let i = 0; i < updated_content.length; i++) {
			// 	const row = updated_content[i];
			// 	if (row === null) continue;
			// 	for (let j = 0; j < row.length; j++) {
			// 		const cell = row[j];
			// 		if (cell === null) continue;
			// 		if (!content[i]) content[i] = [];
			// 		content[i][j] = cell;
			// 	}
			// }
		});

		runtime.EventsOn('cursor-changed', (e) => {
			cursor = { row: e.row, col: e.col };
			top_row = e.top_line;
		});

		runtime.EventsOn('mode-changed', (new_mode) => {
			console.log("mode changed", new_mode)
			$mode = new_mode;
		});

		runtime.EventsOn('cmdline_show', () => {
			//@ts-ignore
			document.getElementById('my_modal_1').showModal();
		});

		runtime.EventsOn('cmdline_hide', () => {
			//@ts-ignore
			document.getElementById('my_modal_1').close();
		});
	});

	onDestroy(() => {
		window.removeEventListener('keydown', send_key);
	});
</script>

<dialog id="my_modal_1" class="modal">
	<input class="input input-ghost" type="text" autofocus />
</dialog>
<div class="flex h-screen flex-col font-mono">
	<div class:mode-n={$mode === 'normal'} class:mode-v={$mode === 'visual'} class="flex-grow whitespace-pre">
		{#each content as row}
			<div>
				{#each row as cell}
					<span
						class=" {cell?.classes}" style="color:{cell?.fg};background-color:{cell?.bg}">
						{cell?.char || ' '}
					</span>
				{/each}
			</div>
		{/each}
	</div>
	<StatusLine {cursor} mode={$mode}></StatusLine>
</div>

<style>
	.victor-mono {
		font-family: VictorMono Nerd Font Mono;
	}
</style>
