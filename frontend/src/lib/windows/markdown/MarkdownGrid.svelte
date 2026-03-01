<script lang="ts">
	import { tick } from 'svelte';
	import LineNumber from '../LineNumber.svelte';
	import MarkdownRow from './MarkdownRow.svelte';
	import Quote from './Quote.svelte';
	import Table from './Table.svelte';
	import CodeBlock from './CodeBlock.svelte';
	let { content, decode, cursor, lineNumbers, relativeLineNumbers, cursorLineColor }: App.GridProps = $props();

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

	function findCodeBlockGroups() {
		if (!content) return [];

		const groups: { start: number; end: number }[] = [];
		let currentStart: number | null = null;

		for (let i = 0; i < content.length; i++) {
			const cb = content[i].markdownOpts?.codeBlock;
			if (cb) {
				if (currentStart === null) {
					currentStart = i;
				}
				if (cb.position === 'last') {
					groups.push({ start: currentStart, end: i });
					currentStart = null;
				}
			} else {
				if (currentStart !== null) {
					groups.push({ start: currentStart, end: i - 1 });
					currentStart = null;
				}
			}
		}
		if (currentStart !== null) {
			groups.push({ start: currentStart, end: content.length - 1 });
		}
		return groups;
	}

	let quoteGroups = $derived(findQuoteGroups());
	let tableGroups = $derived(findTableGroups());
	let codeBlockGroups = $derived(findCodeBlockGroups());

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

	function getTableGroupEndIndex(index: number): number | undefined {
		if (!content) return undefined;
		const group = tableGroups.find((g) => g.start === index);
		if (!group) return undefined;
		return content[group.end].index;
	}

	function isInCodeBlockGroup(index: number): boolean {
		return codeBlockGroups.some((g) => index >= g.start && index <= g.end);
	}

	function isFirstOfCodeBlockGroup(index: number): boolean {
		return codeBlockGroups.some((g) => g.start === index);
	}

	function getCodeBlockGroup(index: number): App.NvimRow[] {
		if (!content) return [];
		const group = codeBlockGroups.find((g) => index >= g.start && index <= g.end);
		if (!group) return [];
		return content.slice(group.start, group.end + 1);
	}

	function getCodeBlockGroupEndIndex(index: number): number | undefined {
		if (!content) return undefined;
		const group = codeBlockGroups.find((g) => g.start === index);
		if (!group) return undefined;
		return content[group.end].index;
	}
</script>

{#each content || [] as row, i (row.index)}
	{#if (isInTableGroup(i) && !isFirstOfTableGroup(i)) || (isInCodeBlockGroup(i) && !isFirstOfCodeBlockGroup(i))}
		<!-- skip non-first grouped rows, they're rendered by Table/CodeBlock component -->
	{:else}
		<div id="row-{row.index}" class="flex overflow-hidden whitespace-pre leading-none" style={row.index === cursor?.row && cursorLineColor ? `background: ${cursorLineColor}` : ''}>
			<LineNumber
				{lineNumbers}
				{relativeLineNumbers}
				{relativeLineNumbersList}
				{cursor}
				index={row.index}
				rangeEnd={isInTableGroup(i) ? getTableGroupEndIndex(i) : isInCodeBlockGroup(i) ? getCodeBlockGroupEndIndex(i) : undefined}
			/>

			{#if isInTableGroup(i)}
				<Table rows={getTableGroup(i) ?? []} {decode} />
			{:else if isInCodeBlockGroup(i)}
				<div class="mx-6">
				<CodeBlock rows={getCodeBlockGroup(i)} {decode} />
				</div>
			{:else if row.markdownOpts?.image}
				<img src={row.markdownOpts.image.url} alt={row.markdownOpts.image.altText}
					class="max-w-full max-h-96 object-contain" />
			{:else if isInQuoteGroup(i)}
				{#if isFirstOfQuoteGroup(i)}
					<Quote rows={getQuoteGroup(i)} {decode} />
				{/if}
			{:else}
				<MarkdownRow {row} {decode} cursorOnRow={cursor?.row === row.index} />
			{/if}
		</div>
	{/if}
{/each}
