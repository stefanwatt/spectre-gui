<script>
	import Form from '$lib/Form.svelte';
	import Results from '$lib/results/Results.svelte';
	import { setup_keymaps } from '$lib/keymaps.service.js';
	import Toast from '$lib/notification/Toast.svelte';
	import { search } from '$lib/results/results.service';
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import {
		toast,
		search_term,
		replace_term,
		dir,
		include,
		exclude,
		case_sensitive,
		regex,
		match_whole_word,
		preserve_case,
		results,
		total_results,
		total_files,
		total_pages,
		page_index
	} from '$lib/store';
	import { listen_for_events } from '$lib/runtime-events.service';
	import { GetAppState, GetReplacementText } from '$lib/wailsjs/go/main/App';

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
		setup_keymaps();
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

<div
	class="min-w-screen flex h-full min-h-screen w-full flex-col overflow-hidden bg-base px-2 py-4"
>
	<Form></Form>
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
</div>
{#if $toast}
	<div transition:fade={{ delay: 250, duration: 300 }}>
		<Toast text={$toast.text} level={$toast.level}></Toast>
	</div>
{/if}
