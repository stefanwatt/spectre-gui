<script lang="ts">
	import SearchIcon from '$lib/icons/Search.svelte';
	import TextWithCursor from './TextWithCursor.svelte';
	import ArrowUpIcon from '$lib/icons/ArrowUp.svelte';
	import ArrowDownIcon from '$lib/icons/ArrowDown.svelte';
	let totalResults = 17;
	let resultIndex = 1;
	let {
		searchIconSize,
		content,
		pos
	}: { searchIconSize: number | string; content?: string; pos?: number } = $props();
</script>

<div class="relative w-screen text-sm">
	<div class="cmdline-container bg-dark right-8 top-8 z-50">
		<div class=" search-mode flex items-center p-2">
			<div class="search-icon-container mr-2 text-overlay2">
				<SearchIcon size={searchIconSize} />
			</div>
			<div class="cmdline-content flex flex-grow">
				{#if content !== undefined && pos !== undefined}
					<TextWithCursor text={content} {pos} />
				{/if}
			</div>
			<div class="search-controls ml-2 flex items-center text-sm text-overlay0">
				<span class="mr-4 text-xs">{resultIndex}/{totalResults}</span>
				<div class="flex space-x-2">
					<button class="rounded p-1 hover:bg-overlay2">
						<ArrowUpIcon size={searchIconSize} />
					</button>
					<button class="rounded p-1 hover:bg-overlay2">
						<ArrowDownIcon size={searchIconSize} />
					</button>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.search-mode {
		width: 25rem;
	}

	.search-controls button {
		transition: background-color 0.07s;
	}
</style>
