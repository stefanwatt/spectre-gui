<script lang="ts">
	import SearchIcon from '$lib/icons/Search.svelte';
	import CommandIcon from '$lib/icons/Command.svelte';
	import ArrowUpIcon from '$lib/icons/ArrowUp.svelte';
	import ArrowDownIcon from '$lib/icons/ArrowDown.svelte';
	import MagicIcon from '$lib/icons/Magic.svelte';
	import GlobalIcon from '$lib/icons/Global.svelte';
	import ReplaceIcon from '$lib/icons/Replace.svelte';
	import ConfirmIcon from '$lib/icons/Confirm.svelte';
	import CaseSensitiveIcon from '$lib/icons/CaseSensitive.svelte';
	import TextWithCursor from './TextWithCursor.svelte';

	let { firstc, prompt, indent, content, pos, visible }: App.CmdLine = $props();
	let totalResults = 17;
	let resultIndex = 1;
	let searchIconSize = '16';

	let isSubstitute = $derived(content && content.includes('s/'));
	let substituteCommand = $derived(
		visible && isSubstitute && content ? parseSubstituteCommand(content) : null
	);

	function parseSubstituteCommand(cmd: string) {
		console.log('parseSubstituteCommand');
		const match = cmd.match(/^(.*?)s\/(.*?)\/(.*?)(?:\/([gicI]*))?$/);
		if (!match) {
			return null;
		}
		const [_, range, search, replace, flags = ''] = match;

		const hasVeryMagic = search.startsWith('\\v');
		const searchTerm = hasVeryMagic ? search.substring(2) : search;
		let searchPos = 0;
		let replacePos = 0;
		let isInSearchField = false;
		let isInReplaceField = false;
		if (pos !== undefined) {
			searchPos =
				pos > range.length + 2 ? Math.min(pos - (range.length + 2), searchTerm.length) : 0;
			replacePos =
				pos > range.length + 3 + search.length
					? Math.min(pos - (range.length + 3 + search.length), replace.length)
					: 0;
			isInSearchField = pos > range.length + 2 && pos <= range.length + 2 + search.length;
			isInReplaceField =
				pos > range.length + 3 + search.length &&
				pos <= range.length + 3 + search.length + replace.length;
		}

		return {
			range,
			search: searchTerm,
			replace,
			flags,
			hasVeryMagic,
			hasGlobal: flags.includes('g'),
			hasIgnoreCase: flags.includes('i'),
			hasCaseSensitive: flags.includes('I'),
			hasConfirm: flags.includes('c'),
			searchPos,
			replacePos,
			isInSearchField,
			isInReplaceField
		};
	}
</script>

{#if isSubstitute && substituteCommand}
	<div class="relative w-screen text-sm">
		<div class="absolute right-8 top-8">
			<div
				class="w-[25rem] rounded-md border border-l-4 border-overlay1 border-l-mauve bg-surface0 bg-opacity-95 shadow-md backdrop-blur-sm"
			>
				<div
					class:bg-surface1={substituteCommand.isInSearchField}
					class="flex items-center border-b border-overlay1 px-3 py-2"
				>
					<div class="mr-2 text-blue">
						<SearchIcon size={searchIconSize} />
					</div>
					<div class="grow text-text">
						{#if substituteCommand?.search !== undefined && pos !== undefined}
							<TextWithCursor text={substituteCommand.search} pos={substituteCommand.searchPos} />
						{/if}
					</div>
					<!-- flags -->
					<div class="justify-right flex">
						{#if substituteCommand.hasVeryMagic}
							<MagicIcon size={searchIconSize} />
						{/if}
						{#if substituteCommand.hasGlobal}
							<span class="ml-1"></span>
							<GlobalIcon size={searchIconSize} />
						{/if}
						{#if substituteCommand.hasCaseSensitive}
							<span class="ml-1"></span>
							<CaseSensitiveIcon size={searchIconSize} />
						{/if}
						{#if substituteCommand.hasIgnoreCase}
							<div class="ml-1 text-overlay0">
								<CaseSensitiveIcon size={searchIconSize} />
							</div>
						{/if}
					</div>
				</div>
				<div
					class:bg-surface1={substituteCommand.isInReplaceField}
					class="flex items-center px-3 py-2"
				>
					<div class="mr-2 text-green">
						<ReplaceIcon size={searchIconSize} />
					</div>
					<div class="grow text-text">
						{#if substituteCommand?.replace !== undefined && pos !== undefined}
							<TextWithCursor text={substituteCommand.replace} pos={substituteCommand.replacePos} />
						{/if}
					</div>
					<div class="justify-right flex">
						{#if substituteCommand.hasConfirm}
							<span class="ml-1"></span>
							<ConfirmIcon size={searchIconSize} />
						{/if}
					</div>
				</div>
			</div>
		</div>
	</div>
{:else if firstc === ':'}
	<div class="relative flex w-screen justify-center">
		<div class="absolute top-24">
			<div
				class="cmdline-container command-mode flex items-center rounded-md !border-l-blue p-2 shadow-md"
			>
				<div class="mr-4 text-blue">
					<CommandIcon />
				</div>
				{#if prompt}
					<span class="cmdline-prompt mr-1 font-semibold text-blue">{prompt}</span>
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
{:else if firstc === '/' || firstc === '?'}
	<div class="relative w-screen text-sm">
		<div class="absolute right-8 top-8">
			<div
				class="cmdline-container search-mode flex items-center rounded-md !border-l-peach p-2 shadow-md"
			>
				<div class="search-icon-container mr-2 text-overlay2">
					<SearchIcon size={searchIconSize} />
				</div>
				<div class="cmdline-content flex flex-grow">
					{#if content !== undefined && pos !== undefined}
						<TextWithCursor text={content} {pos} />
					{/if}
				</div>
				<div class="search-controls ml-2 flex items-center text-sm text-overlay0">
					<span class="mr-4 text-xs">{resultIndex}/{totalResults}</span>
					<div class="flex space-x-2">
						<button class="rounded p-1 hover:bg-overlay2">
							<ArrowUpIcon size={searchIconSize} />
						</button>
						<button class="rounded p-1 hover:bg-overlay2">
							<ArrowDownIcon size={searchIconSize} />
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
{:else}
	<!-- Normal mode or other command types -->
	<div class="relative flex w-screen justify-center">
		<div class="absolute top-24">
			<div class="cmdline-container normal-mode flex items-center rounded-md p-2 shadow-md">
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

<style>
	.cmdline-container {
		background-color: rgba(31, 34, 45, 0.95);
		backdrop-filter: blur(4px);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}
	.search-mode {
		width: 25rem;
		border-left: 3px solid;
	}
	.command-mode {
		width: 60rem;
		border-left: 3px solid;
	}
	.search-controls button {
		transition: background-color 0.07s;
	}

	.cmdline-indent {
		opacity: 0.5;
		user-select: none;
	}
</style>
