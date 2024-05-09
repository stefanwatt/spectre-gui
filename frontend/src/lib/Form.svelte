<script>
	import { debounce } from './utils.service.js';
	import {
		search_term,
		replace_term,
		dir,
		include,
		exclude,
		search_flags,
		preserve_case
	} from './store.js';
	import PreserveCase from './icons/PreserveCase.svelte';

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
		$dir = await debounce(e.target.value);
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
	<div class="w-full">
		<label class="input input-bordered flex items-center bg-mantle">
			<input
				autofocus
				on:keyup={debounced_search_term}
				value={$search_term}
				type="text"
				placeholder="Search..."
				class="grow"
			/>
			{#each $search_flags as flag}
				<svelte:component this={flag.icon} />
			{/each}
		</label>
	</div>
	<div class="w-full pl-2">
		<label class="input input-bordered flex items-center gap-2 bg-mantle">
			<input bind:value={$replace_term} type="text" placeholder="Replace..." class="grow" />
			{#if $preserve_case}
				<PreserveCase></PreserveCase>
			{/if}
		</label>
	</div>
</div>
<div class="flex py-2">
	<div class="w-full">
		<label class="input input-bordered flex items-center bg-mantle">
			<input
				on:keyup={debounced_dir}
				value={$dir}
				type="text"
				placeholder="Search..."
				class="grow"
			/>
		</label>
	</div>
	<div class="flex w-full">
		<label class="input input-bordered ml-2 flex w-full items-center bg-mantle">
			<input
				on:keyup={debounced_exclude}
				value={$exclude}
				type="text"
				placeholder="eg *service.go,src/**/exclude"
				class="grow"
			/>
		</label>
		<label class="input input-bordered ml-2 flex w-full items-center bg-mantle">
			<input
				on:keyup={debounced_include}
				value={$include}
				type="text"
				placeholder="eg *service.go,src/**/include"
				class="grow"
			/>
		</label>
	</div>
</div>
