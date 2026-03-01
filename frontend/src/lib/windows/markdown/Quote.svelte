<script lang="ts">
	import Quote from './Quote.svelte';
	import MarkdownRow from './MarkdownRow.svelte';
	let {
		rows,
		decode
	}: {
		rows: App.NvimRow[];
		decode?: (input: string) => string;
	} = $props();

	function stripQuotePrefix(row: App.NvimRow): App.NvimRow {
		if (!row.tokens.length) return row;
		const fullText = row.tokens.map((t) => t.text).join('');
		const match = fullText.match(/^(\s*>\s*)+/);
		if (!match) return row;
		let charsToStrip = match[0].length;
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
		return { ...row, tokens: result };
	}

	let level = $derived(rows[0]?.markdownOpts?.quoteLevel ?? 0);

	let groupData = $derived.by(() => {
		const current: App.NvimRow[] = [];
		const nested: App.NvimRow[][] = [];

		for (const row of rows) {
			const rowLevel = row.markdownOpts?.quoteLevel ?? 0;
			if (rowLevel === level) {
				current.push(stripQuotePrefix(row));
			} else if (rowLevel > level) {
				if (
					nested.length === 0 ||
					(nested[nested.length - 1][0]?.markdownOpts?.quoteLevel ?? 0) !== rowLevel
				) {
					nested.push([row]);
				} else {
					nested[nested.length - 1].push(row);
				}
			} else {
				break;
			}
		}
		return { current, nested };
	});
</script>

<blockquote class="border-l-4 border-overlay2 bg-crust px-4 py-1">
	<p class="italic leading-relaxed [&_*]:!text-overlay0">
		{#each groupData.current as row}
			<MarkdownRow {row} {decode} />
		{/each}

		{#each groupData.nested as group}
			<Quote rows={group} {decode} />
		{/each}
	</p>
</blockquote>
