<script lang="ts">
	import { Call } from '@wailsio/runtime';
	import { onDestroy } from 'svelte';
	import { getFileExplorerState, getFileExplorerVisible, layout } from '$lib/state.svelte';
	import LocalImage from '$lib/windows/markdown/LocalImage.svelte';
	import DirectoryPane from './DirectoryPane.svelte';
	import DirectoryPreview from './DirectoryPreview.svelte';
	import FilePreviewGrid from './FilePreviewGrid.svelte';

	let state = $derived(getFileExplorerState());
	let visible = $derived(getFileExplorerVisible());
	let activeWindow = $derived(layout.windows.find((win) => win.id === layout.activeWindowId));
	let currentWindowMode = $derived(activeWindow?.mode ?? state?.currentWinMode);

	let previewGridEl = $state<HTMLDivElement | undefined>(undefined);
	let previewResizeObserver: ResizeObserver | undefined = undefined;
	let previewResizeTimer: ReturnType<typeof setTimeout> | undefined = undefined;
	let lastPreviewWidth = -1;
	let lastPreviewHeight = -1;

	const PREVIEW_RESIZE_DEBOUNCE_MS = 75;

	function clearPreviewResizeTimer() {
		if (!previewResizeTimer) {
			return;
		}
		clearTimeout(previewResizeTimer);
		previewResizeTimer = undefined;
	}

	function resetPreviewDimensions() {
		lastPreviewWidth = -1;
		lastPreviewHeight = -1;
	}

	function syncPreviewGridSize() {
		if (!previewGridEl || !visible || state?.preview?.kind !== 'textFile') {
			return;
		}
		const rect = previewGridEl.getBoundingClientRect();
		const width = Math.max(1, Math.round(rect.width));
		const height = Math.max(1, Math.round(rect.height));
		if (width === lastPreviewWidth && height === lastPreviewHeight) {
			return;
		}
		lastPreviewWidth = width;
		lastPreviewHeight = height;
		void Call.ByName('main.App.OnFileExplorerPreviewResize', width, height).catch((err) => {
			console.error('failed syncing file explorer preview size', err);
		});
	}

	function schedulePreviewGridSync() {
		clearPreviewResizeTimer();
		previewResizeTimer = setTimeout(syncPreviewGridSize, PREVIEW_RESIZE_DEBOUNCE_MS);
	}

	$effect(() => {
		if (previewResizeObserver) {
			previewResizeObserver.disconnect();
			previewResizeObserver = undefined;
		}
		if (!previewGridEl || state?.preview?.kind !== 'textFile' || typeof ResizeObserver === 'undefined') {
			return;
		}
		previewResizeObserver = new ResizeObserver(() => {
			schedulePreviewGridSync();
		});
		previewResizeObserver.observe(previewGridEl);
		schedulePreviewGridSync();
		return () => {
			previewResizeObserver?.disconnect();
			previewResizeObserver = undefined;
		};
	});

	$effect(() => {
		if (!visible || state?.preview?.kind !== 'textFile') {
			resetPreviewDimensions();
		}
		schedulePreviewGridSync();
	});

	onDestroy(() => {
		clearPreviewResizeTimer();
		previewResizeObserver?.disconnect();
	});
</script>

{#if visible && state?.current}
	<div
		class="file-explorer victor-mono active-window bg-very-dark text-text pointer-events-none absolute inset-0 z-[150] flex flex-col"
		class:mode-i={currentWindowMode === 'insert'}
		class:mode-n={currentWindowMode === 'normal'}
		class:mode-v={currentWindowMode === 'visual'}
	>
		<div
			class="bg-very-dark border-crust flex h-12 shrink-0 items-center gap-[10px] border-b px-[14px] text-[18px]"
		>
			<span class="text-mauve font-semibold tracking-[0.1em]">EXPLORER</span>
			<span class="text-surface1">|</span>
			<span class="text-overlay0 tracking-[-0.01em]">{state.pathDisplay ?? ''}</span>
		</div>
		<div class="bg-very-dark grid min-h-0 w-full flex-1 grid-cols-[3fr_3fr_6fr] grid-rows-1">
			<DirectoryPane data={state.parent} role="parent" />
			<DirectoryPane data={state.current} role="current" />
			<div class="preview-pane bg-very-dark min-w-0 overflow-hidden">
				{#if state.preview?.kind === 'directory'}
					<DirectoryPreview title={state.preview.title} directory={state.preview.directory} />
				{:else if state.preview?.kind === 'textFile'}
					<FilePreviewGrid
						title={state.preview.title}
						content={state.preview.content}
						bind:contentEl={previewGridEl}
					/>
				{:else if state.preview?.kind === 'localImage'}
					<div class="bg-very-dark flex h-full flex-col overflow-hidden">
						<div
							class="border-mantle text-surface0 flex shrink-0 items-center gap-3 border-b-2 px-6 py-2 text-xl"
						>
							<span class="text-pink">󰋩</span>
							<span class="overflow-hidden text-ellipsis whitespace-nowrap">{state.preview.title}</span>
						</div>
						<div class="min-h-0 flex-1 p-4">
							<LocalImage
								url={state.preview.path}
								altText={state.preview.title}
								class="h-full max-h-full w-full max-w-full object-contain"
							/>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	:global(.preview-pane .flex.overflow-hidden.whitespace-pre.leading-none) {
		background-color: #181825 !important;
	}
	:global(.preview-pane .flex.overflow-hidden.whitespace-pre.leading-none span) {
		background-color: #181825 !important;
	}
</style>
