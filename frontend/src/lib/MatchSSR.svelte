<script>
	import { GetReplacementText } from '$lib/wailsjs/go/main/App';
	import { selected_match, regex, search_term, replace_term } from './store';
	/** @type {App.RipgrepMatch}*/
	export let match;
	/** @param {App.RipgrepMatch} match*/
	function replace_match(match) {
		console.log('replace_match', match);
	}
	/** @type {HTMLButtonElement}*/
	let button;
	$: {
		update_replace_term($replace_term, button);
	}
	replace_term.subscribe((value) => {
		update_replace_term(value, button);
	});
	/** @param {string} value
	 @param {HTMLButtonElement} button*/
	async function update_replace_term(value, button) {
		if (!value || !button) return;
		const replacement_text = await GetReplacementText(
			match.MatchedLine,
			$search_term,
			value,
			$regex
		);
		const replacement_elem = button.querySelector('.spectre-replacement');
		if (!replacement_elem) {
			const match_elem = button.querySelector('.spectre-matched');
			if (!match_elem) {
				return;
			}
			match_elem.insertAdjacentHTML(
				'afterend',
				`<span class="spectre-replacement">${replacement_text || value}</span>`
			);
			return;
		}
		replacement_elem.innerHTML = replacement_text;
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
	{@html match.Html}
</button>
