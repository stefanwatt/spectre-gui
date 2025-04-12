<script lang="ts">
	import { SendKey, OpenFile } from '$lib/wailsjs/go/main/App';
	import { onDestroy, onMount } from 'svelte';
	import StatusLine from './StatusLine.svelte';
	import CmdLine from './CmdLine.svelte';
	import FloatingWindowContainer from './FloatingWindowContainer.svelte';
	import NvimWindow from './NvimWindow.svelte';
	import LiveGrep from '$lib/picker/LiveGrep.svelte';
	import {
		cursor_to_prev_match,
		cursor_to_next_match,
		state as resultsState
	} from '$lib/picker/results/results.service.svelte';

	let cursor = $state<App.NvimPosition>({ row: 0, col: 1 });
	let layout = $state<App.NvimLayout>();
	let nvimWindows = $state<App.NvimWindowMap>({});
	let floatingWindows = $state<App.FloatingWindow[]>([]);
	let mode = $state<App.VimMode>('normal');
	let pickers = $state({ liveGrep: false });

	let cmdline = $state<App.CmdLine>({
		visible: false
	});

	let keymapMode = $derived(
		(() => {
			if (cmdline.visible) return 'cmdline';
			if (pickers.liveGrep) return 'live-grep';
			return 'normal';
		})()
	);

	async function sendKey(e: KeyboardEvent): Promise<void> {
		if (e.key == 'Escape') {
			pickers.liveGrep = false;
			return;
		}
		if (pickers.liveGrep) {
			if (e.key == 'ArrowDown' && !e.shiftKey && !e.altKey && !e.ctrlKey) {
				cursor_to_next_match();
			}
			if (e.key == 'ArrowUp' && !e.shiftKey && !e.altKey && !e.ctrlKey) {
				cursor_to_prev_match();
			}

			if (((e.key == 'ArrowLeft'||e.key == 'ArrowRight') && !e.shiftKey && !e.altKey && e.ctrlKey) ||
      (e.key == "Enter"&& !e.shiftKey && !e.altKey && !e.ctrlKey)) {
        e.preventDefault()
				SendKey(e.key, e.ctrlKey, e.altKey, e.shiftKey, keymapMode);
			}
			return;
		}
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
		SendKey(e.key, e.ctrlKey, e.altKey, e.shiftKey, keymapMode);
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

		runtime.EventsOn(
			'live-grep-prev-page',
			(updatedResults: App.SearchResult, updatedPageIndex: number) => {
        console.log("prev page",updatedResults,updatedPageIndex)
				resultsState.results = updatedResults.GroupedMatches;
				resultsState.pageIndex = updatedPageIndex;
			}
		);

		runtime.EventsOn(
			'live-grep-next-page',
			(updatedResults: App.SearchResult, updatedPageIndex: number) => {
        console.log("next page",updatedResults,updatedPageIndex)
				resultsState.results = updatedResults.GroupedMatches;
				resultsState.pageIndex = updatedPageIndex;
			}
		);

		runtime.EventsOn('live-grep-open-selected-match', () => {
			const selectedMatch = resultsState.selectedMatch;
			if (!selectedMatch) return;
			OpenFile(selectedMatch.AbsolutePath, selectedMatch.Row, selectedMatch.Col);
		});

		runtime.EventsOn('show_live_grep', () => {
			console.log('show live grep');
			pickers.liveGrep = true;
		});

		runtime.EventsOn('hide-live-rep', () => {
			console.log('hide-live-grep');
			pickers.liveGrep = false;
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
		class:mode-i={mode === 'insert'}
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
{#if pickers.liveGrep}
	<div class="live-grep bg-darker rounded-md p-2">
		<LiveGrep />
	</div>
{/if}

<style>
	.live-grep {
		height: 90vh;
		width: 90vw;
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
	}
	.grid {
		gap: 1px;
	}
	.victor-mono {
		font-family: VictorMono Nerd Font Mono;
		font-size: 22px;
		line-height: 22px;
		font-weight: 600;
	}
</style>
