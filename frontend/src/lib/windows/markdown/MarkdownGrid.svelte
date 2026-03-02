<script lang="ts">
	import { tick } from 'svelte';
	import LineNumber from '../LineNumber.svelte';
	import MarkdownRow from './MarkdownRow.svelte';
	import Quote from './Quote.svelte';
	import Table from './Table.svelte';
	import CodeBlock from './CodeBlock.svelte';
	import MarkdownToken from './MarkdownToken.svelte';
	import LocalImage from './LocalImage.svelte';
	let { content, decode, cursor, lineNumbers, relativeLineNumbers, cursorLineColor }: App.GridProps = $props();

	$effect(() => {
		if (!cursor?.row && cursor?.row !== 0) return;
		const row = cursor.row;
		tick().then(() => {
			document.getElementById(`row-${row}`)?.scrollIntoView({ block: 'nearest', behavior: 'instant' });
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

	type RowMeta =
		| { skip: true }
		| { skip: false; kind: 'table'; isFirst: true; rows: App.NvimRow[]; groupEndIndex: number }
		| { skip: false; kind: 'table'; isFirst: false }
		| { skip: false; kind: 'code'; isFirst: true; rows: App.NvimRow[]; groupEndIndex: number }
		| { skip: false; kind: 'code'; isFirst: false }
		| { skip: false; kind: 'quote'; isFirst: true; rows: App.NvimRow[] }
		| { skip: false; kind: 'quote'; isFirst: false }
		| { skip: false; kind: 'image' }
		| { skip: false; kind: 'plain' };

	let rowMeta: RowMeta[] = $derived.by(() => {
		const c = content || [];
		return c.map((row, i) => {
			const tableG = tableGroups.find((g) => i >= g.start && i <= g.end);
			if (tableG) {
				const isFirst = tableG.start === i;
				if (!isFirst) return { skip: true } as RowMeta;
				return { skip: false, kind: 'table', isFirst: true, rows: c.slice(tableG.start, tableG.end + 1), groupEndIndex: c[tableG.end].index } as RowMeta;
			}
			const codeG = codeBlockGroups.find((g) => i >= g.start && i <= g.end);
			if (codeG) {
				const isFirst = codeG.start === i;
				if (!isFirst) return { skip: true } as RowMeta;
				return { skip: false, kind: 'code', isFirst: true, rows: c.slice(codeG.start, codeG.end + 1), groupEndIndex: c[codeG.end].index } as RowMeta;
			}
			const quoteG = quoteGroups.find((g) => i >= g.start && i <= g.end);
			if (quoteG) {
				const isFirst = quoteG.start === i;
				if (!isFirst) return { skip: false, kind: 'quote', isFirst: false } as RowMeta;
				return { skip: false, kind: 'quote', isFirst: true, rows: c.slice(quoteG.start, quoteG.end + 1) } as RowMeta;
			}
			if (row.markdownOpts?.image) return { skip: false, kind: 'image' } as RowMeta;
			return { skip: false, kind: 'plain' } as RowMeta;
		});
	});
</script>

{#each content || [] as row, i (row.index)}
	{@const meta = rowMeta[i]}
	{#if meta && !meta.skip}
		<div id="row-{row.index}" class="flex overflow-hidden whitespace-pre leading-none" style={row.index === cursor?.row && cursorLineColor ? `background: ${cursorLineColor}` : ''}>
			<LineNumber
				{lineNumbers}
				{relativeLineNumbers}
				{relativeLineNumbersList}
				{cursor}
				index={row.index}
				rangeEnd={meta.kind === 'table' && meta.isFirst ? meta.groupEndIndex : meta.kind === 'code' && meta.isFirst ? meta.groupEndIndex : undefined}
			/>

			{#if meta.kind === 'table' && meta.isFirst}
				<Table rows={meta.rows} {decode} cursorRow={cursor?.row} />
			{:else if meta.kind === 'code' && meta.isFirst}
				<div class="mx-6">
				<CodeBlock rows={meta.rows} {decode} cursorRow={cursor?.row} />
				</div>
			{:else if meta.kind === 'image'}
				{@const cursorOnImage = cursor?.row === row.index}
				<div class="h-96 w-full overflow-hidden flex flex-col">
					{#if cursorOnImage}
						<div class="whitespace-pre leading-none">
							{#each row.tokens as token, ti (ti)}<MarkdownToken {token} {decode} extraClasses="" />{/each}
						</div>
						<img src="https://placehold.co/600x400?text=placeholder" alt={row.markdownOpts?.image?.altText}
							class="max-w-full flex-1 min-h-0 object-contain" />
					{:else}
						<LocalImage url={row.markdownOpts?.image?.url ?? ''} altText={row.markdownOpts?.image?.altText ?? ''}
							class="max-w-full h-full object-contain" />
					{/if}
				</div>
			{:else if meta.kind === 'quote' && meta.isFirst}
				<Quote rows={meta.rows} {decode} cursorRow={cursor?.row} />
			{:else}
				<MarkdownRow {row} {decode} cursorOnRow={cursor?.row === row.index} />
			{/if}
		</div>
	{/if}
{/each}
