<script>
	import { selected_match, search_term, replace_term } from './store';
	/** @type {App.RipgrepMatch}*/
	export let match;
	/** @param {App.RipgrepMatch} match*/
	function replace_match(match) {
		console.log('replace_match', match);
	}
</script>

<button
	on:click={() => {
		replace_match(match);
	}}
	class:bg-surface1={$selected_match === match}
	class="m-1 flex w-full cursor-pointer snap-start justify-start rounded-sm p-1"
>
	<div class="flex h-full items-center">
		<code class="language-go">
			<div>{match.TextBeforeMatch}</div>
		</code>
	</div>

	<div class="flex h-full items-center whitespace-pre">
		<div
			class="spectre-match flex whitespace-pre-wrap rounded-sm bg-flamingo px-1 font-mono text-surface1 line-through"
		>
			{match.MatchedText}
		</div>
		<div class="spectre-match ml-1 whitespace-pre rounded-sm bg-surface1 px-1 text-flamingo">
			{match.ReplacementText}
		</div>
	</div>
	<div class="flex h-full items-center">
		<code class="language-go bg-transparent">
			<div>{match.TextAfterMatch}</div>
		</code>
	</div>
</button>

<style>
	.spectre-match {
		font-family: Consolas, Monaco, 'Andale Mono', 'Ubuntu Mono', monospace;
		font-size: 1em;
		text-align: left;
		white-space: pre;
		word-spacing: normal;
		word-break: normal;
		word-wrap: normal;
		-moz-tab-size: 2;
		-o-tab-size: 2;
		tab-size: 2;
		-webkit-hyphens: none;
		-moz-hyphens: none;
		-ms-hyphens: none;
		hyphens: none;
	}
</style>
