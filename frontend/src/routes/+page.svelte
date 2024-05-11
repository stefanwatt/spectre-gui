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
		search_flags,
		preserve_case,
		results
	} from '$lib/store';
	import { listen_for_events } from '$lib/runtime-events.service';
	import {GetFormValues} from "$lib/wailsjs/go/main/App"

	/**@type {App.RipgrepMatch} */
	$: {
		const flags = $search_flags.map((f) => f.text);
		search($search_term, $dir, $include, $exclude, flags, $replace_term, $preserve_case);
	}

	onMount(async () => {
		setup_keymaps();
		listen_for_events();
		/**@type {App.FormValues}*/
		const form_values = await GetFormValues()
		$search_term = form_values.SearchTerm
		$replace_term = form_values.ReplaceTerm
		$dir = form_values.Dir
		$include = form_values.Include
		$exclude = form_values.Exclude
	});
	$: total_matches = $results?.flatMap((result) => result.Matches)?.length || 0;
</script>

<div
	class="min-w-screen flex h-full min-h-screen w-full flex-col overflow-hidden bg-base px-2 py-4"
>
	<Form></Form>
	{#if total_matches !== 0}
		<div class="mt-2 pl-1 text-overlay2">
			<span class="font-bold text-blue">{total_matches}</span>
			<span>Results in </span>
			<span class="font-bold text-blue">{$results.length}</span>
			<span>files</span>
		</div>
	{/if}
	<div class="h-0 grow snap-y snap-mandatory overflow-x-hidden overflow-y-scroll pt-2">
		<Results></Results>
	</div>
</div>
{#if $toast}
	<div transition:fade={{ delay: 250, duration: 300 }}>
		<Toast text={$toast.text} level={$toast.level}></Toast>
	</div>
{/if}
