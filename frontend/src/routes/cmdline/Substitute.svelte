<script lang="ts">
	import TextWithCursor from '../TextWithCursor.svelte';
	import SearchIcon from '$lib/icons/Search.svelte';
	import MagicIcon from '$lib/icons/Magic.svelte';
	import GlobalIcon from '$lib/icons/Global.svelte';
	import ReplaceIcon from '$lib/icons/Replace.svelte';
	import ConfirmIcon from '$lib/icons/Confirm.svelte';
	import CaseSensitiveIcon from '$lib/icons/CaseSensitive.svelte';
  import {type SubstituteCommand} from "./substitute"

	//TODO: prevent the user from moving out of the bounds of the fields
	// sendKey logic for the substitute command should not be handled in +page.svelte
  interface SubstituteCommandProps{
    substituteCommand: SubstituteCommand;
    searchIconSize: number|string;
    pos?: number;
  }
  let {substituteCommand,searchIconSize,pos}:SubstituteCommandProps = $props()
</script>

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
