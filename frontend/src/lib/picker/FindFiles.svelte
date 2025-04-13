<script lang="ts">
	import { registerKeymap } from '$lib/keymaps/keymap-service';
	import { pickers } from '$lib/state.svelte';
	import { FindFiles, OpenFile } from '$lib/wailsjs/go/main/App';

	let results: { filename: string; relativePath: string; icon: string; iconColor: string }[] =
		$state([]);
	let findFilesState = $state({ searchTerm: '' });

	$effect(() => {
		FindFiles(findFilesState.searchTerm).then((updatedResults) => {
			console.log(updatedResults)
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
</script>

<div class="flex h-full w-full flex-col p-2 text-xl">
	<div class="w-full pb-4">
		<input
			placeholder="Find Files..."
			bind:value={findFilesState.searchTerm}
			autofocus
			type="text"
			class="input input-bordered w-full !text-xl"
		/>
	</div>
	<div class="grow overflow-y-scroll px-1">
		{#each results as result, i}
			<div class:bg-surface1={i === selectedMatchIndex} class="flex items-center py-1 text-center rounded-md p-1">
				<span class="text-3xl" style="color: {result.iconColor};">{result.icon}</span>
				<span class="ml-2 text-center">{result.filename}</span>
				<span 
					class:text-surface1={i!== selectedMatchIndex}
					class:text-crust={i=== selectedMatchIndex}
					class="ml-2 text-center">{result.relativePath}</span>
			</div>
		{/each}
	</div>
</div>
