<script>
	import { mode } from './state.svelte';
	/**@type {{node:App.NvimGuiNode}}*/
	let { node } = $props();
</script>

{#if node.text === '\n'}
	<span class="relative">
		{#if node.hl_group === 'cursor'}
			{#if $mode === 'i'}
				<span class="absolute inset-0 h-full w-0 border-l-2 border-rosewater"></span>
			{:else}
				<span
					id={node.id}
					class="nvim-gui-node nvim-gui-node__leaf whitespace-pre text-transparent {node.hl_group} mode-{$mode}"
					>
					.
				</span
				>
			{/if}
		{/if}
		<br />
	</span>
{:else if node.hl_group === 'cursor' && $mode === 'i'}
	<span class="relative">
		<span class="absolute inset-0 h-full w-0 border-l-2 border-rosewater"></span>
		<span
			id={node.id}
			class="nvim-gui-node nvim-gui-node__leaf absolute left-0 whitespace-pre {node.hl_group} mode-{$mode}"
			>{node.text}</span
		>
	</span>
{:else}
	<span
		id={node.id}
		class="nvim-gui-node nvim-gui-node__leaf whitespace-pre {node.hl_group} mode-{$mode}"
		>{node.text}</span
	>
{/if}
