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
		const match = cmd.match(/^(.*?)s\/(.*?)\/(.*?)(?:\/([gicI]*))?$/);
		if (!match) {
			return null;
		}
		const [_, range, search, replace, flags = ''] = match;

		const hasVeryMagic = search.startsWith('\\v');
		const searchTerm = hasVeryMagic ? search.substring(2) : search;

		// Calculate field positions
		const searchStart = range.length + 2; // After "s/"
		const searchEnd = searchStart + search.length;
		const replaceStart = searchEnd + 1; // After the second "/"
		const replaceEnd = replaceStart + replace.length;

		// Determine which field has focus
		let isInSearchField = false;
		let isInReplaceField = false;
		let searchPos = undefined;
		let replacePos = undefined;

		if (pos !== undefined) {
			isInSearchField = pos >= searchStart && pos <= searchEnd;
			isInReplaceField = pos >= replaceStart && pos <= replaceEnd;

			if (isInSearchField) {
				searchPos = pos - searchStart;
			} else if (isInReplaceField) {
				replacePos = pos - replaceStart;
			}
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
			isInReplaceField,
			searchStart,
			replaceStart
		};
	}
	//TODO: prevent the user from moving out of the bounds of the fields
	// sendKey logic for the substitute command should not be handled in +page.svelte
</script>

{#if isSubstitute && substituteCommand}
	<div class="relative w-screen text-sm">
		<div class="cmdline-container bg-dark right-8 top-8 z-50">
			<div class="w-[25rem]">
				<div
					class="flex items-center px-3 py-2 border-b border-b-base"
				>
					<div class="mr-2 ">
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
					class="flex items-center px-3 py-2">
					<div class="mr-2 ">
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
		<div class="top-24 z-50 cmdline-container bg-dark">
			<div class="command-mode flex items-center p-2">
				<div class="mr-4 ">
					<CommandIcon />
				</div>
				{#if prompt}
					<span class="cmdline-prompt mr-1 font-semibold ">{prompt}</span>
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
		<div class="right-8 top-8 z-50 cmdline-container bg-dark">
			<div class=" search-mode flex items-center p-2">
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
		<div class="top-24 bg-dark cmdline-container bg-dark">
			<div class=" normal-mode flex items-center  p-2 shadow-md">
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
	.search-mode {
		width: 25rem;
	}
	.command-mode {
		width: 60rem;
	}
	.cmdline-container {
		border: 1px solid;
		@apply border-base;
		@apply rounded-md;
		position: absolute
	}
	.search-controls button {
		transition: background-color 0.07s;
	}

	.cmdline-indent {
		user-select: none;
	}
</style>
