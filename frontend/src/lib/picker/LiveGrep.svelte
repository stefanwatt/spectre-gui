<script lang="ts">
	import Results from '$lib/picker/results/Results.svelte';
	import {
		state as resultsState,
		search
	} from '$lib/picker/results/results.service.svelte';
	import { onMount } from 'svelte';
	import { GetLiveGrepOpts } from '$lib/wailsjs/go/main/App';
	import SearchForm from '$lib/form/Search.svelte';

	let state = $state<App.LiveGrepOpts>();


	$effect(() => {
		if (!state) return;
		search(
			state.searchTerm,
			state.dir,
			state.exclude,
			state.include,
			state.caseSensitive,
			state.regex,
			state.matchWholeWord
		);
	});
	onMount(async () => {
		console.log('getting LiveGrepOpts');
		state = await GetLiveGrepOpts();
		console.log('got LiveGrepOpts', state);
	});
</script>

{#if state}
	<div class="flex h-full w-full flex-col">
		<SearchForm {state} />
		{#if resultsState.totalResults !== 0}
			<div class="mt-2 pl-1 text-overlay2">
				<span class="font-bold text-blue">{resultsState.totalResults}</span>
				<span>Results in </span>
				<span class="font-bold text-blue">{resultsState.totalFiles}</span>
				<span>files</span>
				<span class="ml-4">Page </span>
				<span class="font-bold text-blue">{resultsState.pageIndex + 1}</span>
				<span>of </span>
				<span class="font-bold text-blue">{resultsState.totalPages}</span>
			</div>
		{/if}
		<div class="flex h-0 min-h-full grow overflow-y-hidden pt-2">
			<Results searchTerm={state.searchTerm} regex={state.regex}></Results>
		</div>
	</div>
{/if}

<style>
	:global(.spectre-matched) {
		padding: 0.2rem;
		text-align: center;
		line-height: 100%;
		margin: 0px 0.2rem;
		font-family: monospace;
		border-radius: 0.15rem;
		background-color: #a6d189;
		color: #51576d;
		text-decoration: none;
	}
</style>
