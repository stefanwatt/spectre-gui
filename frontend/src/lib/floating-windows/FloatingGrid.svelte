<script lang="ts">
	let { content, decode }: App.GridProps = $props();
</script>

{#if content}
	{#each content as row}
		<div id="row-{row.index}" class="flex overflow-hidden whitespace-pre leading-none">
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
{/if}

<style>
	.cell {
		padding-top: 3px;
		padding-bottom: 3px;
	}
</style>
