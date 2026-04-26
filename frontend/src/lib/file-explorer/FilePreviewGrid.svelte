<script lang="ts">
	import Grid from '$lib/windows/Grid.svelte';

	interface Props {
		title: string;
		content: App.NvimRow[];
		contentEl?: HTMLDivElement;
	}

	let { title, content, contentEl = $bindable() }: Props = $props();
</script>

<div class="bg-very-dark flex h-full flex-col overflow-hidden">
	<div class="border-mantle text-surface0 flex shrink-0 items-center gap-3 border-b-2 px-6 py-2 text-xl">
		<span class="text-green">󰈙</span>
		<span class="overflow-hidden text-ellipsis whitespace-nowrap">{title}</span>
	</div>
	<div class="min-h-0 flex-1 overflow-hidden" bind:this={contentEl}>
		{#if content?.length}
			<Grid {content} lineNumbers={false} relativeLineNumbers={false}></Grid>
		{:else}
			<div class="text-overlay0 p-4">Loading preview…</div>
		{/if}
	</div>
</div>
