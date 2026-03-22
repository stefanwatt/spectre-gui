<script lang="ts">
	let {
		lineNumbers,
		cursorRow,
		relativeLineNumbers,
		index,
		rangeEnd
	}: {
		lineNumbers: boolean;
		relativeLineNumbers: boolean;
		cursorRow?: number;
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
	{:else if cursorRow !== undefined && cursorRow === index}
		<span
			class:!pr-[2ch]={relativeLineNumbers}
			class="w-[6ch] shrink-0 text-text cursor-row-{cursorRow} flex items-center justify-end pr-2 text-right"
		>
			{cursorRow}
		</span>
	{:else}
		<span
			class:!text-text={cursorRow === index}
			class:!pr-[1ch]={relativeLineNumbers && cursorRow === index}
			class="flex w-[6ch] shrink-0 items-center justify-end pr-2 text-right text-surface1"
		>
			{#if relativeLineNumbers && cursorRow !== undefined}
				{Math.abs(index - cursorRow)}
			{:else}
				{index}
			{/if}
		</span>
	{/if}
{/if}
