<script lang="ts">
	import MarkdownToken from './MarkdownToken.svelte';
	let {
		row,
		decode,
		cursorOnRow = false
	}: {
		row: App.NvimRow;
		decode?: (input: string) => string;
		cursorOnRow?: boolean;
	} = $props();

	let headingLevel = $derived(row.markdownOpts?.headingLevel || 0);
	let task = $derived(row.markdownOpts?.task);

	// Strip leading "# " (headingLevel hashes + space) from the token stream
	let displayTokens = $derived.by(() => {
		if (headingLevel === 0) return row.tokens;
		let charsToStrip = headingLevel + 1; // e.g. "## " = 3 chars
		const result: App.NvimToken[] = [];
		for (const token of row.tokens) {
			if (charsToStrip <= 0) {
				result.push(token);
			} else if (token.text.length <= charsToStrip) {
				charsToStrip -= token.text.length;
			} else {
				result.push({ ...token, text: token.text.slice(charsToStrip) });
				charsToStrip = 0;
			}
		}
		return result;
	});

	// Strip leading "- [x] " or "- [ ] " (6 chars) from the token stream
	let taskDisplayTokens = $derived.by(() => {
		if (!task) return row.tokens;
		let charsToStrip = 6;
		const result: App.NvimToken[] = [];
		for (const token of row.tokens) {
			if (charsToStrip <= 0) {
				result.push(token);
			} else if (token.text.length <= charsToStrip) {
				charsToStrip -= token.text.length;
			} else {
				result.push({ ...token, text: token.text.slice(charsToStrip) });
				charsToStrip = 0;
			}
		}
		return result;
	});
</script>

{#if task}
	<div class="task-row">
		<input type="checkbox" checked={task.checked} disabled class="checkbox" />
		{#each taskDisplayTokens as token}
			<MarkdownToken extraClasses={task.checked ? ' line-through' : ''} {token} {decode} />
		{/each}
	</div>
{:else if headingLevel > 0}
	<span class="heading heading-{headingLevel}">
		{#each cursorOnRow ? row.tokens : displayTokens as token}
			<MarkdownToken {token} {decode} />
		{/each}
	</span>
{:else}
	{#each row.tokens as token}
		<MarkdownToken {token} {decode} />
	{/each}
{/if}

<style>
	.task-row {
		display: inline-flex;
		align-items: center;
		gap: 4px;
	}
	.heading {
		display: inline-flex;
		align-items: baseline;
		font-weight: bold;
	}
	.heading-1 {
		font-size: 2em;
		line-height: 2.5em;
	}
	.heading-2 {
		font-size: 1.5em;
		line-height: 2em;
	}
	.heading-3 {
		font-size: 1.25em;
		line-height: 1.75em;
	}
	.heading-4 {
		font-size: 1.1em;
		line-height: 1.5em;
	}
	.heading-5 {
		font-size: 1em;
		line-height: 1.25em;
	}
	.heading-6 {
		font-size: 0.9em;
		line-height: 1.2em;
	}
</style>
