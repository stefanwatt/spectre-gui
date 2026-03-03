<script lang="ts">
	import { marked } from 'marked';
	import { getCompletionDocumentation } from '$lib/state.svelte';

	interface DocumentationWindowProps {
		menuRef: HTMLDivElement | undefined;
	}
	let { menuRef }: DocumentationWindowProps = $props();

	let doc = $derived(getCompletionDocumentation());
	let docRef: HTMLDivElement | undefined = $state(undefined);

	$effect(() => {
		console.log('[DocumentationWindow] doc changed:', {
			hasDoc: !!doc,
			text: doc?.text?.substring(0, 50),
			kind: doc?.kind,
			detail: doc?.detail?.substring(0, 50)
		});
	});

	// Position to the right of the completion menu
	let position = $derived.by(() => {
		if (!menuRef || !doc) return { top: '0px', left: '0px', display: 'none' };

		const menuRect = menuRef.getBoundingClientRect();
		const parentEl = menuRef.closest('.nvim-window');
		if (!parentEl) return { top: '0px', left: '0px', display: 'none' };

		const parentRect = parentEl.getBoundingClientRect();
		const top = menuRect.top - parentRect.top;

		// Position to the right with a small gap
		const left = menuRect.right - parentRect.left + 8;

		return {
			top: `${top}px`,
			left: `${left}px`,
			display: 'block'
		};
	});

	// Render documentation as markdown or plaintext
	let renderedDoc = $derived.by(() => {
		if (!doc) return '';

		let content = '';
		if (doc.detail) {
			content = `\`\`\`\n${doc.detail}\n\`\`\`\n\n`;
		}
		if (doc.text) {
			content += doc.text;
		}

		if (doc.kind === 'markdown') {
			return marked.parse(content);
		} else {
			// Plaintext - wrap in pre tag
			return `<pre>${content.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</pre>`;
		}
	});
</script>

{#if doc && (doc.text || doc.detail)}
	<div
		bind:this={docRef}
		class="documentation-window absolute z-50 overflow-y-auto rounded-md border border-surface0 bg-base-100 text-text drop-shadow-lg"
		style="top: {position.top}; left: {position.left}; display: {position.display}; max-height: 400px; max-width: 600px; min-width: 300px;"
	>
		<div class="p-3 prose prose-sm prose-invert max-w-none">
			{@html renderedDoc}
		</div>
	</div>
{/if}

<style>
	.documentation-window {
		font-size: 16px;
		line-height: 1.5;
	}

	.documentation-window :global(pre) {
		background: rgba(0, 0, 0, 0.3);
		padding: 8px;
		border-radius: 4px;
		overflow-x: auto;
	}

	.documentation-window :global(code) {
		background: rgba(0, 0, 0, 0.3);
		padding: 2px 4px;
		border-radius: 2px;
		font-size: 14px;
	}

	.documentation-window :global(h1),
	.documentation-window :global(h2),
	.documentation-window :global(h3) {
		margin-top: 12px;
		margin-bottom: 8px;
	}

	.documentation-window :global(p) {
		margin-bottom: 8px;
	}

	.documentation-window :global(ul),
	.documentation-window :global(ol) {
		margin-left: 20px;
		margin-bottom: 8px;
	}
</style>
