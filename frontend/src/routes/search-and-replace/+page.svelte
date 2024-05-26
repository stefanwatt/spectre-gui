<script>
	import Results from '$lib/results/Results.svelte';
	import { setup_keymaps } from '$lib/keymaps.service.js';
	import { search } from '$lib/results/results.service';
	import { onMount } from 'svelte';
	import {
		search_term,
		replace_term,
		dir,
		include,
		exclude,
		case_sensitive,
		regex,
		match_whole_word,
		preserve_case,
		total_results,
		total_files,
		total_pages,
		page_index
	} from '$lib/store';
	import { listen_for_events } from '$lib/runtime-events.service';
	import { GetAppState } from '$lib/wailsjs/go/main/App';
	import SearchAndReplaceForm from '$lib/form/SearchAndReplace.svelte';
	import { keymaps } from './keymaps';

	export let replace = true;

	/**@type {App.RipgrepMatch} */
	$: {
		search(
			$search_term,
			$dir,
			$exclude,
			$include,
			$case_sensitive,
			$regex,
			$match_whole_word,
			$preserve_case
		);
	}

	onMount(async () => {
		setup_keymaps(keymaps);
		listen_for_events();
		/**@type {App.State}*/
		const app_state = await GetAppState();
		$search_term = app_state.SearchTerm;
		$replace_term = app_state.ReplaceTerm;
		$dir = app_state.Dir;
		$include = app_state.Include;
		$exclude = app_state.Exclude;
		// @ts-ignore
		$case_sensitive = app_state.CaseSensitive;
		// @ts-ignore
		$regex = app_state.Regex;
		// @ts-ignore
		$match_whole_word = app_state.MatchWholeWord;
		// @ts-ignore
		$preserve_case = app_state.PreserveCase;
	});
</script>

<SearchAndReplaceForm />
{#if $total_results !== 0}
	<div class="mt-2 pl-1 text-overlay2">
		<span class="font-bold text-blue">{$total_results}</span>
		<span>Results in </span>
		<span class="font-bold text-blue">{$total_files}</span>
		<span>files</span>
		<span class="ml-4">Page </span>
		<span class="font-bold text-blue">{$page_index + 1}</span>
		<span>of </span>
		<span class="font-bold text-blue">{$total_pages}</span>
	</div>
{/if}
<div class="flex h-0 min-h-full grow overflow-y-hidden pt-2">
	<Results></Results>
</div>

<style>
	:global(.spectre-matched),
	:global(.spectre-replacement) {
		padding: 0.2rem;
		text-align: center;
		line-height: 100%;
		margin: 0px 0.2rem;
		font-family: monospace;
		border-radius: 0.15rem;
	}

	:global(.spectre-matched) {
		background-color: #eebebe;
		color: #51576d;
		text-decoration: line-through;
	}

	:global(.spectre-replacement) {
		color: #51576d;
		background-color: #a6d189;
	}
</style>
