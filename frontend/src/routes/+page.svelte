<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import StatusLine from './StatusLine.svelte';
	import CmdLine from './cmdline/CmdLine.svelte';
	import FloatingWindowContainer from '$lib/floating-windows/FloatingWindowContainer.svelte';
	import NvimWindow from '$lib/windows/NvimWindow.svelte';
	import { cmdline, windowContentRowMap, layout, cursor, nestedState, getFloatingWindows } from '$lib/state.svelte';
	import { init, startListening } from '$lib/runtime-events-service';
	import { handleKeypress, registerKeymap } from '$lib/keymaps/keymap-service';
	import { keymaps as liveGrepKeymaps } from '$lib/keymaps/live-grep';
	import { keymaps as cmdlineKeymaps } from '$lib/keymaps/cmdline';

	onMount(async () => {
		await init();
		startListening();
		const allKeymaps: App.Keymap[] = [...liveGrepKeymaps, ...cmdlineKeymaps];
		allKeymaps.forEach((keymap) => {
			registerKeymap(keymap);
		});
		window.addEventListener('keydown', handleKeypress);
	});

	onDestroy(() => {
		window.removeEventListener('keydown', handleKeypress);
	});
	let floatingWindows = $derived(getFloatingWindows());
	let activeWindow = $derived(layout.windows.find((win) => win.id === layout.activeWindowId));
	$effect(() => {
		//NOTE: we dont want to rely on redraw events for cursor updates
		//so we have a separate event listener for cursor updates
		if (!activeWindow) return;
		activeWindow.cursor = cursor;
	});
	let PickerComponent = $derived(nestedState.activePicker?.component);
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
						class="nvim-window relative border border-solid border-transparent bg-base-100 overflow-x-hidden overflow-y-auto scrollbar-hide"
						style="grid-column-start: {win.colStart}; grid-column-end:{win.colEnd}; grid-row-start: {win.rowStart}; grid-row-end:{win.rowEnd};"
					>
						<NvimWindow {windowContentRowMap} {win} />
						<FloatingWindowContainer
							{floatingWindows}
							{windowContentRowMap}
							anchorWindow={win.id}
						/>
					</div>
				{/each}
			</div>
		{/if}
		<FloatingWindowContainer {floatingWindows} {windowContentRowMap} anchorWindow={0} />
	</div>
	<div class="h-10">
		<StatusLine filepath={activeWindow?.filepath} {cursor} mode={activeWindow?.mode}></StatusLine>
	</div>
</div>

{#if PickerComponent}
	<div class="picker bg-darker victor-mono rounded-md p-2">
		<PickerComponent />
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
	.scrollbar-hide {
		scrollbar-width: none;
	}
	.scrollbar-hide::-webkit-scrollbar {
		display: none;
	}
</style>
