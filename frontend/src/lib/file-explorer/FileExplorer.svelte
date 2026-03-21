<script lang="ts">
	import { Call } from '@wailsio/runtime';
	import { onDestroy } from 'svelte';
	import { getFileExplorerState, getFileExplorerVisible } from '$lib/state.svelte';
	import Grid from '$lib/windows/Grid.svelte';

	let state = $derived(getFileExplorerState());
	let visible = $derived(getFileExplorerVisible());

	let previewPaneEl = $state<HTMLDivElement | undefined>(undefined);
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

	function syncPreviewPaneSize() {
		if (!previewPaneEl || !visible || !state?.preview) {
			return;
		}
		const rect = previewPaneEl.getBoundingClientRect();
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

	function schedulePreviewPaneSync() {
		clearPreviewResizeTimer();
		previewResizeTimer = setTimeout(syncPreviewPaneSize, PREVIEW_RESIZE_DEBOUNCE_MS);
	}

	$effect(() => {
		if (!previewPaneEl || typeof ResizeObserver === 'undefined') {
			return;
		}
		previewResizeObserver?.disconnect();
		previewResizeObserver = new ResizeObserver(() => {
			schedulePreviewPaneSync();
		});
		previewResizeObserver.observe(previewPaneEl);
		schedulePreviewPaneSync();
		return () => {
			previewResizeObserver?.disconnect();
			previewResizeObserver = undefined;
		};
	});

	$effect(() => {
		if (!visible) {
			resetPreviewDimensions();
		}
		schedulePreviewPaneSync();
	});

	onDestroy(() => {
		clearPreviewResizeTimer();
		previewResizeObserver?.disconnect();
	});
</script>

{#snippet pane(data: App.FileExplorerDirectory, role: 'parent' | 'current')}
	<div
		class="overflow-y-auto border-r-2 border-r-surface0 bg-base-100"
		class:focused={role === 'current'}
		class:parent-dir={role === 'parent'}
		class:current-dir={role === 'current'}
	>
		<!-- <div -->
		<!-- 	class="overflow-hidden text-ellipsis whitespace-nowrap border-b-2 border-b-surface0 px-1 py-2 text-xl" -->
		<!-- > -->
		<!-- 	{data.title} -->
		<!-- </div> -->
		<div class="overflow-y-auto">
			{#each data.entries as entry (entry.id)}
				<div
					class="flex cursor-default gap-1 whitespace-nowrap px-1 py-2"
					class:bg-blue-300={data.selectedEntryId === entry.id}
					class:text-black={data.selectedEntryId === entry.id}
				>
					<span class="w-[2ch] shrink-0 text-center">{entry.icon}</span>
					<span class="overflow-hidden text-ellipsis">{entry.text}</span>
				</div>
			{/each}
		</div>
	</div>
{/snippet}

{#if visible && state?.current}
	<div class="file-explorer victor-mono bg-base-100">
		<div class="panes grid h-full w-full grid-rows-1 bg-base-100">
			{#if state.parent}
				{@render pane(state.parent, 'parent')}
			{/if}
			{@render pane(state.current, 'current')}
			{#if state.preview}
				<div class="preview-pane overflow-y-auto bg-base-100" bind:this={previewPaneEl}>
					{#if state.preview.directory}
						<div class="overflow-y-auto">
							{#each state.preview.directory.entries as entry (entry.id)}
								<div
									class="flex cursor-default gap-1 whitespace-nowrap px-1 py-2"
									class:bg-blue-300={state.preview.directory.selectedEntryId === entry.id}
									class:text-black={state.preview.directory.selectedEntryId === entry.id}
								>
									<span class="w-[2ch] shrink-0 text-center">{entry.icon}</span>
									<span class="overflow-hidden text-ellipsis">{entry.text}</span>
								</div>
							{/each}
						</div>
					{:else}
						<div>
							<Grid content={state.preview.content} lineNumbers={false} relativeLineNumbers={false}
							></Grid>
						</div>
					{/if}
				</div>
			{/if}
		</div>
	</div>
{/if}

<!-- class:bg-blue-300={i + 1 === data.cursor_line} -->
<!-- class:text-black={i + 1 === data.cursor_line} -->
<!-- class:directory={entry.fs_type === 'directory'} -->

<style>
	.panes {
		grid-template-columns: 3fr 3fr 6fr;
	}
	.file-explorer {
		position: absolute;
		inset: 0;
		z-index: 150;
		pointer-events: none;
	}
</style>
