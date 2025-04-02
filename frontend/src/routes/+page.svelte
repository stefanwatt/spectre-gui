<script lang="ts">
	import { SendKey } from '$lib/wailsjs/go/main/App';
	import { onDestroy, onMount } from 'svelte';
	import StatusLine from './StatusLine.svelte';
	import { highlights, generateHighlightCSS, updateHighlightStyles } from '$lib/highlights';
	import CmdLine from './CmdLine.svelte';
	import { calculatePosition } from './window.service';
	import FloatingWindow from './FloatingWindow.svelte';
	import Grid from './Grid.svelte';

	let cursor = $state<App.NvimPosition>({ row: 0, col: 1 });
	let layout = $state<App.NvimLayout>();
	let nvimWindows = $state<App.NvimWindowMap>({});
	let floatingWindows = $state<App.FloatingWindow[]>([]);
	let mode = $state<App.VimMode>('normal');

	let cmdline = $state<App.CmdLine>({
		visible: false
	});

	$effect(() => {
		const css = generateHighlightCSS($highlights);
		updateHighlightStyles(css);
	});

	async function sendKey(e: KeyboardEvent): Promise<void> {
		if (e.key === 'Tab' && cmdline.visible && cmdline.content && cmdline.content.includes('s/')) {
			e.preventDefault();
			// Emit custom event for substitute field navigation
			const runtime = await import('$lib/wailsjs/runtime/runtime');
			if (runtime) {
				runtime.EventsEmit('substitute-jump', { shiftKey: e.shiftKey });
			}
			return;
		}
		e.preventDefault();
		SendKey(e.key, e.altKey, e.shiftKey, e.ctrlKey);
	}

	onMount(async () => {
		window.addEventListener('keydown', sendKey);
		const runtime = await import('$lib/wailsjs/runtime/runtime');

		window.addEventListener('resize', function () {
			runtime.EventsEmit('resize');
		});
		runtime.EventsOn('cmdline_show', (data) => {
			cmdline.visible = true;
			cmdline.content = data.content;
			cmdline.pos = data.pos;
			cmdline.firstc = data.firstc;
			cmdline.prompt = data.prompt;
			cmdline.indent = data.indent;
		});

		runtime.EventsOn('cmdline_pos', (data) => {
			console.log('pos changed:', data);
			cmdline.pos = data.pos;
		});

		runtime.EventsOn('cmdline_hide', (data) => {
			cmdline.visible = false;
		});
		runtime.EventsEmit('get-highlights');

		runtime.EventsOn('layout-updated', (updatedLayout: App.NvimLayout) => {
			console.log('updatedLayout', updatedLayout);
			layout = updatedLayout;
		});

		runtime.EventsOn('content-updated', (winId: number, updatedContent: App.NvimCell[][]) => {
			nvimWindows[winId] = updatedContent;
			// if (winId === 1000) {
			// 	content = updatedContent;
			// }
		});

		runtime.EventsOn('highlight_defined', (highlightUpdates: App.NvimHighlight[]) => {
			const updatedHighlights = { ...$highlights };
			highlightUpdates.forEach((highlight) => {
				updatedHighlights[highlight.id] = highlight;
			});
			highlights.set(updatedHighlights);
		});

		runtime.EventsOn(
			'cursor-changed',
			(e: { row: number; col: number; activeWindowId: number }) => {
				cursor = { row: e.row, col: e.col };
				if (!layout) return;
				layout.activeWindowId = e.activeWindowId;
			}
		);

		runtime.EventsOn('mode-changed', (new_mode: App.VimMode) => {
			console.log('mode changed', new_mode);
			mode = new_mode;
		});

		runtime.EventsOn('floating_windows', (windows: App.FloatingWindow[]) => {
			floatingWindows = windows;
		});

		runtime.EventsOn('floating_window_closed', (winId: number) => {
			console.log(`floating_window_closed id: ${winId}`);
			floatingWindows = floatingWindows.filter((win) => win.id !== winId);
		});
	});

	onDestroy(() => {
		window.removeEventListener('keydown', sendKey);
	});
</script>

{#if cmdline.visible}
	<div class="z-50">
		<CmdLine
			visible={true}
			firstc={cmdline.firstc}
			prompt={cmdline.prompt}
			indent={cmdline.indent}
			content={cmdline.content}
			pos={cmdline.pos}
		/>
	</div>
{/if}

<div class="victor-mono flex h-screen flex-col">
	<div
		class:mode-n={mode === 'normal'}
		class:mode-v={mode === 'visual'}
		class="relative flex-grow whitespace-pre"
	>
		{#if layout}
			<div
				style="grid-template-columns: {layout.cols}; grid-template-rows: {layout.rows};"
				class="grid h-full bg-surface0"
			>
				{#each layout.windows as win}
					<div
						id="win-{win.id}"
						class:active-window={layout.activeWindowId === win.id}
						class="nvim-window relative border border-solid border-transparent bg-base-100"
						style="grid-column-start: {win.colStart}; grid-column-end:{win.colEnd}; grid-row-start: {win.rowStart}; grid-row-end:{win.rowEnd};"
					>
						<Grid content={nvimWindows[win.id]} />
						<!-- <Grid {content} /> -->
						{#each floatingWindows.filter((fw) => fw.anchorWindow === win.id) as floatingWin}
							{@const position = calculatePosition(floatingWin)}
							<FloatingWindow {position} win={floatingWin} />
						{/each}
					</div>
				{/each}
			</div>
		{/if}
	</div>
	<div class="h-10">
		<StatusLine {cursor} {mode}></StatusLine>
	</div>
</div>

<style>
	.grid {
		gap: 1px;
	}
	.victor-mono {
		font-family: VictorMono Nerd Font Mono;
		font-size: 22px;
		line-height: 22px;
	}
</style>
