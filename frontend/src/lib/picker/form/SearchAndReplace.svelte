<script>
	import {
		search_term,
		replace_term,
		dir as _dir,
		include,
		exclude,
		preserve_case,
		case_sensitive,
		regex,
		match_whole_word
	} from '$lib/store.js';
	import PreserveCase from '$lib/icons/PreserveCase.svelte';
	import CaseSensitive from '$lib/icons/CaseSensitive.svelte';
	import Regex from '$lib/icons/Regex.svelte';
	import MatchWholeWord from '$lib/icons/MatchWholeWord.svelte';
	import { onMount } from 'svelte';
	import DebouncedInput from './DebouncedInput.svelte';

	let dir = '';
	onMount(() => {
		setTimeout(() => {
			dir = $_dir;
		});
	});
</script>

<div class="flex">
	<div class="w-1/2 pr-1">
		<label class="input input-bordered mr-1 flex w-full items-center bg-mantle">
			<DebouncedInput
				autofocus={true}
				with_label={true}
				store={search_term}
				placeholder={'Search...'}
				value={$search_term}
			/>
			{#if $case_sensitive}
				<CaseSensitive></CaseSensitive>
			{/if}
			{#if $regex}
				<Regex></Regex>
			{/if}
			{#if $match_whole_word}
				<MatchWholeWord></MatchWholeWord>
			{/if}
		</label>
	</div>
	<div class="w-1/2 pl-1">
		<label class="input input-bordered flex w-full items-center gap-2 bg-mantle">
			<input bind:value={$replace_term} type="text" placeholder="Replace..." class="grow" />
			{#if $preserve_case}
				<PreserveCase></PreserveCase>
			{/if}
		</label>
	</div>
</div>
<div class="flex flex-wrap py-2 sm:grid sm:grid-cols-[2fr,1fr,1fr] sm:gap-1">
	<DebouncedInput value={dir} placeholder={'eg. /home/user/Projects'} store={_dir} />
	<div class="w-1/2 pr-1 pt-2 sm:w-auto sm:pl-1 sm:pr-0 sm:pt-0">
		<DebouncedInput
			value={$exclude}
			placeholder={'eg *service.go,src/**/exclude'}
			store={include}
		/>
	</div>
	<div class="w-1/2 pl-1 pt-2 sm:w-auto sm:pt-0">
		<DebouncedInput
			value={$include}
			placeholder={'eg *service.go,src/**/include'}
			store={include}
		/>
	</div>
</div>
