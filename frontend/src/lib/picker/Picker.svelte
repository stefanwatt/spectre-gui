<script lang="ts">
	import { registerKeymap } from '$lib/keymaps/keymap-service';
	import { isInBounds } from '$lib/utils.service';
	import { OpenFile, GetPreview } from '$lib/wailsjs/go/picker/Picker';
	import NoResults from './live-grep-results/NoResults.svelte';
	import { nestedState, windowContentRowMap } from '$lib/state.svelte';
	import FloatingGrid from '$lib/floating-windows/FloatingGrid.svelte';

	interface PickerProps {
		onQueryChanged: (query: string) => void;
		mode: App.KeymapMode;
		title: string;
	}
	let itemHeight = $state(48);
	let { onQueryChanged, mode, title }: PickerProps = $props();
	let pickerState = $state({ query: '' });
	let visibleStartIndex = $state(0);
	let visibleEndIndex = $state(20);
	let itemsToRender = 50;
	$effect(() => {
		const halfBuffer = Math.floor(itemsToRender / 2);
		visibleStartIndex = Math.max(0, selectedMatchIndex - halfBuffer);
		visibleEndIndex = Math.min(
			nestedState.pickerResults.length - 1,
			visibleStartIndex + itemsToRender - 1
		);
	});
	$effect(() => {
		onQueryChanged(pickerState.query);
	});

	let selectedMatchIndex = $state(0);
	function updateItemHeight() {
		const sampleItem = document.querySelector('.flex.snap-end.snap-always');
		if (sampleItem) {
			itemHeight = sampleItem.getBoundingClientRect().height;
		}
	}

	$effect(() => {
		if (nestedState.pickerResults.length > 0) {
			updateItemHeight();
		}
	});
	registerKeymap({
		key: 'ArrowDown',
		mode,
		mods: [],
		action: () => {
			if (selectedMatchIndex + 1 > nestedState.pickerResults.length - 1) {
				selectedMatchIndex = 0;
				return;
			}
			selectedMatchIndex++;
		}
	});

	registerKeymap({
		key: 'ArrowUp',
		mode,
		mods: [],
		action: () => {
			if (selectedMatchIndex - 1 < 0) {
				selectedMatchIndex = nestedState.pickerResults.length - 1;
				return;
			}
			selectedMatchIndex--;
		}
	});

	registerKeymap({
		key: 'Enter',
		mode,
		mods: [],
		action: () => {
			OpenFile(selectedMatch.absolutePath, selectedMatch.row || 1, selectedMatch.col || 1);
			nestedState.activePicker = undefined;
		}
	});

	let selectedMatch = $derived(
		nestedState.pickerResults[selectedMatchIndex] || nestedState.pickerResults[0]
	);
	$effect(() => {
		if (!selectedMatch) {
			selectedMatchIndex = 0;
			return;
		}

		const row = selectedMatch.row || 1;
		const col = selectedMatch.col || 1;
		GetPreview(selectedMatch.absolutePath, row, col);
		if (selectedMatchIndex < visibleStartIndex || selectedMatchIndex > visibleEndIndex) {
			const halfBuffer = Math.floor(itemsToRender / 2);
			visibleStartIndex = Math.max(0, selectedMatchIndex - halfBuffer);
			visibleEndIndex = Math.min(
				nestedState.pickerResults.length - 1,
				visibleStartIndex + itemsToRender - 1
			);
		}
		const selectedMatchElem = document.querySelector('.picker-selected-match');
		if (!selectedMatchElem) return;
		if (isInBounds(selectedMatchElem as HTMLElement, scrollContainer)) return;
		selectedMatchElem.scrollIntoView({ block: 'nearest', inline: 'nearest' });
	});
	let scrollContainer: HTMLElement;
	let previewContent = $derived(
		nestedState.previewWindow && windowContentRowMap[nestedState.previewWindow.id]
	);
</script>

<div class="flex h-full flex-col">
	<div class="w-full p-4">
		<div class="mb-2 text-center text-3xl">{title}</div>
		<input
			bind:value={pickerState.query}
			autofocus
			type="text"
			class="input input-bordered w-full !text-xl"
		/>
	</div>

	<div class="flex w-full grow overflow-hidden">
		{#if nestedState.pickerResults.length === 0}
			<div class="flex h-full w-full flex-col justify-center">
				<NoResults></NoResults>
			</div>
		{:else}
			<!-- else content here -->
			<div class="flex h-full w-1/2 flex-col p-4 text-xl">
				<div bind:this={scrollContainer} class="snap-manatory grow snap-y overflow-y-scroll">
					{#if visibleStartIndex > 0}
						<div style="height: {visibleStartIndex * itemHeight}px"></div>
					{/if}

					<!-- Only render visible items -->
					{#each nestedState.pickerResults.slice(visibleStartIndex, visibleEndIndex + 1) as result, i}
						{@const actualIndex = i + visibleStartIndex}
						<div
							class:picker-selected-match={actualIndex === selectedMatchIndex}
							class:bg-base={actualIndex === selectedMatchIndex}
							class="flex snap-end snap-always items-center rounded-md px-1 py-2 text-center"
						>
							<span
								class="flex h-10 w-10 items-center justify-center rounded-full bg-text px-1 text-3xl"
								style="color: {result.iconColor};">{result.icon}</span
							>
							{#if result.filename}
								<span class="ml-4 text-nowrap text-center">{result.filename}</span>
								<span>:</span>
							{/if}
							{#if result.row && result.col}
								<span class="ml-4 flex w-20 justify-start">
									<span class="text-center text-green">{result.row}</span>
									<span>:</span>
									<span class="text-center text-blue">{result.col}</span>
								</span>
							{/if}
							{#if result.text}
								<span class="ml-12 overflow-hidden overflow-ellipsis whitespace-nowrap text-center"
									>{result.text}</span
								>
							{:else}
								<span
									class:text-surface1={actualIndex !== selectedMatchIndex}
									class:text-darker={actualIndex === selectedMatchIndex}
									class="ovewflow-hidden cut-off-start ml-2">{result.relativePath}</span
								>
							{/if}
						</div>
					{/each}
					{#if visibleEndIndex < nestedState.pickerResults.length - 1}
						<div
							style="height: {(nestedState.pickerResults.length - 1 - visibleEndIndex) *
								itemHeight}px"
						></div>
					{/if}
				</div>
			</div>
			{#if nestedState.previewWindow}
				<div class="h-full w-1/2 overflow-hidden bg-mantle p-4">
					<FloatingGrid content={previewContent} />
				</div>
			{/if}
		{/if}
	</div>
</div>

<style>
	.cut-off-start {
		direction: rtl;
		text-align: left;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
</style>
