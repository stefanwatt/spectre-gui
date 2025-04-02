<script lang="ts">
	import CommandIcon from '$lib/icons/Command.svelte';
	import TextWithCursor from '../TextWithCursor.svelte';

	let { prompt, indent, content, pos }: App.CmdLine = $props();
</script>

<div class="relative flex w-screen justify-center">
	<div class="cmdline-container bg-dark top-24 z-50">
		<div class="command-mode flex items-center p-2">
			<div class="mr-4 text-blue">
				<CommandIcon />
			</div>
			{#if prompt}
				<span class="cmdline-prompt mr-1 font-semibold">{prompt}</span>
			{/if}
			{#if indent && indent > 0}
				<span class="cmdline-indent">{' '.repeat(indent)}</span>
			{/if}
			<div class="cmdline-content flex">
				{#if content !== undefined && pos !== undefined}
					<TextWithCursor text={content} {pos} />
				{/if}
			</div>
		</div>
	</div>
</div>

<style>
	.command-mode {
		width: 60rem;
	}
</style>
