<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import StatusLine from '../../StatusLine.svelte';
	import CmdLine from '../../cmdline/CmdLine.svelte';
	import FloatingWindowContainer from '$lib/floating-windows/FloatingWindowContainer.svelte';
	import NvimWindow from '$lib/windows/NvimWindow.svelte';
	import CompletionMenu from '$lib/completion/CompletionMenu.svelte';
	import {
		cmdline,
		windowContentRowMap,
		cursor,
		nestedState,
		getFloatingWindows,
		completion
	} from '$lib/state.svelte';
	import type { PageProps } from './$types';
	import { init, startListening, setWindowContext } from '$lib/runtime-events-service';
	import { handleKeypress, registerKeymap } from '$lib/keymaps/keymap-service';
	import { keymaps as liveGrepKeymaps } from '$lib/keymaps/live-grep';
	import { keymaps as cmdlineKeymaps } from '$lib/keymaps/cmdline';

	let { data }: PageProps = $props();

	let filepath = $state(data.window.filepath);
	let mode = $state<App.VimMode>(data.window.mode);

	function onBufEnter(e: Event) {
		filepath = (e as CustomEvent).detail;
	}
	function onModeChanged(e: Event) {
		mode = (e as CustomEvent).detail;
	}

	onMount(async () => {
		setWindowContext(data.window.id, data.gridId);
		await init();
		startListening();
		const allKeymaps: App.Keymap[] = [...liveGrepKeymaps, ...cmdlineKeymaps];
		allKeymaps.forEach((keymap) => {
			registerKeymap(keymap);
		});
		window.addEventListener('keydown', handleKeypress);
		window.addEventListener('nvim-bufenter', onBufEnter);
		window.addEventListener('nvim-mode-changed', onModeChanged);
	});

	onDestroy(() => {
		window.removeEventListener('keydown', handleKeypress);
		window.removeEventListener('nvim-bufenter', onBufEnter);
		window.removeEventListener('nvim-mode-changed', onModeChanged);
	});

	let floatingWindows = $derived(getFloatingWindows());
	let isActiveWindow = $derived(cursor.activeWindowId === data.window.id);
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
	<div
		class:mode-i={mode === 'insert'}
		class:mode-n={mode === 'normal'}
		class:mode-v={mode === 'visual'}
		class:active-window={isActiveWindow}
		class="bg-base-100 relative flex-grow overflow-hidden whitespace-pre"
		id="win-{data.window.id}"
	>
		<NvimWindow {windowContentRowMap} win={data.window} />
		<FloatingWindowContainer
			{floatingWindows}
			{windowContentRowMap}
			anchorWindow={data.window.id}
		/>
		{#if isActiveWindow}
			<CompletionMenu {completion} />
		{/if}
	</div>
	<div class="h-10">
		<StatusLine {filepath} {cursor} {mode} />
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
		z-index: 200;
	}
	.victor-mono {
		font-family: VictorMono Nerd Font Mono;
		font-size: 22px;
		line-height: 22px;
		font-weight: 600;
	}
</style>
