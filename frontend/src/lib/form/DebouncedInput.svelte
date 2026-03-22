<script lang="ts">
	import { debounce } from '$lib/utils.service';
	import type { KeyboardEventHandler } from 'svelte/elements';

	interface DebouncedInputProps {
		value: string;
		placeholder: string;
		withLabel?: boolean;
		autofocus?: boolean;
		debounceMs?: number;
		updateValue: (updatedValue: string) => void;
	}
	let { value, placeholder, withLabel, autofocus, updateValue, debounceMs }: DebouncedInputProps =
		$props();

	const debouncedUpdate: KeyboardEventHandler<HTMLInputElement> = function (e) {
		if (!e?.target) return;
		const target: HTMLInputElement = e.target as HTMLInputElement;
		debounce(target.value, debounceMs).then((updatedValue) => {
			updateValue(updatedValue);
		});
	};
</script>

<input
	class:input={!withLabel}
	class:input-bordered={!withLabel}
	class:w-full={!withLabel}
	class="grow"
	onkeyup={debouncedUpdate}
	{value}
	type="text"
	{placeholder}
	{autofocus}
/>
