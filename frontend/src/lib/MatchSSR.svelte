<script lang="ts">
	import { GetReplacementText } from '$lib/wailsjs/go/main/App.js';
	import { OpenFile } from '$lib/wailsjs/go/picker/Picker.js';

	let {
		match,
		selectedMatch,
		replaceTerm,
		searchTerm,
		regex
	}: {
		match: App.RipgrepMatch;
		selectedMatch: App.RipgrepMatch | null;
		replaceTerm: string;
		searchTerm: string;
		regex: boolean;
	} = $props();

	async function openFile(match: App.RipgrepMatch) {
		OpenFile(match.AbsolutePath, match.Row, match.Col);
	}
	let button: HTMLButtonElement;

	$effect(() => {
		updateReplaceTerm(replaceTerm, button);
	});

	async function updateReplaceTerm(value: string, button: HTMLButtonElement) {
		if (!value || !button) return;
		const replacement_text = await GetReplacementText(match.MatchedLine, searchTerm, value, regex);
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
	onclick={() => {
		openFile(match);
	}}
	class:bg-surface1={selectedMatch === match}
	class="m-1 flex w-full cursor-pointer snap-start justify-start rounded-sm p-1"
>
	{@html match.Html}
</button>
