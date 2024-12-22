<script>
	/**@type {{node:App.NvimGuiNode}}*/
	let { node } = $props();
</script>

{#if node.text === '\n'}
	<br />
{:else}
	<span id="{node.id}" class="nvim-gui-node nvim-gui-node__leaf whitespace-pre {node.hl_group}">{node.text}</span>
{/if}
