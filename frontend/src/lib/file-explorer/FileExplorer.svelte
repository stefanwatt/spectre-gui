<script lang="ts">
	import { getFileExplorerState, getFileExplorerVisible } from '$lib/state.svelte';
	import Grid from '$lib/windows/Grid.svelte';

	let state = $derived(getFileExplorerState());
	let visible = $derived(getFileExplorerVisible());
</script>

{#snippet pane(data: App.FileExplorerDirectory, role: 'parent' | 'current' | 'preview')}
	<div
		class="overflow-y-auto border-r-2 border-r-surface0"
		class:focused={role === 'current'}
		class:parent-dir={role === 'parent'}
		class:current-dir={role === 'current'}
		class:preview={role === 'preview'}
	>
		<!-- <div -->
		<!-- 	class="overflow-hidden text-ellipsis whitespace-nowrap border-b-2 border-b-surface0 px-1 py-2 text-xl" -->
		<!-- > -->
		<!-- 	{data.title} -->
		<!-- </div> -->
		<div class="overflow-y-auto">
			{#each data.entries as entry (entry.id)}
				<div
					class="flex cursor-default gap-1 whitespace-nowrap px-1 py-2"
					class:bg-blue-300={data.selectedEntryId === entry.id}
					class:text-black={data.selectedEntryId === entry.id}
				>
					<span class="w-[2ch] shrink-0 text-center">{entry.icon}</span>
					<span class="overflow-hidden text-ellipsis">{entry.text}</span>
				</div>
			{/each}
		</div>
	</div>
{/snippet}

{#if visible && state?.current}
	<div class="file-explorer victor-mono">
		<div class="panes grid h-screen w-screen grid-rows-1 ">
			{#if state.parent}
				{@render pane(state.parent, 'parent')}
			{/if}
			{@render pane(state.current, 'current')}
			{#if state.preview}
				{#if state.preview.directory}
					{@render pane(state.preview.directory, 'preview')}
				{:else}
					<div>
						<Grid content={state.preview.content} lineNumbers={false} relativeLineNumbers={false}
						></Grid>
					</div>
				{/if}
			{/if}
		</div>
	</div>
{/if}

<!-- class:bg-blue-300={i + 1 === data.cursor_line} -->
<!-- class:text-black={i + 1 === data.cursor_line} -->
<!-- class:directory={entry.fs_type === 'directory'} -->

<style>
	.panes {
		grid-template-columns: 3fr 3fr 6fr;
	}
	.file-explorer {
		position: absolute;
		top: 0;
		left: 0;
		z-index: 100;
		pointer-events: none;
	}
	.pane:last-child {
		border-right: none;
	}
	.pane.focused .pane-title {
		opacity: 1;
	}
</style>
