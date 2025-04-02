<script lang="ts">
	let { content, decode }: App.GridProps = $props();
</script>

{#if content}
	{#each content as row}
		<div class="flex overflow-hidden leading-none">
			{#each row as cell}
				{#if cell?.char}
					<span class="cell inline-block h-full hl-{cell?.highlight} {cell?.classes}">
						{decode ? decode(cell.char) : cell.char}
					</span>
				{/if}
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
