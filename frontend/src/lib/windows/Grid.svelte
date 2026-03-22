<script lang="ts">
	import LineNumber from "./LineNumber.svelte";

	let { content, decode, cursor, lineNumbers, relativeLineNumbers, cursorLineColor }: App.GridProps = $props();
</script>

{#each content || [] as row (row.index)}
	<div id="row-{row.index}" class="flex overflow-hidden whitespace-pre leading-none" style={row.index === cursor?.row && cursorLineColor ? `background: ${cursorLineColor}` : ''}>
		<LineNumber
			{lineNumbers}
			{relativeLineNumbers}
			cursorRow={cursor?.row}
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
