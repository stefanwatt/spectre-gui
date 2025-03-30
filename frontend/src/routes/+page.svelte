<script lang="ts">
	import { SendKey } from '$lib/wailsjs/go/main/App';
	import { onDestroy, onMount } from 'svelte';
	import StatusLine from './StatusLine.svelte';
	import { mode } from './state.svelte';
	import { highlights, generateHighlightCSS, updateHighlightStyles } from '$lib/highlights';
	import CmdLine from './CmdLine.svelte';
	import { calculatePosition } from './window.service';
	import FloatingWindow from './FloatingWindow.svelte';
	import Grid from './Grid.svelte';

	let cursor = $state<App.NvimPosition>({ row: 0, col: 1 });
	let top_row = $state<number>(0);
	let content = $state<App.NvimCell[][]>([]);
	let floatingWindows = $state<App.FloatingWindow[]>([]);

	let cmdline = $state<App.CmdLine>({
		visible: false
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
		runtime.EventsOn('cmdline_show', (data) => {
			cmdline.visible = true;
			cmdline.content = data.content;
			cmdline.pos = data.pos;
			cmdline.firstc = data.firstc;
			cmdline.prompt = data.prompt;
			cmdline.indent = data.indent;
		});

		runtime.EventsOn('cmdline_pos', (data) => {
			console.log("pos changed:", data)
			cmdline.pos = data.pos;
		});

		runtime.EventsOn('cmdline_hide', (data) => {
			cmdline.visible = false;
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

		runtime.EventsOn('mode-changed', (new_mode: App.VimMode) => {
			console.log('mode changed', new_mode);
			$mode = new_mode;
		});
		runtime.EventsOn('floating_windows', (windows: App.FloatingWindow[]) => {
			console.log('Floating windows:', windows);
			floatingWindows = windows;
		});

		runtime.EventsOn('floating_window_closed', (winId: number) => {
			floatingWindows = floatingWindows.filter(win=>win.id === winId)
		});
	});

	onDestroy(() => {
		window.removeEventListener('keydown', send_key);
	});
</script>

{#if cmdline.visible}
	<CmdLine
		visible={true}
		firstc={cmdline.firstc}
		prompt={cmdline.prompt}
		indent={cmdline.indent}
		content={cmdline.content}
		pos={cmdline.pos}
	/>
{/if}
<div class="victor-mono flex h-screen flex-col">
	<div
		class:mode-n={$mode === 'normal'}
		class:mode-v={$mode === 'visual'}
		class="flex-grow whitespace-pre relative"
	>
		<Grid {content}/>
	{#each floatingWindows as win}
			{@const position = calculatePosition(win)}
			<FloatingWindow {position} win={win}/>
		{/each}
	</div>
	<StatusLine {cursor} mode={$mode}></StatusLine>
</div>

<style>
	.victor-mono {
		font-family: VictorMono Nerd Font Mono;
		font-size: 22px;
	}
</style>
