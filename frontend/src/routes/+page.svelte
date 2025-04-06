<script lang="ts">
	import { SendKey } from '$lib/wailsjs/go/main/App';
	import { onDestroy, onMount } from 'svelte';
	import StatusLine from './StatusLine.svelte';
	import { highlights, generateHighlightCSS, updateHighlightStyles } from '$lib/highlights';
	import CmdLine from './CmdLine.svelte';
	import FloatingWindowContainer from './FloatingWindowContainer.svelte';
	import NvimWindow from './NvimWindow.svelte';

	let cursor = $state<App.NvimPosition>({ row: 0, col: 1 });
	let layout = $state<App.NvimLayout>();
	let nvimWindows = $state<App.NvimWindowMap>({});
	let floatingWindows = $state<App.FloatingWindow[]>([]);
	let mode = $state<App.VimMode>('normal');

	let cmdline = $state<App.CmdLine>({
		visible: false
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
			cmdline.pos = data.pos;
		});

		runtime.EventsOn('cmdline_hide', (data) => {
			cmdline.visible = false;
		});
		runtime.EventsEmit('get-highlights');

		runtime.EventsOn('layout-updated', (updatedLayout: App.NvimLayout) => {
			console.log('updatedLayout', updatedLayout);
			console.log($state.snapshot(nvimWindows));
			layout = updatedLayout;
		});

		runtime.EventsOn('content-updated', (winId: number, updatedContent: App.NvimRow[]) => {
			console.log(`content-updated for winId=${winId}`, updatedContent);
			nvimWindows[winId] = updatedContent;
			// if (winId === 1000) {
			// 	content = updatedContent;
			// }
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
			mode = new_mode;
		});

		runtime.EventsOn('floating_windows', (windows: App.FloatingWindow[]) => {
			console.log('floating windows', windows);
			floatingWindows = windows;
		});

		runtime.EventsOn('floating_window_closed', (winId: number) => {
			console.log(`floating_window_closed id: ${winId}`);
			floatingWindows = floatingWindows.filter((win) => win.id !== winId);
		});

		runtime.EventsOn('hide-window', (winId: number) => {
			delete nvimWindows[winId];
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

<div class="victor-mono flex h-screen flex-col overflow-hidden">
	<div
		class:mode-n={mode === 'normal'}
		class:mode-v={mode === 'visual'}
		class="relative flex-grow overflow-hidden whitespace-pre"
	>
		{#if layout}
			<div
				style="grid-template-columns: {layout.cols}; grid-template-rows: {layout.rows};"
				class="grid h-full bg-surface0"
			>
				{#each layout.windows as win (win.id)}
					<div
						id="win-{win.id}"
						class:active-window={layout.activeWindowId === win.id}
						class="nvim-window relative border border-solid border-transparent bg-base-100"
						style="grid-column-start: {win.colStart}; grid-column-end:{win.colEnd}; grid-row-start: {win.rowStart}; grid-row-end:{win.rowEnd};"
					>
						<NvimWindow {nvimWindows} {win} {cursor} />
						<FloatingWindowContainer {floatingWindows} {nvimWindows} anchorWindow={win.id} />
					</div>
				{/each}
			</div>
		{/if}
		<FloatingWindowContainer {floatingWindows} {nvimWindows} anchorWindow={0} />
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
