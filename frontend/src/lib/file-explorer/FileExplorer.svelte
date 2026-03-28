<script lang="ts">
	import { Call } from '@wailsio/runtime';
	import { onDestroy } from 'svelte';
	import { getFileExplorerState, getFileExplorerVisible } from '$lib/state.svelte';
	import Grid from '$lib/windows/Grid.svelte';

	let state = $derived(getFileExplorerState());
	let visible = $derived(getFileExplorerVisible());
	let currentWindowMode = $derived(state?.currentWinMode);

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

	function currentCursorDisplayCol(entry: App.FileExplorerEntry, cursorCol: number) {
		const iconWidth = entry.icon ? 2 : 0;
		const contentStart = iconWidth + 1; // icon slot + gap
		const colInContent = Math.max(0, cursorCol - contentStart);
		return Math.min(colInContent, entry.text.length);
	}

	function splitEntryTextAtCursor(entry: App.FileExplorerEntry, cursorCol: number) {
		const cursorDisplayCol = currentCursorDisplayCol(entry, cursorCol);
		const hasCharUnderCursor = cursorDisplayCol < entry.text.length;
		return {
			before: entry.text.slice(0, cursorDisplayCol),
			cursor: hasCharUnderCursor ? entry.text.slice(cursorDisplayCol, cursorDisplayCol + 1) : '\u00a0',
			after: hasCharUnderCursor ? entry.text.slice(cursorDisplayCol + 1) : '',
			atEnd: !hasCharUnderCursor
		};
	}
</script>

{#snippet pane(data: App.FileExplorerDirectory, role: 'parent' | 'current')}
	<div
		class="bg-very-dark overflow-y-auto border-r-2 border-r-surface0"
		class:focused={role === 'current'}
		class:parent-dir={role === 'parent'}
		class:current-dir={role === 'current'}
	>
		<!-- <div -->
		<!-- 	class="overflow-hidden text-ellipsis whitespace-nowrap border-b-2 border-b-surface0 px-1 py-2 text-xl" -->
		<!-- > -->
		<!-- 	{data.title} -->
		<!-- </div> -->
		<div class="overflow-y-auto px-4">
			{#each data.entries as entry (entry.id)}
				{@const selected = data.selectedEntryId === entry.id}
				{@const isCurrentPane = role === 'current'}
				{@const cursorActive = isCurrentPane && selected}
				{@const split = cursorActive ? splitEntryTextAtCursor(entry, data.cursorCol ?? 0) : null}
				<div
					class="border-2 border-transparent rounded-lg flex cursor-default gap-1 whitespace-nowrap px-1 py-2"
					class:border-blue={selected}
				>
					<span
						class:bg-very-dark={selected}
						class={`w-[2ch] shrink-0 text-center ${entry.iconClass ?? ''}`}>{entry.icon}</span
					>
					<span class="overflow-hidden text-ellipsis">
						{#if cursorActive && split}
							<span>{split.before}</span><span class:cursor={true} class:end-cursor={split.atEnd}
								>{split.cursor}</span
							><span>{split.after}</span>
						{:else}
							{entry.text}
						{/if}
					</span>
				</div>
			{/each}
		</div>
	</div>
{/snippet}

{#if visible && state?.current}
	<div
		class="file-explorer victor-mono bg-very-dark text-text active-window"
		class:mode-i={currentWindowMode === 'insert'}
		class:mode-n={currentWindowMode === 'normal'}
		class:mode-v={currentWindowMode === 'visual'}
	>
		<div class="panes bg-very-dark grid h-full w-full grid-rows-1">
			{#if state.parent}
				{@render pane(state.parent, 'parent')}
			{/if}
			{@render pane(state.current, 'current')}
			{#if state.preview}
				<div class="preview-pane bg-very-dark overflow-y-auto" bind:this={previewPaneEl}>
					{#if state.preview.directory}
						<div class="overflow-y-auto">
							{#each state.preview.directory.entries as entry (entry.id)}
								{@const selected = state.preview.directory.selectedEntryId === entry.id}
								<div
									class="flex cursor-default gap-1 whitespace-nowrap px-1 py-2"
									class:bg-blue={selected}
									class:text-black={selected}
								>
									<span class={`w-[2ch] shrink-0 text-center ${entry.iconClass ?? ''}`}
										>{entry.icon}</span
									>
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

<style>
	.bg-blue {
		background-color: rgba(var(--ctp-blue), var(--tw-bg-opacity)) !important;
	}
	:global(.preview-pane .flex.overflow-hidden.whitespace-pre.leading-none) {
		background-color: #181825 !important;
	}
	:global(.preview-pane .flex.overflow-hidden.whitespace-pre.leading-none span) {
		background-color: #181825 !important;
	}
	.panes {
		grid-template-columns: 3fr 3fr 6fr;
	}
	.file-explorer {
		position: absolute;
		inset: 0;
		z-index: 150;
		pointer-events: none;
	}
	.end-cursor {
		display: inline-block;
		min-width: 1ch;
	}
</style>
