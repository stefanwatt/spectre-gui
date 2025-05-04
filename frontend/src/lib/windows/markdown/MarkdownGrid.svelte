<script lang="ts">
	import LineNumber from '../LineNumber.svelte';
	import MarkdownRow from './MarkdownRow.svelte';
	import Quote from './Quote.svelte';
	let { content, decode, cursor, lineNumbers, relativeLineNumbers }: App.GridProps = $props();

	function computeRelativeLineNumbers() {
		if (!content || !cursor || !relativeLineNumbers || cursor.row === undefined) return [];
		const numbers = {};
		for (const row of content) {
			const relNum = row.index - cursor.row;
			//@ts-ignore
			numbers[row.index] = Math.abs(row.index === cursor.row ? row.index : relNum);
		}
		return numbers;
	}

	let relativeLineNumbersList: { [key: number]: number } = $derived(computeRelativeLineNumbers());

	function findQuoteGroups() {
		if (!content) return [];

		const groups = [];
		let currentStart = null;

		for (let i = 0; i < content.length; i++) {
			const row = content[i];
			const quoteLevel = row.markdownOpts?.quoteLevel || 0;
			if (quoteLevel > 0 && currentStart === null) {
				currentStart = i;
			}
			if (!quoteLevel && currentStart !== null) {
				groups.push({ start: currentStart, end: i - 1 });
				currentStart = null;
			}
		}
		if (currentStart !== null) {
			groups.push({ start: currentStart, end: content.length - 1 });
		}
		return groups;
	}

	let quoteGroups = $derived(findQuoteGroups());

	function getQuoteGroup(index: number): App.NvimRow[] {
		if (!content) return [];
		const group = quoteGroups.find((group) => index >= group.start && index <= group.end);
		if (!group) return [];
		return content.slice(group.start, group.end + 1);
	}
</script>

{#each content || [] as row, i (row.index)}
	<div id="row-{row.index}" class="flex overflow-hidden whitespace-pre leading-none">
		<LineNumber
			{lineNumbers}
			{relativeLineNumbers}
			{relativeLineNumbersList}
			{cursor}
			index={row.index}
		/>

		{#if row.markdownOpts?.quoteLevel && content?.length}
			{#if i === 0 || !content[i - 1].markdownOpts?.quoteLevel}
				<!-- This is the first row of a quote group -->
				<Quote rows={getQuoteGroup(i)} {decode} />
			{/if}
		{:else}
			<MarkdownRow {row} {decode} />
		{/if}
	</div>
{/each}
