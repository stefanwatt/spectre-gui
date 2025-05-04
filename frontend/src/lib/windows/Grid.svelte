<script lang="ts">
	import LineNumber from "./LineNumber.svelte";

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
</script>

{#each content || [] as row (row.index)}
	<div id="row-{row.index}" class="flex overflow-hidden whitespace-pre leading-none">
		<LineNumber
			{lineNumbers}
			{relativeLineNumbers}
			{relativeLineNumbersList}
			{cursor}
			index={row.index}
		/>
		{#each row.tokens as token}
			{#if token?.text}
				<span class="cell inline-block h-full {token?.classes}">
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
