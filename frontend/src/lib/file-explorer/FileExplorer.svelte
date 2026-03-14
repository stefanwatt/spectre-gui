<script lang="ts">
	import { getFileExplorerState, getFileExplorerVisible } from '$lib/state.svelte';

	let state = $derived(getFileExplorerState());
	let visible = $derived(getFileExplorerVisible());
</script>

{#snippet pane(data: App.FileExplorerPane, role: 'parent' | 'current' | 'preview')}
	<div
		class="border-r-2 border-r-surface0 overflow-y-auto"
		class:focused={role === 'current'}
		class:parent-dir={role === 'parent'}
		class:current-dir={role === 'current'}
		class:preview={role === 'preview'}
	>
		<div class="px-1 py-2 text-xl whitespace-nowrap overflow-hidden text-ellipsis border-b-2 border-b-surface0">{data.title}</div>
		<div class="overflow-y-auto">
			{#each data.entries as entry, i}
				<div
					class="flex cursor-default gap-1 whitespace-nowrap px-1 py-2"
					class:bg-blue-300={i + 1 === data.cursor_line}
					class:text-black={i + 1 === data.cursor_line}
					class:directory={entry.fs_type === 'directory'}
				>
					<span class="w-[2ch] text-center shrink-0">{entry.icon}</span>
					<span class="overflow-hidden text-ellipsis">{entry.name}</span>
				</div>
			{/each}
		</div>
	</div>
{/snippet}

{#if visible && state?.current_dir}
	<div class="file-explorer victor-mono">
		<div class="panes grid h-screen w-screen grid-rows-1 bg-base-100">
			{#if state.parent_dir}
				{@render pane(state.parent_dir, 'parent')}
			{/if}
			{@render pane(state.current_dir, 'current')}
			{#if state.preview}
				{@render pane(state.preview, 'preview')}
			{/if}
		</div>
	</div>
{/if}

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
