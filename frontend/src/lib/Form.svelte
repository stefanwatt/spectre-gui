<script>
	import { debounce } from './utils.service.js';
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
	} from './store.js';
	import PreserveCase from './icons/PreserveCase.svelte';
	import CaseSensitive from './icons/CaseSensitive.svelte';
	import Regex from './icons/Regex.svelte';
	import MatchWholeWord from './icons/MatchWholeWord.svelte';
	import { onMount } from 'svelte';

	let dir = '';
	onMount(() => {
		setTimeout(() => {
			dir = $_dir;
		});
	});
	/**
	 * @param {KeyboardEvent & { target: HTMLInputElement }} e
	 */
	async function debounced_search_term(e) {
		$search_term = await debounce(e.target.value);
	}

	/**
	 * @param {KeyboardEvent & { target: HTMLInputElement }} e
	 */
	async function debounced_dir(e) {
		$_dir = await debounce(e.target.value);
	}

	/**
	 * @param {KeyboardEvent & { target: HTMLInputElement }} e
	 */
	async function debounced_exclude(e) {
		$exclude = await debounce(e.target.value);
	}

	/**
	 * @param {KeyboardEvent & { target: HTMLInputElement }} e
	 */
	async function debounced_include({ target: { value } }) {
		$include = await debounce(value);
	}
</script>

<div class="flex">
	<div class="w-1/2 pr-1">
		<label class="input input-bordered mr-1 flex w-full items-center bg-mantle">
			<input
				autofocus
				on:keyup={debounced_search_term}
				value={$search_term}
				type="text"
				placeholder="Search..."
				class="grow"
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
	<input
		class="input input-bordered w-full"
		on:keyup={debounced_dir}
		value={dir}
		type="text"
		placeholder="eg. /home/user/Projects"
	/>
	<div class="w-1/2 pr-1 pt-2 sm:w-auto sm:pl-1 sm:pr-0 sm:pt-0">
		<input
			on:keyup={debounced_exclude}
			value={$exclude}
			type="text"
			placeholder="eg *service.go,src/**/exclude"
			class="input input-bordered w-full bg-mantle"
		/>
	</div>
	<div class="w-1/2 pl-1 pt-2 sm:w-auto sm:pt-0">
		<input
			on:keyup={debounced_include}
			value={$include}
			type="text"
			placeholder="eg *service.go,src/**/include"
			class="input input-bordered w-full"
		/>
	</div>
</div>
