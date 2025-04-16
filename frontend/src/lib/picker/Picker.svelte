<script lang="ts">
	import { registerKeymap } from '$lib/keymaps/keymap-service';
	import { pickers } from '$lib/state.svelte';
	import { isInBounds } from '$lib/utils.service';
	import { OpenFile } from '$lib/wailsjs/go/main/App';
	import NoResults from './results/NoResults.svelte';
	import { nestedState } from '$lib/state.svelte';

	interface PickerProps {
		onQueryChanged: (query: string) => void;
		mode: App.KeymapMode;
		pickerPropName: keyof App.Pickers;
		title: string;
	}
	let { onQueryChanged, mode, pickerPropName, title }: PickerProps = $props();
	let pickerState = $state({ query: '' });

	$effect(() => {
		onQueryChanged(pickerState.query);
	});

	let selectedMatchIndex = $state(0);

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
			pickers[pickerPropName] = false;
		}
	});

	registerKeymap({
		key: 'Escape',
		mode,
		mods: [],
		action: () => {
			pickers[pickerPropName] = false;
		}
	});

	let selectedMatch = $derived(nestedState.pickerResults[selectedMatchIndex]);
	$effect(() => {
		if (!selectedMatch) return;
		const selectedMatchElem = document.querySelector('.picker-selected-match');
		if (!selectedMatchElem) return;
		if (isInBounds(selectedMatchElem as HTMLElement, scrollContainer)) return;
		selectedMatchElem.scrollIntoView({ block: 'nearest', inline: 'nearest' });
	});
	let scrollContainer: HTMLElement;
</script>

<div class="flex h-full w-full flex-col p-4 text-xl">
	<div class="w-full pb-4">
		<div class="mb-2 text-center text-3xl">{title}</div>
		<input
			bind:value={pickerState.query}
			autofocus
			type="text"
			class="input input-bordered w-full !text-xl"
		/>
	</div>
	<div bind:this={scrollContainer} class="snap-manatory grow snap-y overflow-y-scroll">
		{#each nestedState.pickerResults as result, i}
			<div
				class:picker-selected-match={i === selectedMatchIndex}
				class:bg-base={i === selectedMatchIndex}
				class="flex snap-end snap-always items-center rounded-md px-1 py-2 text-center"
			>
				<span
					class="flex h-10 w-10 items-center justify-center rounded-full bg-text px-1 text-3xl"
					style="color: {result.iconColor};">{result.icon}</span
				>
				{#if result.filename}
					<span class="ml-4 text-center">{result.filename}</span>
					<span>:</span>
				{/if}
				{#if result.row && result.col}
					<span class="ml-4 w-20 flex justify-start">
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
						class:text-surface1={i !== selectedMatchIndex}
						class:text-darker={i === selectedMatchIndex}
						class="ml-2 text-center">{result.relativePath}</span
					>
				{/if}
			</div>
		{:else}
			<div class="h-full flex flex-col justify-center">
				<NoResults></NoResults>
			</div>
		{/each}
	</div>
</div>
