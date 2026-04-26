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
		const hiddenPrefixWidth = String(entry.id).length + 1; // "<id>/"
		const colInContent = Math.max(0, cursorCol - hiddenPrefixWidth);
		return Math.min(colInContent, entry.text.length);
	}

	function splitEntryTextAtCursor(entry: App.FileExplorerEntry, cursorCol: number) {
		const cursorDisplayCol = currentCursorDisplayCol(entry, cursorCol);
		const hasCharUnderCursor = cursorDisplayCol < entry.text.length;
		return {
			before: entry.text.slice(0, cursorDisplayCol),
			cursor: hasCharUnderCursor
				? entry.text.slice(cursorDisplayCol, cursorDisplayCol + 1)
				: '\u00a0',
			after: hasCharUnderCursor ? entry.text.slice(cursorDisplayCol + 1) : '',
			atEnd: !hasCharUnderCursor
		};
	}

	function scrollSelectedIntoView(node: HTMLElement, isSelected: boolean) {
		if (isSelected) {
			node.scrollIntoView({ block: 'center' });
		}
		return {
			update(isSelected: boolean) {
				if (isSelected) {
					node.scrollIntoView({ block: 'center' });
				}
			}
		};
	}
</script>

{#snippet pane(data: App.FileExplorerDirectory, role: 'parent' | 'current')}
	{@const isCurrentPane = role === 'current'}
	<div
		class="border-r-mantle flex h-full flex-col overflow-hidden border-r-2"
    class:bg-very-dark={!isCurrentPane}
    class:bg-pretty-dark={isCurrentPane}
		class:focused={isCurrentPane}
		class:parent-dir={role === 'parent'}
		class:current-dir={isCurrentPane}
	>
		<div class="pane-header">
			{#if isCurrentPane}
				<span class="title-top-marker" aria-hidden="true"></span>
			{/if}
			<div
				class:text-mauve={isCurrentPane}
				class:text-surface0={!isCurrentPane}
				class="pane-title mx-2 mb-2 flex items-center gap-4 overflow-hidden text-ellipsis whitespace-nowrap px-4 pb-4 text-xl"
			>
				{#if isCurrentPane}
					<span class="current-indicator"> </span>
				{/if}
				<span>
					{data.title}
				</span>
			</div>
		</div>
		<div class="min-h-0 flex-1 overflow-y-auto">
			{#each data.entries as entry (entry.id)}
				{@const selected = data.selectedEntryId === entry.id}
				{@const cursorActive = isCurrentPane && selected}
				{@const split = cursorActive ? splitEntryTextAtCursor(entry, data.cursorCol ?? 0) : null}
				<div
					use:scrollSelectedIntoView={selected}
					class="flex h-9 cursor-default items-center gap-1 whitespace-nowrap px-4 py-1"
					class:bg-surface0={selected}
				>
					<span
            class:text-peach={entry.isDir}
						class={`grid h-[2ch] w-[2ch] shrink-0 place-items-center rounded-full ${entry.iconClass ?? ''}`}
						>{entry.icon}</span
					>
					<span class="flex h-full overflow-hidden text-ellipsis">
						{#if cursorActive && split}
							<span class="grid h-full place-items-center">{split.before}</span>
							<span
								class="grid h-full place-items-center"
								class:cursor={true}
								class:end-cursor={split.atEnd}>{split.cursor}</span
							><span class="grid h-full place-items-center">{split.after}</span>
						{:else}
							<span class="grid h-full place-items-center">
								{entry.text}
							</span>
						{/if}
					</span>
				</div>
			{/each}
		</div>
	</div>
{/snippet}

{#if visible && state?.current}
	<div
		class="file-explorer victor-mono bg-very-dark active-window text-text"
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

	.pane-header {
		position: relative;
		flex-shrink: 0;
	}

	.title-top-marker {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 2px;
		border-radius: 9999px;
		background: linear-gradient(
			90deg,
			rgba(var(--ctp-mauve), 1) 0%,
			rgba(var(--ctp-blue), 0.35) 65%,
			rgba(var(--ctp-blue), 0) 100%
		);
		box-shadow:
			0 0 8px rgba(var(--ctp-mauve), 0.8),
			0 0 16px rgba(var(--ctp-mauve), 0.35);
	}

	.current-indicator {
		display: inline-block;
		text-shadow:
			0 0 2px rgba(var(--ctp-mauve), 0.95),
			0 0 5px rgba(var(--ctp-mauve), 0.6);
		filter: drop-shadow(0 0 6px rgba(var(--ctp-mauve), 0.8));
	}
</style>
