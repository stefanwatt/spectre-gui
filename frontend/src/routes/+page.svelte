<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import StatusLine from './StatusLine.svelte';
	import CmdLine from './cmdline/CmdLine.svelte';
	import FloatingWindowContainer from './floating-windows/FloatingWindowContainer.svelte';
	import NvimWindow from './windows/NvimWindow.svelte';
	import LiveGrep from '$lib/picker/LiveGrep.svelte';
	import { cmdline, pickers, nvimWindows, layout, cursor, nestedState } from '$lib/state.svelte';
	import { init, startListening } from '$lib/runtime-events-service';
	import { handleKeypress, registerKeymap } from '$lib/keymaps/keymap-service';
	import { keymaps as liveGrepKeymaps } from '$lib/keymaps/live-grep';
	import { keymaps as findFilesKeymaps } from '$lib/keymaps/find-files';
	import { keymaps as cmdlineKeymaps } from '$lib/keymaps/cmdline';
	import FindFiles from '$lib/picker/FindFiles.svelte';
	import FindReferences from '$lib/picker/FindReferences.svelte';
	import FindBufferSymbols from '$lib/picker/FindBufferSymbols.svelte';

	onMount(async () => {
		await init();
		startListening();
		const allKeymaps: App.Keymap[] = [...liveGrepKeymaps, ...findFilesKeymaps, ...cmdlineKeymaps];
		allKeymaps.forEach((keymap) => {
			registerKeymap(keymap);
		});
		window.addEventListener('keydown', handleKeypress);
	});

	onDestroy(() => {
		window.removeEventListener('keydown', handleKeypress);
	});
	let floatingWindows = $derived(nestedState.floatingWindows);
	let activeWindow = $derived(layout.windows.find((win) => win.id === layout.activeWindowId));
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
	<div class="relative flex-grow overflow-hidden whitespace-pre">
		{#if layout}
			<div
				style="grid-template-columns: {layout.cols}; grid-template-rows: {layout.rows};"
				class="grid h-full bg-surface0"
			>
				{#each layout.windows as win (win.id)}
					<div
						id="win-{win.id}"
						class:mode-i={win.mode === 'insert'}
						class:mode-n={win.mode === 'normal'}
						class:mode-v={win.mode === 'visual'}
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
		<StatusLine filepath={activeWindow?.filepath} {cursor} mode={activeWindow?.mode}></StatusLine>
	</div>
</div>
{#if pickers.liveGrep}
	<div class="picker bg-darker rounded-md p-2">
		<LiveGrep />
	</div>
{/if}
{#if pickers.findFiles}
	<div class="picker bg-darker victor-mono rounded-md p-2">
		<FindFiles />
	</div>
{/if}
{#if pickers.findReferences}
	<div class="picker bg-darker victor-mono rounded-md p-2">
		<FindReferences />
	</div>
{/if}
{#if pickers.findBufferSymbols}
	<div class="picker bg-darker victor-mono rounded-md p-2">
		<FindBufferSymbols />
	</div>
{/if}
<style>
	.picker {
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
