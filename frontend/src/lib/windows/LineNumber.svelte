<script lang="ts">
	let {
		lineNumbers,
		cursor,
		relativeLineNumbers,
		relativeLineNumbersList,
		index,
		rangeEnd
	}: {
		lineNumbers: boolean;
		relativeLineNumbers: boolean;
		relativeLineNumbersList: {
			[key: number]: number;
		};
		cursor?: { row: number; col: number };
		index: number;
		rangeEnd?: number;
	} = $props();
</script>

{#if lineNumbers}
	{#if rangeEnd !== undefined}
		<span
			class="w-[6ch] shrink-0 text-surface1 flex items-center justify-end pr-2 text-right text-[0.7em]"
		>
			{index}-{rangeEnd}
		</span>
	{:else if cursor && cursor.row === index}
		<span
			class:!pr-[2ch]={relativeLineNumbers}
			class="w-[6ch] shrink-0 text-text cursor-row-{cursor.row} flex items-center justify-end pr-2 text-right"
		>
			{cursor.row}
		</span>
	{:else}
		<span
			class:!text-text={cursor?.row === index}
			class:!pr-[1ch]={relativeLineNumbers && cursor?.row === index}
			class="flex w-[6ch] shrink-0 items-center justify-end pr-2 text-right text-surface1"
		>
			{#if relativeLineNumbers}
				{relativeLineNumbersList[index]}
			{:else}
				{index}
			{/if}
		</span>
	{/if}
{/if}
