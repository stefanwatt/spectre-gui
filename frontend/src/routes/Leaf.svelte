<script>
	import { mode } from './state.svelte';
	/**@type {{node:App.NvimGuiNode}}*/
	let { node } = $props();
</script>

{#if node.text === '\n'}
	<br />
{:else if node.hl_group == 'cursor' && $mode === 'i'}
 <span class="relative">
    <span 
      class="absolute inset-0 w-0 h-full border-l-2 border-rosewater" 
    ></span>
    <span
      id={node.id}
      class="absolute left-0 nvim-gui-node nvim-gui-node__leaf whitespace-pre {node.hl_group} mode-{$mode}"
    >{node.text}</span>
  </span>
{:else}
	<span
		id={node.id}
		class="nvim-gui-node nvim-gui-node__leaf whitespace-pre {node.hl_group} mode-{$mode}"
		>{node.text}</span
	>
{/if}
