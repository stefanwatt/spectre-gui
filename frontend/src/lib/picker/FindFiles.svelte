<script lang="ts">
	import { registerKeymap } from '$lib/keymaps/keymap-service';
	import { pickers } from '$lib/state.svelte';
	import { isInBounds } from '$lib/utils.service';
	import { FindFiles, OpenFile } from '$lib/wailsjs/go/main/App';
	import NoResults from './results/NoResults.svelte';

	let results: { filename: string; relativePath: string; icon: string; iconColor: string }[] =
		$state([]);
	let findFilesState = $state({ searchTerm: '' });

	$effect(() => {
		FindFiles(findFilesState.searchTerm).then((updatedResults) => {
			console.log(updatedResults);
			results = updatedResults;
		});
	});

	let selectedMatchIndex = $state(0);

	registerKeymap({
		key: 'ArrowDown',
		mode: 'find-files',
		mods: [],
		action: () => {
			if (selectedMatchIndex + 1 > results.length - 1) {
				selectedMatchIndex = 0;
				return;
			}
			selectedMatchIndex++;
		}
	});

	registerKeymap({
		key: 'ArrowUp',
		mode: 'find-files',
		mods: [],
		action: () => {
			if (selectedMatchIndex - 1 < 0) {
				selectedMatchIndex = results.length - 1;
				return;
			}
			selectedMatchIndex--;
		}
	});

	registerKeymap({
		key: 'Enter',
		mode: 'find-files',
		mods: [],
		action: () => {
			OpenFile(selectedMatch.absolutePath, 1, 1);
			pickers.findFiles = false;
		}
	});

	let selectedMatch = $derived(results[selectedMatchIndex]);
	$effect(() => {
		if (!selectedMatch) return;
		const selectedMatchElem = document.querySelector('.find-files-selected-match');
		if (!selectedMatchElem) return;
		if (isInBounds(selectedMatchElem as HTMLElement, scrollContainer)) return;
		selectedMatchElem.scrollIntoView({ block: 'nearest', inline: 'nearest' });
	});
	let scrollContainer: HTMLElement;
</script>

<div class="flex h-full w-full flex-col p-4 text-xl">
	<div class="w-full pb-4">
		<div class="mb-2 text-center text-3xl">Find Files</div>
		<input
			bind:value={findFilesState.searchTerm}
			autofocus
			type="text"
			class="input input-bordered w-full !text-xl"
		/>
	</div>
	<div bind:this={scrollContainer} class="snap-manatory grow snap-y overflow-y-scroll">
		{#each results as result, i}
			<div
				class:find-files-selected-match={i === selectedMatchIndex}
				class:bg-base={i === selectedMatchIndex}
				class="flex snap-end snap-always items-center rounded-md px-1 py-2 text-center"
			>
				<span
					class="flex h-10 w-10 items-center justify-center rounded-full bg-text px-1 text-3xl"
					style="color: {result.iconColor};">{result.icon}</span
				>
				<span class="ml-4 text-center">{result.filename}</span>
				<span
					class:text-surface1={i !== selectedMatchIndex}
					class:text-darker={i === selectedMatchIndex}
					class="ml-2 text-center">{result.relativePath}</span
				>
			</div>
		{:else}
			<div class="h-full flex flex-col justify-center">
				<NoResults></NoResults>
			</div>
		{/each}
	</div>
</div>
