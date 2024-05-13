<script>
	import { selected_match, replace_term } from './store';
	/** @type {App.RipgrepMatch}*/
	export let match;
	/** @param {App.RipgrepMatch} match*/
	function replace_match(match) {
		console.log('replace_match', match);
	}
	/** @type {HTMLButtonElement}*/
	let button;
	$: {
		update_replace_term($replace_term);
		console.log(button);
	}
	replace_term.subscribe((value) => {
		update_replace_term(value);
	});

	/** @param {string} value*/
	function update_replace_term(value) {
		if (!value) return;
		console.log('replace_term changed: ', value);
		if (!button) return;
		const replacement_elem = button.querySelector('.spectre-replacement');
		if (!replacement_elem) {
			const match_elem = button.querySelector('.spectre-matched');
			if (!match_elem) return;
			match_elem.insertAdjacentHTML(
				'afterend',
				`<span class="spectre-replacement">${value}</span>`
			);
			return;
		}
		console.log('replacing inner html of elem: ', replacement_elem);
		replacement_elem.innerHTML = value;
	}
</script>

<button
	bind:this={button}
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
			{match.ReplacementText}
		</div>
	</div>
	<div class="flex h-full items-center">
		<code class="language-go bg-transparent">
			<div>{match.TextAfterMatch}</div>
		</code>
	</div>
</button>
