<script lang="ts">
	import { SendKey } from '$lib/wailsjs/go/main/App';
	import { onDestroy, onMount } from 'svelte';
	import StatusLine from './StatusLine.svelte';
	import { scroll_into_view } from './utils.service';
	import { mode } from './state.svelte';
	import { highlights, generateHighlightCSS, updateHighlightStyles } from '$lib/highlights';

	let cursor = $state<App.NvimPosition>({ row: 0, col: 1 });
	let top_row = $state<number>(0);
	let scroll_top = $state<number | undefined>(0);
	let content = $state<App.NvimCell[][]>([]);


	let cmdlineVisible = $state(false);
	let cmdlineContent = $state('');
	let cmdlinePos = $state(0);
	let cmdlineFirstc = $state('');
	let cmdlinePrompt = $state('');
	let cmdlineIndent = $state(0);

	$effect(() => {
		console.log('top row changed', top_row);
		if (!top_row) return;
		scroll_top = top_row * 28;
		scroll_into_view(top_row);
	});

	$effect(() => {
		const css = generateHighlightCSS($highlights);
		updateHighlightStyles(css);
	});

	function send_key(e: KeyboardEvent): void {
		e.preventDefault();
		SendKey(e.key, e.altKey, e.shiftKey, e.ctrlKey);
	}

	onMount(async () => {
		window.addEventListener('keydown', send_key);
		const runtime = await import('$lib/wailsjs/runtime/runtime');
		// Cmdline event handlers
		runtime.EventsOn('cmdline_show', (data) => {
			console.log("cmdline-show",data)
			cmdlineVisible = true;
			cmdlineContent = data.content;
			cmdlinePos = data.pos;
			cmdlineFirstc = data.firstc;
			cmdlinePrompt = data.prompt;
			cmdlineIndent = data.indent;
		});

		runtime.EventsOn('cmdline_pos', (data) => {
			cmdlinePos = data.pos;
		});

		runtime.EventsOn('cmdline_hide', (data) => {
			cmdlineVisible = false;
		});
		runtime.EventsEmit('get-highlights');

		runtime.EventsOn('flush', (updated_content: App.NvimCell[][]) => {
			content = updated_content;
		});

		runtime.EventsOn('highlight_defined', (highlightUpdates: App.NvimHighlight[]) => {
			const updatedHighlights = { ...$highlights };
			highlightUpdates.forEach((highlight) => {
				updatedHighlights[highlight.id] = highlight;
			});
			highlights.set(updatedHighlights);
		});

		runtime.EventsOn('cursor-changed', (e: { row: number; col: number; top_line: number }) => {
			cursor = { row: e.row, col: e.col };
			top_row = e.top_line;
		});

		runtime.EventsOn('mode-changed', (new_mode: string) => {
			console.log('mode changed', new_mode);
			$mode = new_mode;
		});
	});

	onDestroy(() => {
		window.removeEventListener('keydown', send_key);
	});
</script>

<div class="victor-mono flex h-screen flex-col">
	<div
		class:mode-n={$mode === 'normal'}
		class:mode-v={$mode === 'visual'}
		class="flex-grow whitespace-pre"
	>
		{#each content as row}
			<div class="flex overflow-hidden leading-none">
				{#each row as cell}
					{#if cell?.char}
						<span
							class="cell inline-block h-full hl-{cell?.highlight} {cell?.classes}"
							style="color:{cell?.fg};background-color:{cell?.bg}"
						>
							{cell.char}
						</span>
					{/if}
				{/each}
			</div>
		{/each}
	</div>
	{#if cmdlineVisible}
		<div class="cmdline-container flex items-center border-t border-gray-700 p-1">
			{#if cmdlineFirstc}
				<span class="cmdline-firstc mr-1">{cmdlineFirstc}</span>
			{/if}
			{#if cmdlinePrompt}
				<span class="cmdline-prompt mr-1 ">{cmdlinePrompt}</span>
			{/if}
			{#if cmdlineIndent > 0}
				<span class="cmdline-indent">{' '.repeat(cmdlineIndent)}</span>
			{/if}
			<div class="cmdline-content flex">
				{cmdlineContent}
			</div>
		</div>
	{/if}
	<StatusLine {cursor} mode={$mode}></StatusLine>
</div>

<style>
	.cell {
		padding-top: 3px;
		padding-bottom: 3px;
	}
	.victor-mono {
		font-family: VictorMono Nerd Font Mono;
		font-size: 22px;
	}
</style>
