<script lang="ts">
	let {
		rows,
		decode
	}: {
		rows: App.NvimRow[];
		decode?: (input: string) => string;
	} = $props();

	let innerRows = $derived(rows.slice(1, -1));
	$effect(() => {
		console.log('CodeBlock rows:', rows.length, 'innerRows:', innerRows.length);
		console.log('CodeBlock all rows:', rows.map(r => ({ index: r.index, pos: r.markdownOpts?.codeBlock?.position, text: r.tokens.map(t => t.text).join('') })));
		console.log('CodeBlock innerRows:', innerRows.map(r => ({ index: r.index, text: r.tokens.map(t => t.text).join('') })));
	});
</script>

<div class="mockup-code">
	{#each innerRows as row (row.index)}
		<pre data-prefix={row.index}><code>{#each row.tokens as token, i (i)}{@const text = decode ? decode(token.text) : token.text}{#if text}<span class="cell {token.classes}">{text}</span>{:else}<span class="cell">&#x2800;</span>{/if}{/each}</code></pre>
	{/each}
</div>

<style>
	.mockup-code {
		position: relative;
		overflow-x: auto;
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.04);
		padding: 1em 0;
		width: 100%;
	}
	.mockup-code pre {
		display: flex;
		padding: 0 1.5em;
	}
	.mockup-code pre::before {
		content: attr(data-prefix);
		opacity: 0.4;
		margin-right: 1.5em;
		user-select: none;
		min-width: 2ch;
		text-align: right;
	}
	.mockup-code code {
		display: flex;
	}
</style>
