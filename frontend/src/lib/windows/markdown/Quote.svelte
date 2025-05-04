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
	$effect(() => {
		rows.forEach((row) => {
			if (!row.tokens.length) return;
			row.tokens[0].text = row.tokens[0]?.text?.replace(/^(\s*>\s*)+/, '');
		});
	});
	const level = rows[0]?.markdownOpts?.quoteLevel ?? 0;
	let currentLevelGroup: App.NvimRow[] = [];
	let nextLevelGroups: App.NvimRow[][] = [];

	for (const row of rows) {
		const rowLevel = row.markdownOpts!.quoteLevel;
		if (rowLevel === level) {
			currentLevelGroup.push(row);
		} else if (rowLevel > level) {
			if (
				nextLevelGroups.length === 0 ||
				(nextLevelGroups[nextLevelGroups.length - 1][0]?.markdownOpts?.quoteLevel ?? 0) !== rowLevel
			) {
				nextLevelGroups.push([row]);
			} else {
				nextLevelGroups[nextLevelGroups.length - 1].push(row);
			}
		} else {
			break;
		}
	}
</script>

<blockquote class="my-4 border-l-4 border-overlay2 bg-crust p-4">
	<p class="italic leading-relaxed [&_*]:!text-overlay0">
		{#each currentLevelGroup as row}
			<MarkdownRow {row} {decode} />
		{/each}

		{#each nextLevelGroups as group}
			<Quote rows={group} {decode} />
    {/each}
	</p>
</blockquote>
