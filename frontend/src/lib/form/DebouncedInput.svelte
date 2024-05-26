<script>
	import { debounce } from '$lib/utils.service.js';

	export let autofocus = false;
	export let with_label = false;
	/** @type {string}*/
	export let placeholder = '';
	/** @type {string}*/
	export let value;
	/** @type {import("svelte/store").Writable<string>}*/
	export let store;

	/**
	 * @param {KeyboardEvent & { target: HTMLInputElement }} e
	 */
	async function debounced_update(e) {
		$store = await debounce(e.target.value);
	}
</script>

<input
	class:input={!with_label}
	class:input-bordered={!with_label}
	class:w-full={!with_label}
	class="grow"
	on:keyup={debounced_update}
	{value}
	type="text"
	{placeholder}
	{autofocus}
/>
