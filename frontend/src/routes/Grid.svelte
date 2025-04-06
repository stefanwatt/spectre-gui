<script lang="ts">
	let { content, decode, cursor }: App.GridProps = $props();
	let relativeNumber = $state(true);

	function computeRelativeLineNumbers() {
		if (!content || !cursor || !relativeNumber || cursor.row === undefined) return [];

		const numbers = {};
		for (const row of content) {
			const relNum = row.index - cursor.row;
			//@ts-ignore
			numbers[row.index] = Math.abs(row.index === cursor.row ? row.index : relNum);
		}

		return numbers;
	}
	let relativeLineNumbers: { [key: number]: number } = $derived(computeRelativeLineNumbers());
</script>

{#each content || [] as row (row.index)}
	<div id="row-{row.index}" class="flex overflow-hidden whitespace-pre leading-none">
		{#if cursor && cursor.row === row.index}
			<span
				class:!pr-[2ch]={relativeNumber}
				class="w-[6ch] text-text cursor-row-{cursor.row} pr-2 text-right"
			>
				{cursor.row}
			</span>
		{:else}
			<span
				class:!text-text={cursor?.row === row.index}
				class:!pr-[1ch]={relativeNumber && cursor?.row === row.index}
				class="w-[6ch] pr-2 text-right text-surface1"
			>
				{#if relativeNumber}
					{relativeLineNumbers[row.index]}
				{:else}
					{row.index}
				{/if}
			</span>
		{/if}
		{#each row.tokens as token}
			{#if token?.text}
				<span class="cell inline-block h-full hl-{token?.highlight} {token?.classes}">
					{decode ? decode(token.text) : token.text}
				</span>
			{:else}
				<span class="cell inline-block h-full">⠀</span>
			{/if}
		{:else}
			<span class="cell inline-block h-full">⠀</span>
		{/each}
	</div>
{/each}

<style>
	.cell {
		padding-top: 3px;
		padding-bottom: 3px;
	}
</style>
