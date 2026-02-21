<script lang="ts">
	import MarkdownToken from './MarkdownToken.svelte';
	let {
		row,
		decode
	}: {
		row: App.NvimRow;
		decode?: (input: string) => string;
	} = $props();
</script>

{#each row.tokens as token}
	<MarkdownToken {token} {decode} />
{/each}
