<script lang="ts">
	import FloatingGrid from '$lib/floating-windows/FloatingGrid.svelte';
	import { fileExplorer, windowContentRowMap } from '$lib/state.svelte';

	let centerContent = $derived(windowContentRowMap.get(fileExplorer.centerWindowId));
	let previewContent = $derived(windowContentRowMap.get(fileExplorer.previewWindowId));

	let currentDirName = $derived(fileExplorer.currentDir.split('/').filter(Boolean).pop() ?? '/');

	let breadcrumbs = $derived(fileExplorer.currentDir.split('/').filter(Boolean));
</script>

<div class="file-explorer">
	<div class="header">
		<span class="breadcrumbs">
			{#each breadcrumbs as crumb, i}
				{#if i > 0}<span class="separator">/</span>{/if}
				<span class="crumb" class:current={i === breadcrumbs.length - 1}>{crumb}</span>
			{/each}
		</span>
	</div>

	<div class="columns">
		<!-- Left: parent directory listing -->
		<div class="column parent-column">
			{#each fileExplorer.parentEntries as entry}
				<div
					class="entry"
					class:dir={entry.isDir}
					class:active={entry.name === currentDirName}
				>
					{entry.isDir ? entry.name + '/' : entry.name}
				</div>
			{/each}
		</div>

		<!-- Center: editable neovim buffer -->
		<div class="column center-column">
			<FloatingGrid content={centerContent} lineNumbers={false} relativeLineNumbers={false} />
		</div>

		<!-- Right: preview -->
		<div class="column preview-column">
			{#if fileExplorer.previewIsFile}
				<FloatingGrid content={previewContent} lineNumbers={false} relativeLineNumbers={false} />
			{:else}
				{#each fileExplorer.previewEntries as entry}
					<div class="entry" class:dir={entry.isDir}>
						{entry.isDir ? entry.name + '/' : entry.name}
					</div>
				{/each}
			{/if}
		</div>
	</div>
</div>

<style>
	.file-explorer {
		position: absolute;
		inset: 0;
		bottom: 40px; /* above status line */
		z-index: 100;
		display: flex;
		flex-direction: column;
		background: var(--base-100, #1e1e2e);
		color: var(--text, #cdd6f4);
	}

	.header {
		padding: 6px 12px;
		border-bottom: 1px solid var(--surface0, #313244);
		font-size: 16px;
		opacity: 0.8;
	}

	.breadcrumbs {
		display: flex;
		gap: 2px;
		align-items: center;
	}

	.separator {
		opacity: 0.4;
		margin: 0 2px;
	}

	.crumb.current {
		opacity: 1;
		font-weight: 700;
	}

	.crumb:not(.current) {
		opacity: 0.5;
	}

	.columns {
		display: grid;
		grid-template-columns: 1fr 2fr 1fr;
		flex: 1;
		overflow: hidden;
	}

	.column {
		overflow-y: auto;
		overflow-x: hidden;
		padding: 4px 0;
		scrollbar-width: none;
	}

	.column::-webkit-scrollbar {
		display: none;
	}

	.parent-column {
		border-right: 1px solid var(--surface0, #313244);
		opacity: 0.7;
	}

	.preview-column {
		border-left: 1px solid var(--surface0, #313244);
		opacity: 0.8;
	}

	.entry {
		padding: 3px 12px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		line-height: 22px;
	}

	.entry.dir {
		color: var(--blue, #89b4fa);
	}

	.entry.active {
		background: var(--surface0, #313244);
		opacity: 1;
		font-weight: 700;
	}
</style>
