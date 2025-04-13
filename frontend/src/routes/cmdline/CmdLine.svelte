<script lang="ts">
	import { parseSubstituteCommand } from './substitute';
	import Substitute from './Substitute.svelte';
	import Command from './Command.svelte';
	import Search from './Search.svelte';

	let { firstc, prompt, indent, content, pos, visible }: App.CmdLine = $props();
	let searchIconSize = '16';
	let isSubstitute = $derived(content && content.includes('s/'));
	let substituteCommand = $derived(
		visible && isSubstitute && content ? parseSubstituteCommand(content, pos) : null
	);
</script>

{#if isSubstitute && substituteCommand}
	<Substitute {substituteCommand} {pos} {searchIconSize} />
{:else if firstc === ':'}
	<Command {prompt} {indent} {content} {pos} {visible} />
{:else if firstc === '/' || firstc === '?'}
	<Search {content} {searchIconSize} {pos} />
{:else}
	<!-- fallback -->
	<div class="relative flex w-screen justify-center">
		<div class="bg-dark cmdline-container bg-dark top-24">
			<div class=" normal-mode flex items-center p-2 shadow-md">
				{#if prompt}
					<span class="cmdline-prompt mr-1 text-green">{prompt}</span>
				{/if}
				{#if indent && indent > 0}
					<span class="cmdline-indent">{' '.repeat(indent)}</span>
				{/if}
				<div class="cmdline-content flex">
					{content}
				</div>
			</div>
		</div>
	</div>
{/if}
