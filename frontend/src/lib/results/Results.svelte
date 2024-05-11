<script>
	import ChevronDown from '$lib/icons/ChevronDown.svelte';
	import ChevronUp from '$lib/icons/ChevronUp.svelte';
	import ResultsHeader from './ResultsHeader.svelte';
	import { results } from '$lib/store';
	import Match from '$lib/Match.svelte';
	import NoResults from './NoResults.svelte';

	let collapsed = false;
</script>

{#if $results?.length}
	<div class="grid w-full grid-cols-[1fr,15fr] overflow-x-hidden md:grid-cols-[5rem,auto]">
		{#each $results as item (item.Path)}
			<div class="mb-2 flex w-full items-center justify-end text-blue">
				{#if !collapsed}
					<ChevronDown></ChevronDown>
				{:else}
					<ChevronUp></ChevronUp>
				{/if}
			</div>
			<div class="mb-2">
				<ResultsHeader path={item.Path} match_count={item.Matches.length}></ResultsHeader>
			</div>
			{#each item.Matches as match (match.Id)}
				{#if match.Col < 10000}
					<div class="flex w-full items-center justify-end pr-2 text-overlay0">
						{match.Row}:{match.Col}
					</div>
					<Match {match}></Match>
				{/if}
			{/each}
		{/each}
	</div>
{:else}
	<NoResults></NoResults>
{/if}
