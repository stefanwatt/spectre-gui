<script>
	import { onMount } from 'svelte';
	import '$lib/assets/prism.css';
	import ChevronDown from './icons/ChevronDown.svelte';
	import ChevronUp from './icons/ChevronUp.svelte';
	import ResultsHeader from './ResultsHeader.svelte';
	import { highlight_all } from './results.service';
	import { results } from './store';
	import Match from './Match.svelte';
	import NoResults from './NoResults.svelte';

	let collapsed = false;
	onMount(async () => {
		await import('./assets/prism.js');
		setTimeout(() => {
			highlight_all();
		});
	});
</script>

<div class="flex w-full flex-col">
	{#each $results as item (item.path)}
		<div class="grid snap-start grid-cols-[1fr,15fr] pt-2">
			<div class="mb-2 flex w-full items-center justify-end text-blue">
				{#if !collapsed}
					<ChevronDown></ChevronDown>
				{:else}
					<ChevronUp></ChevronUp>
				{/if}
			</div>
			<div class="mb-2">
				<ResultsHeader path={item.path} match_count={item.matches.length}></ResultsHeader>
			</div>
			{#each item.matches as match (match.Id)}
				{#if match.Col < 10000}
					<div class="flex w-full items-center justify-end pr-2 text-overlay0">
						{match.Row}:{match.Col}
					</div>
					<Match {match}></Match>
				{/if}
			{/each}
		</div>
	{:else}
		<NoResults></NoResults>
	{/each}
</div>
