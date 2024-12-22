<script>
	import Leaf from './Leaf.svelte';

	/**@type {{children: App.NvimGuiNode[]}}*/
	let { children } = $props();
</script>

{#each children as node}
	{#if node?.children?.length}
		<span id="{node.id}" class="nvim-gui-node nvim-gui-node__container {node.hl_group}">
		<svelte:self {...node} />
		</span>
	{:else}
		<Leaf {node}></Leaf>
	{/if}
{/each}
