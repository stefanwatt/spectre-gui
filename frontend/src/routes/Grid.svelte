<script lang="ts">
	let { content, decode }: App.GridProps = $props();
</script>

{#if content}
	{#each content as row (row.index)}
		<div id="row-{row.index}" class="flex overflow-hidden whitespace-pre leading-none">
			{#each row.tokens as token}
				{#if token?.text}
					<span class="cell inline-block h-full hl-{token?.highlight} {token?.classes}">
						{decode ? decode(token.text) : token.text}
					</span>
				{/if}
			{:else}
				<span class="cell inline-block h-full">⠀</span>
			{/each}
		</div>
	{/each}
{/if}

<style>
	.cell {
		padding-top: 3px;
		padding-bottom: 3px;
	}
</style>
