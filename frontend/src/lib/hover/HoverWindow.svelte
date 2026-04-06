<script lang="ts">
	import { marked } from 'marked';
	import { getHoverWindow } from '$lib/state.svelte';
	import { windowContentRowMap, layout } from '$lib/state.svelte';

	let hover = $derived(getHoverWindow());

	// Map the hover row (1-based buffer line) to a DOM element for pixel positioning
	let domTop: number | null = $state(null);

	$effect(() => {
		if (!hover) {
			domTop = null;
			return;
		}
		// hover.row is 1-based buffer line number
		const el = document.getElementById(`row-${hover.row}`);
		if (el) {
			domTop = el.offsetTop;
		} else {
			domTop = null;
		}
	});

	let position = $derived.by(() => {
		if (!hover) return { top: '0px', left: '0px', display: 'none' };

		const lineHeight = 28; // 22px font + 3px padding top + 3px padding bottom
		let topPx: number;

		if (domTop !== null) {
			// Position below the line where K was pressed
			topPx = domTop + lineHeight;
		} else {
			topPx = hover.row * lineHeight + 3;
		}

		return {
			top: `${topPx}px`,
			left: `${hover.col + 6}ch`,
			display: 'block'
		};
	});

	// Render hover content as markdown
	let renderedContent = $derived.by(() => {
		if (!hover?.content) return '';
		return marked.parse(hover.content);
	});
</script>

{#if hover && hover.content}
	<div
		class="hover-window absolute z-[200] overflow-y-auto rounded-md border border-surface0 bg-base-100 text-text drop-shadow-lg"
		style="top: {position.top}; left: {position.left}; display: {position.display}; max-height: 400px; max-width: 600px; min-width: 200px;"
	>
		<div class="p-3 prose prose-sm prose-invert max-w-none">
			{@html renderedContent}
		</div>
	</div>
{/if}

<style>
	.hover-window {
		font-size: 16px;
		line-height: 1.5;
		overflow-x: hidden;
		word-wrap: break-word;
		overflow-wrap: break-word;
	}

	.hover-window :global(pre) {
		background: rgba(0, 0, 0, 0.3);
		padding: 8px;
		border-radius: 4px;
		overflow-x: auto;
	}

	.hover-window :global(code) {
		background: rgba(0, 0, 0, 0.3);
		padding: 2px 4px;
		border-radius: 2px;
		font-size: 14px;
	}

	.hover-window :global(h1),
	.hover-window :global(h2),
	.hover-window :global(h3) {
		margin-top: 12px;
		margin-bottom: 8px;
	}

	.hover-window :global(p) {
		margin-bottom: 8px;
		word-wrap: break-word;
		overflow-wrap: break-word;
		word-break: break-word;
		white-space: normal;
		max-width: 100%;
	}

	.hover-window :global(a) {
		word-wrap: break-word;
		overflow-wrap: break-word;
		word-break: break-all;
		white-space: normal;
	}

	.hover-window :global(ul),
	.hover-window :global(ol) {
		margin-left: 20px;
		margin-bottom: 8px;
	}
</style>
