<script>
	import { GetReplacementText } from '$lib/wailsjs/go/main/App';
	import { selected_match, regex, preserve_case, replace_term } from './store';
	/** @type {App.RipgrepMatch}*/
	export let match;
	/** @param {App.RipgrepMatch} match*/
	function replace_match(match) {
		console.log('replace_match', match);
	}
	let replacement_text = '';
	replace_term.subscribe(async (value) => {
		if (!$preserve_case && !$regex) {
			replacement_text = value;
			return;
		}
		replacement_text = await GetReplacementText(
			match.MatchedLine,
			match.MatchedText,
			value,
			$regex
		);
	});
</script>

<button
	on:click={() => {
		replace_match(match);
	}}
	class:bg-surface1={$selected_match === match}
	class="m-1 flex w-full cursor-pointer snap-start justify-start rounded-sm p-1"
>
	{#if match.TextBeforeMatch}
		<div class="flex h-full items-center">
			<code class="language-go">
				<div>{match.TextBeforeMatch}</div>
			</code>
		</div>
	{/if}

	<div class="flex h-full items-center whitespace-pre">
		<div
			class="spectre-matched flex whitespace-pre-wrap rounded-sm bg-flamingo font-mono text-surface1 line-through"
		>
			{match.MatchedText}
		</div>
		<div class="spectre-replacement ml-1 whitespace-pre rounded-sm bg-surface1 text-flamingo">
			{replacement_text}
		</div>
	</div>
	<div class="flex h-full items-center">
		<code class="language-go bg-transparent">
			<div>{match.TextAfterMatch}</div>
		</code>
	</div>
</button>
