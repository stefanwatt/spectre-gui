<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import StatusLine from './StatusLine.svelte';
	import CmdLine from './cmdline/CmdLine.svelte';
	import FloatingWindowContainer from '$lib/floating-windows/FloatingWindowContainer.svelte';
	import NvimWindow from '$lib/windows/NvimWindow.svelte';
	import CompletionMenu from '$lib/completion/CompletionMenu.svelte';
	import {
		cmdline,
		windowContentRowMap,
		layout,
		cursor,
		nestedState,
		getFloatingWindows,
		completion
	} from '$lib/state.svelte';
	import { init, startListening, setWindowContext } from '$lib/runtime-events-service';
	import { handleKeypress, registerKeymap } from '$lib/keymaps/keymap-service';
	import { keymaps as liveGrepKeymaps } from '$lib/keymaps/live-grep';
	import { keymaps as cmdlineKeymaps } from '$lib/keymaps/cmdline';

	const params = new URLSearchParams(window.location.search);
	const myWinId = params.get('winId') ? parseInt(params.get('winId')!) : null;
	const myGridId = params.get('gridId') ? parseInt(params.get('gridId')!) : null;
	const isExternalWindow = myWinId !== null;

	// For external windows, derive a synthetic window object from content
	let externalWin = $derived.by(() => {
		if (!isExternalWindow) return null;
		// Try to find this window in layout (it may appear briefly)
		const fromLayout = layout.windows.find((w) => w.id === myWinId);
		if (fromLayout) return fromLayout;
		// Synthetic fallback — content is keyed by winId in windowContentRowMap
		return {
			id: myWinId,
			type: 'external',
			width: 100,
			height: 100,
			colStart: 1,
			colEnd: 2,
			rowStart: 1,
			rowEnd: 2,
			lineNumbers: true,
			relativeLineNumbers: true,
			floatingWindows: [],
			filetype: '',
			filepath: '',
			mode: 'normal',
			cursor: cursor,
			colorColumns: [],
			colorColumnColor: '',
			cursorLineColor: ''
		} satisfies App.NvimWindow;
	});

	onMount(async () => {
		setWindowContext(myWinId, myGridId);
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

{#if isExternalWindow}
	<!-- External window mode: single fullscreen NvimWindow -->
	<div class="victor-mono flex h-screen flex-col overflow-hidden">
		<div
			class:mode-i={externalWin?.mode === 'insert'}
			class:mode-n={externalWin?.mode === 'normal'}
			class:mode-v={externalWin?.mode === 'visual'}
			class:active-window={layout.activeWindowId === externalWin?.id}
			class="relative flex-grow overflow-hidden whitespace-pre bg-base-100"
		>
			{#if externalWin}
				<NvimWindow {windowContentRowMap} win={externalWin} />
				<FloatingWindowContainer {floatingWindows} {windowContentRowMap} anchorWindow={myWinId} />
				{#if layout.activeWindowId === myWinId}
					<CompletionMenu {completion} />
				{/if}
			{/if}
		</div>
	</div>
{:else}
	<!-- Main window mode: CSS grid layout -->
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
							class="nvim-window scrollbar-hide relative overflow-y-auto overflow-x-hidden border border-solid border-transparent bg-base-100"
							style="grid-column-start: {win.colStart}; grid-column-end:{win.colEnd}; grid-row-start: {win.rowStart}; grid-row-end:{win.rowEnd};"
						>
							<NvimWindow {windowContentRowMap} {win} />
							<FloatingWindowContainer
								{floatingWindows}
								{windowContentRowMap}
								anchorWindow={win.id}
							/>
							{#if layout.activeWindowId === win.id}
								<CompletionMenu {completion} />
							{/if}
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
{/if}

<style>
	.picker {
		height: 90vh;
		width: 90vw;
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		z-index: 200;
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
