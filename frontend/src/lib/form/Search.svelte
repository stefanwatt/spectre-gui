<script lang="ts">
	import CaseSensitive from '$lib/icons/CaseSensitive.svelte';
	import Regex from '$lib/icons/Regex.svelte';
	import MatchWholeWord from '$lib/icons/MatchWholeWord.svelte';
	import DebouncedInput from '$lib/form/DebouncedInput.svelte';

	let { state }: { state: App.LiveGrepOpts } = $props();
</script>

<div class="flex">
	<div class="w-1/2 pr-1">
		<label class="input input-bordered mr-1 flex w-full items-center bg-mantle">
			<DebouncedInput
				autofocus={true}
				withLabel={true}
				value={state.searchTerm}
				updateValue={(updatedValue)=>{state.searchTerm=updatedValue}}
				placeholder={'Search...'}
			/>
			{#if state.caseSensitive}
				<CaseSensitive></CaseSensitive>
			{/if}
			{#if state.regex}
				<Regex></Regex>
			{/if}
			{#if state.matchWholeWord}
				<MatchWholeWord></MatchWholeWord>
			{/if}
		</label>
	</div>
	<div class="w-1/2 pl-1">
		<DebouncedInput
			value={state.dir}
			placeholder={'eg. /home/user/Projects'}
			updateValue={(updatedValue)=>{state.dir=updatedValue}}
		/>
	</div>
</div>
<div class="flex flex-wrap py-2 sm:grid sm:grid-cols-[1fr,1fr] sm:gap-1">
	<div class="w-1/2 pr-1 pt-2 sm:w-auto sm:pr-0 sm:pt-0">
		<DebouncedInput
			value={state.exclude}
			placeholder={'eg *service.go,src/**/exclude'}
			updateValue={(updatedValue)=>{state.exclude=updatedValue}}
		/>
	</div>
	<div class="w-1/2 pl-1 pt-2 sm:w-auto sm:pt-0">
		<DebouncedInput
			value={state.include}
			placeholder={'eg *service.go,src/**/include'}
			updateValue={(updatedValue)=>{state.include=updatedValue}}
		/>
	</div>
</div>
