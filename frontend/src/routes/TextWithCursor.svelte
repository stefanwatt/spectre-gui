<script lang="ts">
	let { text, pos }: { text: string; pos: number } = $props();
</script>

{#if !text}
	<span class="border-l border-l-text"> ⠀ </span>
{:else}
	<span class="relative">
		<span class="absolute border-l border-l-text" style="left:{pos}ch"> ⠀ </span>
		{text}
	</span>
{/if}
