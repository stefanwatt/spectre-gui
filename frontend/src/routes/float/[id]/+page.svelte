<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import FloatingGrid from '$lib/floating-windows/FloatingGrid.svelte';
	import { windowContentRowMap } from '$lib/state.svelte';
	import type { PageProps } from './$types';
	import { init, startListening, setWindowContext } from '$lib/runtime-events-service';
	import { handleKeypress, registerKeymap } from '$lib/keymaps/keymap-service';
	import { keymaps as liveGrepKeymaps } from '$lib/keymaps/live-grep';
	import { keymaps as cmdlineKeymaps } from '$lib/keymaps/cmdline';

	let { data }: PageProps = $props();

	onMount(async () => {
		setWindowContext(data.winId, data.gridId);
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

	let content = $derived(windowContentRowMap.get(data.winId) ?? []);
</script>
<svelte:head>
	<title>nvim-gui-float</title>
</svelte:head>

<div class="victor-mono h-screen w-screen overflow-hidden whitespace-pre">
	<FloatingGrid {content} lineNumbers={false} relativeLineNumbers={false} />
</div>

<style>
	.victor-mono {
		font-family: VictorMono Nerd Font Mono;
		font-size: 22px;
		line-height: 22px;
		font-weight: 600;
	}
</style>
