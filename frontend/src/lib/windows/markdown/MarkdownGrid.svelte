<script lang="ts">
	import { tick } from 'svelte';
	import LineNumber from '../LineNumber.svelte';
	import MarkdownRow from './MarkdownRow.svelte';
	import Quote from './Quote.svelte';
	import Table from './Table.svelte';
	let { content, decode, cursor, lineNumbers, relativeLineNumbers }: App.GridProps = $props();

	$effect(() => {
		if (!cursor?.row && cursor?.row !== 0) return;
		const _ = content;
		tick().then(() => {
			const el = document.getElementById(`row-${cursor.row}`);
			el?.scrollIntoView({ block: 'nearest', behavior: 'instant' });
		});
	});

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

	function findTableGroups() {
		if (!content) return [];

		const groups: { start: number; end: number; tableId: number }[] = [];
		let currentStart: number | null = null;
		let currentTableId: number | null = null;

		for (let i = 0; i < content.length; i++) {
			const tableId = content[i].markdownOpts?.table?.tableId;
			if (tableId !== undefined && tableId !== null) {
				if (currentStart === null || currentTableId !== tableId) {
					if (currentStart !== null) {
						groups.push({ start: currentStart, end: i - 1, tableId: currentTableId! });
					}
					currentStart = i;
					currentTableId = tableId;
				}
			} else {
				if (currentStart !== null) {
					groups.push({ start: currentStart, end: i - 1, tableId: currentTableId! });
					currentStart = null;
					currentTableId = null;
				}
			}
		}
		if (currentStart !== null) {
			groups.push({ start: currentStart, end: content.length - 1, tableId: currentTableId! });
		}
		return groups;
	}

	let quoteGroups = $derived(findQuoteGroups());
	let tableGroups = $derived(findTableGroups());

	function isInQuoteGroup(index: number): boolean {
		return quoteGroups.some((g) => index >= g.start && index <= g.end);
	}

	function isFirstOfQuoteGroup(index: number): boolean {
		return quoteGroups.some((g) => g.start === index);
	}

	function getQuoteGroup(index: number): App.NvimRow[] {
		if (!content) return [];
		const group = quoteGroups.find((group) => index >= group.start && index <= group.end);
		if (!group) return [];
		return content.slice(group.start, group.end + 1);
	}

	function isInTableGroup(index: number): boolean {
		return tableGroups.some((g) => index >= g.start && index <= g.end);
	}

	function isFirstOfTableGroup(index: number): boolean {
		return tableGroups.some((g) => g.start === index);
	}

	function getTableGroup(index: number): App.NvimRow[] | null {
		if (!content) return null;
		const group = tableGroups.find((g) => index >= g.start && index <= g.end);
		if (!group) return null;
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

		{#if isInTableGroup(i)}
			{#if isFirstOfTableGroup(i)}
				<Table rows={getTableGroup(i) ?? []} {decode} />
			{/if}
		{:else if row.markdownOpts?.image}
			<img src={row.markdownOpts.image.url} alt={row.markdownOpts.image.altText}
				class="max-w-full max-h-96 object-contain" />
		{:else if isInQuoteGroup(i)}
			{#if isFirstOfQuoteGroup(i)}
				<Quote rows={getQuoteGroup(i)} {decode} />
			{/if}
		{:else}
			<MarkdownRow {row} {decode} />
		{/if}
	</div>
{/each}
